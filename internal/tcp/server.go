// Package tcp provides the TCP server and per-connection handler for GPS devices.
// SRP: Server is solely responsible for accepting connections and enforcing the
// concurrency limit via a semaphore — it delegates protocol handling to ConnectionHandler.
package tcp

import (
	"context"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"

	"github.com/waltherx/motos-socket/internal/services"
)

// Server listens for incoming TCP connections from GPS devices.
type Server struct {
	addr        string
	maxConns    int
	handler     *ConnectionHandler
	activeCount atomic.Int64
	wg          sync.WaitGroup
}

// NewServer creates a Server bound to addr with a maximum of maxConns simultaneous connections.
func NewServer(addr string, maxConns int, handler *ConnectionHandler) *Server {
	return &Server{
		addr:     addr,
		maxConns: maxConns,
		handler:  handler,
	}
}

// Start begins accepting connections in a background goroutine.
// Returns the net.Listener so the caller can close it on shutdown.
func (s *Server) Start(ctx context.Context) (net.Listener, error) {
	l, err := net.Listen("tcp", s.addr)
	if err != nil {
		return nil, err
	}

	slog.Info("TCP Listener GT06 iniciando", "addr", s.addr)

	sem := make(chan struct{}, s.maxConns)

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				select {
				case <-ctx.Done():
					slog.Info("TCP Listener detenido por shutdown")
				default:
					slog.Error("Error aceptando conexión TCP", "error", err)
					continue
				}
				return
			}

			select {
			case sem <- struct{}{}:
				s.wg.Add(1)
				s.activeCount.Add(1)
				services.TCPConnectionsTotal.Inc()
				services.TCPConnectionsActive.Inc()

				go func() {
					defer func() {
						<-sem
						s.wg.Done()
						s.activeCount.Add(-1)
						services.TCPConnectionsActive.Dec()
					}()
					s.handler.Handle(ctx, conn)
				}()

			default:
				services.TCPConnectionsRejected.Inc()
				slog.Warn("Límite de conexiones TCP alcanzado, rechazando", "max", s.maxConns)
				conn.Close()
			}
		}
	}()

	return l, nil
}

// ActiveCount returns the number of currently open GPS connections.
func (s *Server) ActiveCount() int64 {
	return s.activeCount.Load()
}

// Wait blocks until all active connections have closed.
func (s *Server) Wait() {
	s.wg.Wait()
}
