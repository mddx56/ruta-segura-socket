package debug

import (
	"github.com/waltherx/motos-socket/internal/modules/debug/controller"
	"github.com/waltherx/motos-socket/pkg/redis"
)

func ProvideDebugController(redisClient redis.RedisService) controller.DebugController {
	return controller.NewDebugController(redisClient)
}
