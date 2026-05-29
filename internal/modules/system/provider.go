package system

import (
	"github.com/waltherx/motos-socket/internal/modules/system/controller"
	"github.com/waltherx/motos-socket/internal/tcp"
	"github.com/waltherx/motos-socket/internal/worker"
)

func ProvideSystemController(tcpSrv *tcp.Server, pw *worker.PositionWorker) controller.SystemController {
	return controller.NewSystemController(tcpSrv, pw)
}
