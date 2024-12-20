package hub

import (
	"gin_template/middlewares"
	"gin_template/variable"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var Instance *Server

type Server struct {
	HttpEngine *gin.Engine
}

var logger = logrus.WithField("hub", "internal")

// Init 快速初始化
func Init() {
	httpEngine := gin.New()
	httpEngine.Use(RequestLogMiddleware(), gin.Recovery())
	// 访问来源限制
	httpEngine.Use((middlewares.OriginMiddleware(variable.Origin)))
	// 最大并发连接数为 1
	httpEngine.Use(middlewares.LimitHandler(1))
	Instance = &Server{
		HttpEngine: httpEngine,
	}
}

// Run 正式开启服务
func Run() {
	go func() {
		logger.Info("http engine starting...")
		if err := Instance.HttpEngine.Run("0.0.0.0:9955"); err != nil {
			logger.Fatal(err)
		} else {
			logger.Info("http engine running...")
		}
	}()
}

// StartService 启动服务
// 根据 Module 生命周期 此过程应在Login前调用
// 请勿重复调用
func StartService() {
	logger.Infof("initializing modules ...")
	for _, mi := range modules {
		mi.Instance.Init()
	}
	for _, mi := range modules {
		mi.Instance.PostInit()
	}
	logger.Info("all modules initialized")

	logger.Info("registering modules serve functions ...")
	for _, mi := range modules {
		mi.Instance.Serve(Instance)
	}
	logger.Info("all modules serve functions registered")

	logger.Info("starting modules tasks ...")
	for _, mi := range modules {
		go mi.Instance.Start(Instance)
	}
	logger.Info("tasks running")
}

// Stop 停止所有服务
func Stop() {
	logger.Warn("stopping ...")
	wg := sync.WaitGroup{}
	for _, mi := range modules {
		wg.Add(1)
		mi.Instance.Stop(Instance, &wg)
	}
	wg.Wait()
	logger.Info("stopped")
	modules = make(map[string]ModuleInfo)
}
