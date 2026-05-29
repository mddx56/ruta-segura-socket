package system

import (
	"github.com/gin-gonic/gin"
	"github.com/waltherx/motos-socket/internal/modules/system/controller"
)

func RegisterRoutes(router *gin.Engine, ctrl controller.SystemController) {
	router.GET("/health", ctrl.Health)
	router.GET("/metrics", ctrl.Metrics())
}
