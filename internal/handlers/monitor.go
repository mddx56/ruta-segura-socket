package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type MonitorHandler struct {
	MonitorHTML []byte
}

func NewMonitorHandler(monitorHTML []byte) *MonitorHandler {
	return &MonitorHandler{
		MonitorHTML: monitorHTML,
	}
}

func (mh *MonitorHandler) Serve(c *gin.Context) {
	c.Data(http.StatusOK, "text/html", mh.MonitorHTML)
}
