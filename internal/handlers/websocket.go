package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/waltherx/motos-socket/internal/services"
)

type WebSocketHandler struct {
	Hub      WebSocketHub
	Upgrader websocket.Upgrader
}

type WebSocketHub interface {
	RegisterClient(client *websocket.Conn)
	UnregisterClient(client *websocket.Conn)
}

func NewWebSocketHandler(hub WebSocketHub) *WebSocketHandler {
	return &WebSocketHandler{
		Hub: hub,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

func (wsh *WebSocketHandler) Handle(c *gin.Context) {
	ws, err := wsh.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Error("Error upgrading WebSocket", "error", err)
		return
	}

	wsh.Hub.RegisterClient(ws)
	services.WSClientsActive.Inc()
	services.WSClientsTotal.Inc()
	slog.Info("Cliente WebSocket conectado", "remote", ws.RemoteAddr().String())

	// ReadPump: Detecta desconexiones del cliente React
	// Sin esto, los clientes que cierran el navegador quedan como zombies
	go wsh.readPump(ws)
}

// readPump lee mensajes del cliente WS para detectar desconexión.
// Los clientes React normalmente no envían datos, pero necesitamos
// leer para detectar el close frame y responder a pings.
func (wsh *WebSocketHandler) readPump(ws *websocket.Conn) {
	defer func() {
		wsh.Hub.UnregisterClient(ws)
		ws.Close()
		services.WSClientsActive.Dec()
		slog.Info("Cliente WebSocket desconectado", "remote", ws.RemoteAddr().String())
	}()

	// Configuración de timeouts para detectar clientes muertos
	ws.SetReadLimit(512) // No esperamos mensajes grandes del frontend
	ws.SetReadDeadline(time.Now().Add(60 * time.Second))

	// Cuando recibimos un Pong, renovamos el deadline
	ws.SetPongHandler(func(string) error {
		ws.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			// Error normal de desconexión, no logueamos como error
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				slog.Warn("WebSocket cerrado inesperadamente", "remote", ws.RemoteAddr().String(), "error", err)
			}
			return
		}
	}
}
