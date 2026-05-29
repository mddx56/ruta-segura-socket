package debug

import (
	"github.com/gin-gonic/gin"
	"github.com/waltherx/motos-socket/internal/modules/debug/controller"
)

func RegisterRoutes(router *gin.Engine, ctrl controller.DebugController) {
	debugGroup := router.Group("/debug")
	{
		debugGroup.GET("/redis-devices", ctrl.RedisDevices)
	}
}
