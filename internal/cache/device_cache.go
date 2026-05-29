// Package cache provides a thread-safe in-memory cache for GPS devices.
// SRP: sole responsibility is storing and checking device IMEI membership.
// DIP: consumers depend on the DeviceCache interface, not the concrete type.
package cache

import (
	"context"
	"log/slog"
	"sync"
	"time"

	grpcclient "github.com/waltherx/motos-socket/pkg/grpc_client"
	"github.com/waltherx/motos-socket/pkg/redis"
)

// DeviceCache defines the contract for validating device IMEIs.
type DeviceCache interface {
	IsKnown(imei string) bool
	StartAutoRefresh(interval time.Duration)
}

type redisDeviceCache struct {
	redisClient redis.RedisService
	grpcClient  *grpcclient.GRPCClient
	cacheTTL    time.Duration
	mu          sync.Mutex // Previene consultas gRPC concurrentes
}

// NewDeviceCache creates a new DeviceCache using Redis with Cache-Aside pattern.
func NewDeviceCache(redisClient redis.RedisService, grpcClient *grpcclient.GRPCClient, ttl time.Duration) DeviceCache {
	return &redisDeviceCache{
		redisClient: redisClient,
		grpcClient:  grpcClient,
		cacheTTL:    ttl,
	}
}

// IsKnown checks if the IMEI exists in the Redis cache.
// If the cache is expired, it fetches the list via gRPC (Cache-Aside).
func (c *redisDeviceCache) IsKnown(imei string) bool {
	ctx := context.Background()

	// 1. Verificar si la caché está vigente
	_, err := c.redisClient.Get(ctx, "devices:loaded")
	if err != nil {
		// La caché expiró o no existe, traemos de gRPC (Plan B)
		c.loadFromGRPC(ctx)
	}

	// 2. Verificar si el IMEI está en el Set de Redis
	isMember, err := c.redisClient.Client().SIsMember(ctx, "devices:known", imei).Result()
	if err != nil {
		slog.Error("Error verificando IMEI en Redis", "error", err)
		return false
	}

	return isMember
}

// loadFromGRPC fetches devices from gRPC and updates Redis
func (c *redisDeviceCache) loadFromGRPC(ctx context.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Doble validación para evitar que múltiples goroutines llamen a gRPC al mismo tiempo
	_, err := c.redisClient.Get(ctx, "devices:loaded")
	if err == nil {
		return // Ya fue cargado por otra goroutine
	}

	resp, err := c.grpcClient.ListDevices()
	if err != nil {
		slog.Warn("Error actualizando caché de dispositivos via gRPC", "error", err)
		return
	}

	devices := resp.GetDevices()
	if len(devices) == 0 {
		return
	}

	imeis := make([]interface{}, 0, len(devices))
	for _, dev := range devices {
		imeis = append(imeis, dev.GetImei())
	}

	// Usamos un pipeline para asegurar atomicidad y mayor velocidad
	pipe := c.redisClient.Client().Pipeline()
	pipe.Del(ctx, "devices:known")
	pipe.SAdd(ctx, "devices:known", imeis...)
	pipe.Set(ctx, "devices:loaded", "1", c.cacheTTL)

	_, err = pipe.Exec(ctx)
	if err != nil {
		slog.Error("Error guardando dispositivos en Redis", "error", err)
		return
	}

	slog.Info("Caché de dispositivos actualizada en Redis via gRPC", "count", len(imeis))
}

// StartAutoRefresh can be used to pre-warm the cache periodically in the background
func (c *redisDeviceCache) StartAutoRefresh(interval time.Duration) {
	ctx := context.Background()
	c.loadFromGRPC(ctx)
	
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			// Forzamos la recarga eliminando la flag antes de cargar
			c.redisClient.Delete(ctx, "devices:loaded")
			c.loadFromGRPC(ctx)
		}
	}()
}
