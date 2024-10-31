package scheduler

import (
	errors2 "errors"
	"fmt"
	"github.com/apex/log"
	"github.com/crawlab-team/crawlab/core/constants"
	"github.com/crawlab-team/crawlab/core/errors"
	"github.com/crawlab-team/crawlab/core/grpc/server"
	"github.com/crawlab-team/crawlab/core/interfaces"
	models2 "github.com/crawlab-team/crawlab/core/models/models/v2"
	"github.com/crawlab-team/crawlab/core/models/service"
	nodeconfig "github.com/crawlab-team/crawlab/core/node/config"
	"github.com/crawlab-team/crawlab/core/task/handler"
	"github.com/crawlab-team/crawlab/core/utils"
	"github.com/crawlab-team/crawlab/grpc"
	"github.com/crawlab-team/crawlab/trace"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongo2 "go.mongodb.org/mongo-driver/mongo"
	"time"
)

type ServiceV2 struct {
	// dependencies
	nodeCfgSvc interfaces.NodeConfigService
	svr        *server.GrpcServer
	handlerSvc *handler.Service

	// settings
	interval time.Duration
}

func (svc *ServiceV2) Start() {
	go svc.initTaskStatus()
	go svc.cleanupTasks()
	utils.DefaultWait()
}

func (svc *ServiceV2) Enqueue(t *models2.TaskV2, by primitive.ObjectID) (t2 *models2.TaskV2, err error) {
	// set task status
	t.Status = constants.TaskStatusPending
	t.SetCreated(by)
	t.SetUpdated(by)

	// add task
	taskModelSvc := service.NewModelServiceV2[models2.TaskV2]()
	id, err := taskModelSvc.InsertOne(*t)
	if err != nil {
		return nil, err
	}

	// task queue item
	tq := models2.TaskQueueItemV2{
		Priority: t.Priority,
		NodeId:   t.NodeId,
	}
	tq.SetId(id)
	tq.SetCreated(by)
	tq.SetUpdated(by)

	// task stat
	ts := models2.TaskStatV2{}
	ts.SetId(id)
	ts.SetCreated(by)
	ts.SetUpdated(by)

	// enqueue task
	_, err = service.NewModelServiceV2[models2.TaskQueueItemV2]().InsertOne(tq)
	if err != nil {
		return nil, trace.TraceError(err)
	}

	// add task stat
	_, err = service.NewModelServiceV2[models2.TaskStatV2]().InsertOne(ts)
	if err != nil {
		return nil, trace.TraceError(err)
	}

	// success
	return t, nil
}

func (svc *ServiceV2) Cancel(id, by primitive.ObjectID, force bool) (err error) {
	// task
	t, err := service.NewModelServiceV2[models2.TaskV2]().GetById(id)
	if err != nil {
		log.Errorf("task not found: %s", id.Hex())
		return err
	}

	// initial status
	initialStatus := t.Status

	// set status of pending tasks as "cancelled" and remove from task item queue
	if initialStatus == constants.TaskStatusPending {
		// remove from task item queue
		if err := service.NewModelServiceV2[models2.TaskQueueItemV2]().DeleteById(t.Id); err != nil {
			log.Errorf("failed to delete task queue item: %s", t.Id.Hex())
			return err
		}
		return nil
	}

	// whether task is running on master node
	isMasterTask, err := svc.isMasterNode(t)
	if err != nil {
		err := fmt.Errorf("failed to check if task is running on master node: %s", t.Id.Hex())
		t.Status = constants.TaskStatusAbnormal
		t.Error = err.Error()
		return svc.SaveTask(t, by)
	}

	if isMasterTask {
		// cancel task on master
		return svc.cancelOnMaster(t, by, force)
	} else {
		// send to cancel task on worker nodes
		return svc.cancelOnWorker(t, by, force)
	}
}

func (svc *ServiceV2) cancelOnMaster(t *models2.TaskV2, by primitive.ObjectID, force bool) (err error) {
	if err := svc.handlerSvc.Cancel(t.Id, force); err != nil {
		log.Errorf("failed to cancel task on master: %s", t.Id.Hex())
		return err
	}

	// set task status as "cancelled"
	t.Status = constants.TaskStatusCancelled
	return svc.SaveTask(t, by)
}

func (svc *ServiceV2) cancelOnWorker(t *models2.TaskV2, by primitive.ObjectID, force bool) (err error) {
	// get subscribe stream
	stream, ok := svc.svr.TaskSvr.GetSubscribeStream(t.Id)
	if !ok {
		err := fmt.Errorf("stream not found for task: %s", t.Id.Hex())
		log.Errorf(err.Error())
		t.Status = constants.TaskStatusAbnormal
		t.Error = err.Error()
		return svc.SaveTask(t, by)
	}

	// send cancel request
	err = stream.Send(&grpc.TaskServiceSubscribeResponse{
		Code:   grpc.TaskServiceSubscribeCode_CANCEL,
		TaskId: t.Id.Hex(),
		Force:  force,
	})
	if err != nil {
		log.Errorf("failed to send cancel request to worker: %s", t.Id.Hex())
		return err
	}

	return nil
}

func (svc *ServiceV2) SetInterval(interval time.Duration) {
	svc.interval = interval
}

func (svc *ServiceV2) SaveTask(t *models2.TaskV2, by primitive.ObjectID) (err error) {
	if t.Id.IsZero() {
		t.SetCreated(by)
		t.SetUpdated(by)
		_, err = service.NewModelServiceV2[models2.TaskV2]().InsertOne(*t)
		return err
	} else {
		t.SetUpdated(by)
		return service.NewModelServiceV2[models2.TaskV2]().ReplaceById(t.Id, *t)
	}
}

// initTaskStatus initialize task status of existing tasks
func (svc *ServiceV2) initTaskStatus() {
	// set status of running tasks as TaskStatusAbnormal
	runningTasks, err := service.NewModelServiceV2[models2.TaskV2]().GetMany(bson.M{
		"status": bson.M{
			"$in": []string{
				constants.TaskStatusPending,
				constants.TaskStatusRunning,
			},
		},
	}, nil)
	if err != nil {
		if errors2.Is(err, mongo2.ErrNoDocuments) {
			return
		}
		trace.PrintError(err)
	}
	for _, t := range runningTasks {
		go func(t *models2.TaskV2) {
			t.Status = constants.TaskStatusAbnormal
			if err := svc.SaveTask(t, primitive.NilObjectID); err != nil {
				trace.PrintError(err)
			}
		}(&t)
	}
	if err := service.NewModelServiceV2[models2.TaskQueueItemV2]().DeleteMany(nil); err != nil {
		return
	}
}

func (svc *ServiceV2) isMasterNode(t *models2.TaskV2) (ok bool, err error) {
	if t.NodeId.IsZero() {
		return false, trace.TraceError(errors.ErrorTaskNoNodeId)
	}
	n, err := service.NewModelServiceV2[models2.NodeV2]().GetById(t.NodeId)
	if err != nil {
		if errors2.Is(err, mongo2.ErrNoDocuments) {
			return false, trace.TraceError(errors.ErrorTaskNodeNotFound)
		}
		return false, trace.TraceError(err)
	}
	return n.IsMaster, nil
}

func (svc *ServiceV2) cleanupTasks() {
	for {
		// task stats over 30 days ago
		taskStats, err := service.NewModelServiceV2[models2.TaskStatV2]().GetMany(bson.M{
			"created_ts": bson.M{
				"$lt": time.Now().Add(-30 * 24 * time.Hour),
			},
		}, nil)
		if err != nil {
			time.Sleep(30 * time.Minute)
			continue
		}

		// task ids
		var ids []primitive.ObjectID
		for _, ts := range taskStats {
			ids = append(ids, ts.Id)
		}

		if len(ids) > 0 {
			// remove tasks
			if err := service.NewModelServiceV2[models2.TaskV2]().DeleteMany(bson.M{
				"_id": bson.M{"$in": ids},
			}); err != nil {
				trace.PrintError(err)
			}

			// remove task stats
			if err := service.NewModelServiceV2[models2.TaskStatV2]().DeleteMany(bson.M{
				"_id": bson.M{"$in": ids},
			}); err != nil {
				trace.PrintError(err)
			}
		}

		time.Sleep(30 * time.Minute)
	}
}

func NewTaskSchedulerServiceV2() (svc2 *ServiceV2, err error) {
	// service
	svc := &ServiceV2{
		interval: 5 * time.Second,
	}
	svc.nodeCfgSvc = nodeconfig.GetNodeConfigService()
	svc.svr, err = server.GetGrpcServerV2()
	if err != nil {
		log.Errorf("failed to get grpc server: %v", err)
		return nil, err
	}
	svc.handlerSvc, err = handler.GetTaskHandlerService()
	if err != nil {
		log.Errorf("failed to get task handler service: %v", err)
		return nil, err
	}

	return svc, nil
}

var svcV2 *ServiceV2

func GetTaskSchedulerServiceV2() (svr *ServiceV2, err error) {
	if svcV2 != nil {
		return svcV2, nil
	}
	svcV2, err = NewTaskSchedulerServiceV2()
	if err != nil {
		return nil, err
	}
	return svcV2, nil
}
