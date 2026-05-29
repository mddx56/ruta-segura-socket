package services

import "github.com/prometheus/client_golang/prometheus"

// Métricas Prometheus para el servidor GT06
var (
	// Conexiones GPS (TCP)
	TCPConnectionsActive = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "gt06_tcp_connections_active",
		Help: "Número de conexiones GPS TCP activas en este momento",
	})

	TCPConnectionsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "gt06_tcp_connections_total",
		Help: "Total de conexiones GPS TCP recibidas desde el inicio",
	})

	TCPConnectionsRejected = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "gt06_tcp_connections_rejected_total",
		Help: "Conexiones GPS rechazadas por límite de semáforo",
	})

	// Paquetes GPS por protocolo
	PacketsReceived = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "gt06_packets_received_total",
		Help: "Paquetes GT06 recibidos por tipo de protocolo",
	}, []string{"protocol"}) // "login", "heartbeat", "location", "unknown"

	PacketsIgnored = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "gt06_packets_ignored_total",
		Help: "Paquetes GT06 ignorados por tipo de razón",
	}, []string{"reason"}) // "no_imei", "zero_coords", "invalid_header"

	// WebSocket clientes
	WSClientsActive = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "gt06_ws_clients_active",
		Help: "Clientes WebSocket conectados actualmente",
	})

	WSClientsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "gt06_ws_clients_total",
		Help: "Total de clientes WebSocket conectados desde el inicio",
	})

	WSBroadcastTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "gt06_ws_broadcast_total",
		Help: "Total de mensajes enviados por broadcast WebSocket",
	})

	// Worker Pool de posiciones
	PositionAPISuccess = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "gt06_position_api_success_total",
		Help: "Posiciones enviadas exitosamente a la API backend",
	})

	PositionAPIErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "gt06_position_api_errors_total",
		Help: "Errores al enviar posiciones a la API backend",
	})

	PositionAPIDropped = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "gt06_position_api_dropped_total",
		Help: "Posiciones descartadas por worker pool lleno o circuit breaker abierto",
	})

	WorkerPoolQueueSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "gt06_worker_pool_queue_size",
		Help: "Mensajes encolados actualmente en el worker pool",
	})

	// Circuit Breaker
	CircuitBreakerState = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gt06_circuit_breaker_state",
		Help: "Estado del circuit breaker (0=closed/ok, 1=half-open, 2=open/failing)",
	}, []string{"api"})
)

// InitMetrics registra todas las métricas en el registry de Prometheus
func InitMetrics() {
	prometheus.MustRegister(
		TCPConnectionsActive,
		TCPConnectionsTotal,
		TCPConnectionsRejected,
		PacketsReceived,
		PacketsIgnored,
		WSClientsActive,
		WSClientsTotal,
		WSBroadcastTotal,
		PositionAPISuccess,
		PositionAPIErrors,
		PositionAPIDropped,
		WorkerPoolQueueSize,
		CircuitBreakerState,
	)
}
