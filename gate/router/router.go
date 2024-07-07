package router

import (
	"common/config"
	"gate/api"
	"github.com/gin-gonic/gin"
)

// 路由注册
func RegisterRouter() *gin.Engine {
	if config.Conf.Log.Level == "Debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化grpc的client gate
	engine := gin.Default()

	userHandler := api.NewUserHandler()
	engine.POST("/register", userHandler.Register)

	return engine
}
