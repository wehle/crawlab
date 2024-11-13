package controllers

import (
	"net/http"

	"github.com/crawlab-team/crawlab/core/middlewares"
	"github.com/crawlab-team/crawlab/core/models/models"
	"github.com/gin-gonic/gin"
)

type RouterGroups struct {
	AuthGroup      *gin.RouterGroup
	AnonymousGroup *gin.RouterGroup
}

func NewRouterGroups(app *gin.Engine) (groups *RouterGroups) {
	return &RouterGroups{
		AuthGroup:      app.Group("/", middlewares.AuthorizationMiddleware()),
		AnonymousGroup: app.Group("/"),
	}
}

func RegisterController[T any](group *gin.RouterGroup, basePath string, ctr *BaseController[T]) {
	actionPaths := make(map[string]bool)
	for _, action := range ctr.actions {
		group.Handle(action.Method, basePath+action.Path, action.HandlerFunc)
		path := basePath + action.Path
		key := action.Method + " - " + path
		actionPaths[key] = true
	}
	registerBuiltinHandler(group, http.MethodGet, basePath+"", ctr.GetList, actionPaths)
	registerBuiltinHandler(group, http.MethodGet, basePath+"/:id", ctr.GetById, actionPaths)
	registerBuiltinHandler(group, http.MethodPost, basePath+"", ctr.Post, actionPaths)
	registerBuiltinHandler(group, http.MethodPut, basePath+"/:id", ctr.PutById, actionPaths)
	registerBuiltinHandler(group, http.MethodPatch, basePath+"", ctr.PatchList, actionPaths)
	registerBuiltinHandler(group, http.MethodDelete, basePath+"/:id", ctr.DeleteById, actionPaths)
	registerBuiltinHandler(group, http.MethodDelete, basePath+"", ctr.DeleteList, actionPaths)
}

func RegisterActions(group *gin.RouterGroup, basePath string, actions []Action) {
	for _, action := range actions {
		group.Handle(action.Method, basePath+action.Path, action.HandlerFunc)
	}
}

func registerBuiltinHandler(group *gin.RouterGroup, method, path string, handlerFunc gin.HandlerFunc, existingActionPaths map[string]bool) {
	key := method + " - " + path
	_, ok := existingActionPaths[key]
	if ok {
		return
	}
	group.Handle(method, path, handlerFunc)
}

func InitRoutes(app *gin.Engine) (err error) {
	// routes groups
	groups := NewRouterGroups(app)

	RegisterController(groups.AuthGroup, "/data/collections", NewController[models.DataCollection]())
	RegisterController(groups.AuthGroup, "/environments", NewController[models.Environment]())
	RegisterController(groups.AuthGroup, "/nodes", NewController[models.Node]())
	RegisterController(groups.AuthGroup, "/projects", NewController[models.Project]([]Action{
		{
			Method:      http.MethodGet,
			Path:        "",
			HandlerFunc: GetProjectList,
		},
	}...))
	RegisterController(groups.AuthGroup, "/schedules", NewController[models.Schedule]([]Action{
		{
			Method:      http.MethodPost,
			Path:        "",
			HandlerFunc: PostSchedule,
		},
		{
			Method:      http.MethodPut,
			Path:        "/:id",
			HandlerFunc: PutScheduleById,
		},
		{
			Method:      http.MethodPost,
			Path:        "/:id/enable",
			HandlerFunc: PostScheduleEnable,
		},
		{
			Method:      http.MethodPost,
			Path:        "/:id/disable",
			HandlerFunc: PostScheduleDisable,
		},
	}...))
	RegisterController(groups.AuthGroup, "/spiders", NewController[models.Spider]([]Action{
		{
			Method:      http.MethodGet,
			Path:        "/:id",
			HandlerFunc: GetSpiderById,
		},
		{
			Method:      http.MethodGet,
			Path:        "",
			HandlerFunc: GetSpiderList,
		},
		{
			Method:      http.MethodPost,
			Path:        "",
			HandlerFunc: PostSpider,
		},
		{
			Method:      http.MethodPut,
			Path:        "/:id",
			HandlerFunc: PutSpiderById,
		},
		{
			Method:      http.MethodDelete,
			Path:        "/:id",
			HandlerFunc: DeleteSpiderById,
		},
		{
			Method:      http.MethodDelete,
			Path:        "",
			HandlerFunc: DeleteSpiderList,
		},
		{
			Method:      http.MethodGet,
			Path:        "/:id/files/list",
			HandlerFunc: GetSpiderListDir,
		},
		{
			Method:      http.MethodGet,
			Path:        "/:id/files/get",
			HandlerFunc: GetSpiderFile,
		},
		{
			Method:      http.MethodGet,
			Path:        "/:id/files/info",
			HandlerFunc: GetSpiderFileInfo,
		},
		{
			Method:      http.MethodPost,
			Path:        "/:id/files/save",
			HandlerFunc: PostSpiderSaveFile,
		},
		{
			Method:      http.MethodPost,
			Path:        "/:id/files/save/batch",
			HandlerFunc: PostSpiderSaveFiles,
		},
		{
			Method:      http.MethodPost,
			Path:        "/:id/files/save/dir",
			HandlerFunc: PostSpiderSaveDir,
		},
		{
			Method:      http.MethodPost,
			Path:        "/:id/files/rename",
			HandlerFunc: PostSpiderRenameFile,
		},
		{
			Method:      http.MethodDelete,
			Path:        "/:id/files",
			HandlerFunc: DeleteSpiderFile,
		},
		{
			Method:      http.MethodPost,
			Path:        "/:id/files/copy",
			HandlerFunc: PostSpiderCopyFile,
		},
		{
			Method:      http.MethodPost,
			Path:        "/:id/files/export",
			HandlerFunc: PostSpiderExport,
		},
		{
			Method:      http.MethodPost,
			Path:        "/:id/run",
			HandlerFunc: PostSpiderRun,
		},
		{
			Method:      http.MethodGet,
			Path:        "/:id/results",
			HandlerFunc: GetSpiderResults,
		},
	}...))
	RegisterController(groups.AuthGroup, "/tasks", NewController[models.Task]([]Action{
		{
			Method:      http.MethodGet,
			Path:        "/:id",
			HandlerFunc: GetTaskById,
		},
		{
			Method:      http.MethodGet,
			Path:        "",
			HandlerFunc: GetTaskList,
		},
		{
			Method:      http.MethodDelete,
			Path:        "/:id",
			HandlerFunc: DeleteTaskById,
		},
		{
			Method:      http.MethodDelete,
			Path:        "",
			HandlerFunc: DeleteList,
		},
		{
			Method:      http.MethodPost,
			Path:        "/run",
			HandlerFunc: PostTaskRun,
		},
		{
			Method:      http.MethodPost,
			Path:        "/:id/restart",
			HandlerFunc: PostTaskRestart,
		},
		{
			Method:      http.MethodPost,
			Path:        "/:id/cancel",
			HandlerFunc: PostTaskCancel,
		},
		{
			Method:      http.MethodGet,
			Path:        "/:id/logs",
			HandlerFunc: GetTaskLogs,
		},
	}...))
	RegisterController(groups.AuthGroup, "/tokens", NewController[models.Token]([]Action{
		{
			Method:      http.MethodPost,
			Path:        "",
			HandlerFunc: PostToken,
		},
	}...))
	RegisterController(groups.AuthGroup, "/users", NewController[models.User]([]Action{
		{
			Method:      http.MethodPost,
			Path:        "/:id",
			HandlerFunc: GetUserById,
		},
		{
			Method:      http.MethodGet,
			Path:        "",
			HandlerFunc: GetUserList,
		},
		{
			Method:      http.MethodPost,
			Path:        "",
			HandlerFunc: PostUser,
		},
		{
			Method:      http.MethodPost,
			Path:        "/:id/change-password",
			HandlerFunc: PostUserChangePassword,
		},
		{
			Method:      http.MethodGet,
			Path:        "/me",
			HandlerFunc: GetUserMe,
		},
		{
			Method:      http.MethodPut,
			Path:        "/me",
			HandlerFunc: PutUserMe,
		},
		{
			Method:      http.MethodPost,
			Path:        "/me/change-password",
			HandlerFunc: PostUserMeChangePassword,
		},
	}...))

	RegisterActions(groups.AuthGroup, "/export", []Action{
		{
			Method:      http.MethodPost,
			Path:        "/:type",
			HandlerFunc: PostExport,
		},
		{
			Method:      http.MethodGet,
			Path:        "/:type/:id",
			HandlerFunc: GetExport,
		},
		{
			Method:      http.MethodGet,
			Path:        "/:type/:id/download",
			HandlerFunc: GetExportDownload,
		},
	})
	RegisterActions(groups.AuthGroup, "/filters", []Action{
		{
			Method:      http.MethodGet,
			Path:        "/:col",
			HandlerFunc: GetFilterColFieldOptions,
		},
		{
			Method:      http.MethodGet,
			Path:        "/:col/:value",
			HandlerFunc: GetFilterColFieldOptions,
		},
		{
			Method:      http.MethodGet,
			Path:        "/:col/:value/:label",
			HandlerFunc: GetFilterColFieldOptions,
		},
	})
	RegisterActions(groups.AuthGroup, "/settings", []Action{
		{
			Method:      http.MethodGet,
			Path:        "/:id",
			HandlerFunc: GetSetting,
		},
		{
			Method:      http.MethodPost,
			Path:        "/:id",
			HandlerFunc: PostSetting,
		},
		{
			Method:      http.MethodPut,
			Path:        "/:id",
			HandlerFunc: PutSetting,
		},
	})
	RegisterActions(groups.AuthGroup, "/stats", []Action{
		{
			Method:      http.MethodGet,
			Path:        "/overview",
			HandlerFunc: GetStatsOverview,
		},
		{
			Method:      http.MethodGet,
			Path:        "/daily",
			HandlerFunc: GetStatsDaily,
		},
		{
			Method:      http.MethodGet,
			Path:        "/tasks",
			HandlerFunc: GetStatsTasks,
		},
	})

	RegisterActions(groups.AnonymousGroup, "/system-info", []Action{
		{
			Path:        "",
			Method:      http.MethodGet,
			HandlerFunc: GetSystemInfo,
		},
	})
	RegisterActions(groups.AnonymousGroup, "/", []Action{
		{
			Method:      http.MethodPost,
			Path:        "/login",
			HandlerFunc: PostLogin,
		},
		{
			Method:      http.MethodPost,
			Path:        "/logout",
			HandlerFunc: PostLogout,
		},
	})
	RegisterActions(groups.AnonymousGroup, "/sync", []Action{
		{
			Method:      http.MethodGet,
			Path:        "/:id/scan",
			HandlerFunc: GetSyncScan,
		},
		{
			Method:      http.MethodGet,
			Path:        "/:id/download",
			HandlerFunc: GetSyncDownload,
		},
	})

	return nil
}
