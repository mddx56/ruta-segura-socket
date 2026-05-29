package controller

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/waltherx/motos-socket/pkg/redis"
)

type DebugController interface {
	RedisDevices(c *gin.Context)
}

type debugController struct {
	redisClient redis.RedisService
}

func NewDebugController(redisClient redis.RedisService) DebugController {
	return &debugController{
		redisClient: redisClient,
	}
}

// @Summary List Devices in Redis Cache
// @Description Devuelve la lista de IMEIs cacheados en Redis y el estado de la flag de vigencia.
// @Tags Debug
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /debug/redis-devices [get]
func (ctrl *debugController) RedisDevices(c *gin.Context) {
	ctx := context.Background()

	// Verificar si la bandera de caché está vigente
	_, err := ctrl.redisClient.Get(ctx, "devices:loaded")
	cacheValid := err == nil

	// Obtener todos los IMEIs del set
	imeis, err := ctrl.redisClient.Client().SMembers(ctx, "devices:known").Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo leer de Redis", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"cache_valid": cacheValid,
		"total_count": len(imeis),
		"devices":     imeis,
	})
}
