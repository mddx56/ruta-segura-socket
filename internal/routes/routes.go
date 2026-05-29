package routes

import (
	"github.com/gin-gonic/gin"
)

type Handlers struct {
	Health    HealthHandler
	WebSocket WebSocketHandler
	Monitor   MonitorHandler
}

type HealthHandler interface {
	Check(c *gin.Context)
}

type WebSocketHandler interface {
	ServeWs(c *gin.Context)
}

type MonitorHandler interface {
	Serve(c *gin.Context)
}

func SetupRoutes(router *gin.Engine, handlers Handlers) {
	router.GET("/health", handlers.Health.Check)
	router.GET("/ws", handlers.WebSocket.ServeWs)
	router.GET("/", handlers.Monitor.Serve)
}
