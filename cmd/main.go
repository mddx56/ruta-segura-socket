// @title RutaSegura Server Socket
// @version 2.0
// @description Servidor TCP para GPS GT06. Persiste posiciones via gRPC al monitoring-api.
// @host localhost:6060
// @BasePath /
package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.com/waltherx/motos-socket/docs"
	"github.com/waltherx/motos-socket/internal/cache"
	"github.com/waltherx/motos-socket/internal/config"
	"github.com/waltherx/motos-socket/internal/modules/debug"
	"github.com/waltherx/motos-socket/internal/modules/system"
	"github.com/waltherx/motos-socket/internal/services"
	"github.com/waltherx/motos-socket/internal/tcp"
	"github.com/waltherx/motos-socket/internal/worker"
	grpcclient "github.com/waltherx/motos-socket/pkg/grpc_client"
	"github.com/waltherx/motos-socket/pkg/redis"
)

const (
	maxTCPConnections = 2000
	workerCount       = 20
	workerBufferSize  = 1000
	cacheRefreshEvery = 5 * time.Minute
)

func main() {
	// ── Logging ───────────────────────────────────────────────────────────────
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	// ── Bootstrap ────────────────────────────────────────────────────────────
	services.InitMetrics()
	services.InitCircuitBreaker()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Error cargando configuración:", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ── gRPC Client ──────────────────────────────────────────────────────────
	grpcClient, err := grpcclient.New(cfg.GRPCAddr)
	if err != nil {
		log.Fatal("Error conectando al servidor gRPC:", err)
	}
	defer grpcClient.Close()
	slog.Info("Cliente gRPC conectado", "addr", cfg.GRPCAddr)

	// ── Redis Client ──────────────────────────────────────────────────────────
	redisClient, err := redis.NewRedisService(cfg.RedisHost, cfg.RedisPort, cfg.RedisPassword)
	if err != nil {
		log.Fatal("Error conectando a Redis:", err)
	}

	// ── Device Cache (DIP: DeviceCache interface) ─────────────────────────────
	deviceCache := cache.NewDeviceCache(redisClient, grpcClient, cacheRefreshEvery)
	deviceCache.StartAutoRefresh(cacheRefreshEvery)

	// ── Position Worker Pool (SRP) ────────────────────────────────────────────
	posWorker := worker.NewPositionWorker(workerBufferSize, workerCount)
	defer posWorker.Stop()

	// ── TCP Server (SRP: wires ConnectionHandler + Server) ───────────────────
	connHandler := tcp.NewConnectionHandler(deviceCache, posWorker, grpcClient)
	tcpServer := tcp.NewServer(cfg.ConnHost+":"+cfg.ConnPort, maxTCPConnections, connHandler)

	tcpListener, err := tcpServer.Start(ctx)
	if err != nil {
		log.Fatal("Error iniciando servidor TCP:", err)
	}

	// ── HTTP Server (Health + Metrics + Swagger) ──────────────────────────────
	httpServer := startHTTPServer(cfg, tcpServer, posWorker, redisClient)

	// ── Graceful Shutdown ─────────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	slog.Warn("Señal de shutdown recibida", "signal", sig.String())

	// 1. Dejar de aceptar nuevas conexiones TCP
	tcpListener.Close()

	// 2. Apagar HTTP con gracia
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if httpServer != nil {
		httpServer.Shutdown(shutdownCtx)
	}

	// 3. Cancelar contexto (cierra las goroutines de conexiones activas)
	cancel()

	// 4. Esperar a que todas las conexiones TCP terminen (máximo 10s)
	done := make(chan struct{})
	go func() {
		tcpServer.Wait()
		close(done)
	}()

	select {
	case <-done:
		slog.Info("Todas las conexiones TCP cerradas correctamente")
	case <-time.After(10 * time.Second):
		slog.Warn("Timeout esperando conexiones TCP, forzando cierre")
	}

	slog.Info("Servidor apagado correctamente ✅")
}

// startHTTPServer exposes health, metrics and swagger endpoints.
func startHTTPServer(cfg *config.Config, tcpSrv *tcp.Server, pw *worker.PositionWorker, redisClient redis.RedisService) *http.Server {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// ── Setup Modules ────────────────────────────────────────────────────────
	systemCtrl := system.ProvideSystemController(tcpSrv, pw)
	system.RegisterRoutes(router, systemCtrl)

	debugCtrl := debug.ProvideDebugController(redisClient)
	debug.RegisterRoutes(router, debugCtrl)

	srv := &http.Server{
		Addr:    ":" + cfg.WSPort,
		Handler: router,
	}
	go func() {
		slog.Info("HTTP API iniciado", "port", cfg.WSPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Error HTTP:", err)
		}
	}()
	return srv
}
