// Package worker manages a fixed-size pool of goroutines that persist
// GPS positions to the monitoring-api via gRPC.
// SRP: sole responsibility is enqueueing and executing SavePosition calls.
package worker

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/sony/gobreaker"
	"github.com/waltherx/motos-socket/internal/gps/gt06"
	"github.com/waltherx/motos-socket/internal/services"
	grpcclient "github.com/waltherx/motos-socket/pkg/grpc_client"
)

// PositionTask is the unit of work enqueued by the TCP connection handler.
type PositionTask struct {
	Client *grpcclient.GRPCClient
	IMEI   string
	Packet gt06.LocationPacket
}

// PositionWorker is a bounded worker pool for persisting GPS positions via gRPC.
type PositionWorker struct {
	jobs chan PositionTask
}

// NewPositionWorker creates a PositionWorker with the given channel buffer size
// and starts n goroutines to consume tasks.
func NewPositionWorker(bufferSize, workers int) *PositionWorker {
	pw := &PositionWorker{
		jobs: make(chan PositionTask, bufferSize),
	}
	for i := 0; i < workers; i++ {
		go pw.work()
	}
	return pw
}

// Enqueue submits a task to the pool. Returns false if the buffer is full
// (caller should drop the packet and increment the dropped metric).
func (pw *PositionWorker) Enqueue(task PositionTask) (ok bool) {
	// Prevents panics if Enqueue is called after pw.Stop() has closed the channel
	defer func() {
		if r := recover(); r != nil {
			ok = false
		}
	}()

	select {
	case pw.jobs <- task:
		return true
	default:
		return false
	}
}

// QueueLen returns the current number of pending tasks (used by health endpoint).
func (pw *PositionWorker) QueueLen() int {
	return len(pw.jobs)
}

// Stop drains and closes the job channel, signalling all workers to exit.
func (pw *PositionWorker) Stop() {
	close(pw.jobs)
}

func (pw *PositionWorker) work() {
	for task := range pw.jobs {
		services.WorkerPoolQueueSize.Set(float64(len(pw.jobs)))
		if err := pw.save(task); err != nil {
			services.PositionAPIErrors.Inc()
		} else {
			services.PositionAPISuccess.Inc()
		}
	}
}

func (pw *PositionWorker) save(task PositionTask) error {
	attrsJSON := buildAttrsJSON(task.Packet)

	_, err := services.PositionCircuitBreaker.Execute(func() (interface{}, error) {
		resp, err := task.Client.SavePosition(
			task.IMEI, task.Packet.DeviceUnix,
			task.Packet.Latitude, task.Packet.Longitude,
			task.Packet.Speed, task.Packet.Course,
			attrsJSON,
		)
		if err != nil {
			return nil, err
		}
		if !resp.GetSuccess() {
			return nil, fmt.Errorf("gRPC SavePosition failed: %s", resp.GetMessage())
		}
		slog.Info("Posición guardada via gRPC",
			"imei", task.IMEI,
			"lat", task.Packet.Latitude,
			"lng", task.Packet.Longitude,
			"speed", task.Packet.Speed,
		)
		return nil, nil
	})

	if err != nil {
		if err == gobreaker.ErrOpenState || err == gobreaker.ErrTooManyRequests {
			services.PositionAPIDropped.Inc()
			slog.Debug("Circuit breaker abierto, descartando posición", "imei", task.IMEI)
		} else {
			slog.Warn("Error gRPC SavePosition", "imei", task.IMEI, "error", err)
		}
	}
	return err
}

func buildAttrsJSON(p gt06.LocationPacket) string {
	b, err := json.Marshal(map[string]interface{}{
		"satellites": p.Satellites,
		"battery":    p.Battery,
		"ignition":   p.Ignition,
	})
	if err != nil {
		return "{}"
	}
	return string(b)
}
