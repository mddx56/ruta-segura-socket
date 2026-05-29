package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/waltherx/motos-socket/internal/services"
	"github.com/waltherx/motos-socket/internal/tcp"
	"github.com/waltherx/motos-socket/internal/worker"
)

type SystemController interface {
	Health(c *gin.Context)
	Metrics() gin.HandlerFunc
}

type systemController struct {
	tcpSrv *tcp.Server
	pw     *worker.PositionWorker
}

func NewSystemController(tcpSrv *tcp.Server, pw *worker.PositionWorker) SystemController {
	return &systemController{
		tcpSrv: tcpSrv,
		pw:     pw,
	}
}

// @Summary Check API Health
// @Description Devuelve el estado del servidor TCP y del worker pool.
// @Tags System
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func (ctrl *systemController) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":             "ok",
		"tcp_connections":    ctrl.tcpSrv.ActiveCount(),
		"worker_pool_buffer": ctrl.pw.QueueLen(),
		"circuit_breaker":    services.PositionCircuitBreaker.State().String(),
	})
}

// @Summary Prometheus Metrics
// @Description Expone métricas en formato Prometheus para Grafana scraping.
// @Tags System
// @Produce text/plain
// @Success 200 {string} string "metrics"
// @Router /metrics [get]
func (ctrl *systemController) Metrics() gin.HandlerFunc {
	return gin.WrapH(promhttp.Handler())
}

