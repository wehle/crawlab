package stats

import (
	log2 "github.com/apex/log"
	"github.com/crawlab-team/crawlab/core/constants"
	"github.com/crawlab-team/crawlab/core/database"
	interfaces2 "github.com/crawlab-team/crawlab/core/database/interfaces"
	"github.com/crawlab-team/crawlab/core/interfaces"
	models2 "github.com/crawlab-team/crawlab/core/models/models/v2"
	"github.com/crawlab-team/crawlab/core/models/service"
	nodeconfig "github.com/crawlab-team/crawlab/core/node/config"
	"github.com/crawlab-team/crawlab/core/task/log"
	"github.com/crawlab-team/crawlab/core/utils"
	"github.com/crawlab-team/crawlab/db/mongo"
	"github.com/crawlab-team/crawlab/trace"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"sync"
	"time"
)

type databaseServiceItem struct {
	taskId    primitive.ObjectID
	spiderId  primitive.ObjectID
	dbId      primitive.ObjectID
	dbSvc     interfaces2.DatabaseService
	tableName string
	time      time.Time
}

type Service struct {
	// dependencies
	nodeCfgSvc interfaces.NodeConfigService

	// internals
	mu                   sync.Mutex
	databaseServiceItems map[string]*databaseServiceItem
	databaseServiceTll   time.Duration
	logDriver            log.Driver
}

func (svc *Service) Init() (err error) {
	go svc.cleanup()
	return nil
}

func (svc *Service) InsertData(taskId primitive.ObjectID, records ...map[string]interface{}) (err error) {
	count := 0

	item, err := svc.getDatabaseServiceItem(taskId)
	if err != nil {
		return err
	}
	dbId := item.dbId
	dbSvc := item.dbSvc
	tableName := item.tableName
	if utils.IsPro() && dbSvc != nil {
		for _, record := range records {
			if err := dbSvc.CreateRow(dbId, "", tableName, svc.normalizeRecord(item, record)); err != nil {
				log2.Errorf("failed to insert data: %v", err)
				continue
			}
			count++
		}
	} else {
		var records2 []interface{}
		for _, record := range records {
			records2 = append(records2, svc.normalizeRecord(item, record))
		}
		_, err = mongo.GetMongoCol(tableName).InsertMany(records2)
		if err != nil {
			log2.Errorf("failed to insert data: %v", err)
			return err
		}
		count = len(records)
	}

	go svc.updateTaskStats(taskId, count)

	return nil
}

func (svc *Service) InsertLogs(id primitive.ObjectID, logs ...string) (err error) {
	return svc.logDriver.WriteLines(id.Hex(), logs)
}

func (svc *Service) getDatabaseServiceItem(taskId primitive.ObjectID) (item *databaseServiceItem, err error) {
	// atomic operation
	svc.mu.Lock()
	defer svc.mu.Unlock()

	// attempt to get from cache
	item, ok := svc.databaseServiceItems[taskId.Hex()]
	if ok {
		// hit in cache
		item.time = time.Now()
		return item, nil
	}

	// task
	t, err := service.NewModelService[models2.TaskV2]().GetById(taskId)
	if err != nil {
		return nil, err
	}

	// spider
	s, err := service.NewModelService[models2.SpiderV2]().GetById(t.SpiderId)
	if err != nil {
		return nil, err
	}

	// database service
	var dbSvc interfaces2.DatabaseService
	if utils.IsPro() {
		if dbRegSvc := database.GetDatabaseRegistryService(); dbRegSvc != nil {
			dbSvc, err = dbRegSvc.GetDatabaseService(s.DataSourceId)
			if err != nil {
				return nil, err
			}
		}
	}

	// item
	item = &databaseServiceItem{
		taskId:    taskId,
		spiderId:  s.Id,
		dbId:      s.DataSourceId,
		dbSvc:     dbSvc,
		tableName: s.ColName,
		time:      time.Now(),
	}

	// store in cache
	svc.databaseServiceItems[taskId.Hex()] = item

	return item, nil
}

func (svc *Service) updateTaskStats(id primitive.ObjectID, resultCount int) {
	err := service.NewModelService[models2.TaskStatV2]().UpdateById(id, bson.M{
		"$inc": bson.M{
			"result_count": resultCount,
		},
	})
	if err != nil {
		trace.PrintError(err)
	}
}

func (svc *Service) cleanup() {
	for {
		// atomic operation
		svc.mu.Lock()

		for k, v := range svc.databaseServiceItems {
			if time.Now().After(v.time.Add(svc.databaseServiceTll)) {
				delete(svc.databaseServiceItems, k)
			}
		}

		svc.mu.Unlock()

		time.Sleep(10 * time.Minute)
	}
}

func (svc *Service) normalizeRecord(item *databaseServiceItem, record map[string]interface{}) (res map[string]interface{}) {
	res = record

	// set task id
	res[constants.TaskKey] = item.taskId

	// set spider id
	res[constants.SpiderKey] = item.spiderId

	return res
}

func NewTaskStatsServiceV2() (svc2 *Service, err error) {
	// service
	svc := &Service{
		mu:                   sync.Mutex{},
		databaseServiceItems: map[string]*databaseServiceItem{},
		databaseServiceTll:   10 * time.Minute,
	}

	svc.nodeCfgSvc = nodeconfig.GetNodeConfigService()

	// log driver
	svc.logDriver, err = log.GetLogDriver(log.DriverTypeFile)
	if err != nil {
		return nil, err
	}

	return svc, nil
}

var _serviceV2 *Service

func GetTaskStatsServiceV2() (svr *Service, err error) {
	if _serviceV2 != nil {
		return _serviceV2, nil
	}
	_serviceV2, err = NewTaskStatsServiceV2()
	if err != nil {
		return nil, err
	}
	return _serviceV2, nil
}