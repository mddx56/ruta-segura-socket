package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	Hub       HubInterface
	TCPServer string
	WSPort    string
	LogURL    string
}

type HubInterface interface {
	GetClientCount() int
}

func NewHealthHandler(hub HubInterface, tcpServer, wsPort, logURL string) *HealthHandler {
	return &HealthHandler{
		Hub:       hub,
		TCPServer: tcpServer,
		WSPort:    wsPort,
		LogURL:    logURL,
	}
}

func (h *HealthHandler) Check(c *gin.Context) {
	clientCount := h.Hub.GetClientCount()

	c.JSON(http.StatusOK, gin.H{
		"status":         "ok",
		"message":        "Socket server está corriendo",
		"tcp_server":     h.TCPServer,
		"websocket_port": h.WSPort,
		"active_clients": clientCount,
		"log_endpoint":   h.LogURL,
	})
}
