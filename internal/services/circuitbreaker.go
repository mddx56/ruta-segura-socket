package services

import (
	"log/slog"
	"time"

	"github.com/sony/gobreaker"
)

// PositionCircuitBreaker protege la llamada a la API de posiciones.
// Si la API falla repetidamente, el breaker se "abre" y deja de
// intentar enviar durante un tiempo, evitando acumular goroutines/errores.
var PositionCircuitBreaker *gobreaker.CircuitBreaker

func InitCircuitBreaker() {
	settings := gobreaker.Settings{
		Name: "position-api",

		// Cuántas peticiones se permiten en estado "Half-Open" (prueba de recuperación)
		MaxRequests: 3,

		// Tiempo mínimo en estado "Open" antes de pasar a "Half-Open"
		Interval: 30 * time.Second,

		// Tiempo que el breaker permanece "Open" antes de probar de nuevo
		Timeout: 60 * time.Second,

		// Condición para abrir el circuit breaker:
		// Si en una ventana de 10 peticiones, >= 60% fallan → OPEN
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 5 && failureRatio >= 0.6
		},

		// Callback cuando cambia el estado
		OnStateChange: func(name string, from, to gobreaker.State) {
			switch to {
			case gobreaker.StateOpen:
				slog.Warn("Circuit Breaker ABIERTO - API de posiciones no responde",
					"api", name, "from", from.String(), "to", to.String())
				CircuitBreakerState.WithLabelValues(name).Set(2)
			case gobreaker.StateHalfOpen:
				slog.Info("Circuit Breaker SEMI-ABIERTO - probando recuperación de API",
					"api", name, "from", from.String(), "to", to.String())
				CircuitBreakerState.WithLabelValues(name).Set(1)
			case gobreaker.StateClosed:
				slog.Info("Circuit Breaker CERRADO - API de posiciones recuperada",
					"api", name, "from", from.String(), "to", to.String())
				CircuitBreakerState.WithLabelValues(name).Set(0)
			}
		},
	}

	PositionCircuitBreaker = gobreaker.NewCircuitBreaker(settings)
	// Estado inicial: cerrado (OK)
	CircuitBreakerState.WithLabelValues("position-api").Set(0)
}
