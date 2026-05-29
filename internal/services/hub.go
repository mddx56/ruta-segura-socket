package services

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Hub struct {
	Clients    map[*websocket.Conn]bool
	Broadcast  chan []byte
	Register   chan *websocket.Conn
	Unregister chan *websocket.Conn
	Mu         sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[*websocket.Conn]bool),
		Broadcast:  make(chan []byte, 256), // Previene bloqueos TCP
		Register:   make(chan *websocket.Conn),
		Unregister: make(chan *websocket.Conn),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Mu.Lock()
			h.Clients[client] = true
			h.Mu.Unlock()
		case client := <-h.Unregister:
			h.Mu.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				client.Close()
			}
			h.Mu.Unlock()
		case message := <-h.Broadcast:
			WSBroadcastTotal.Inc()
			h.Mu.Lock()
			for client := range h.Clients {
				// Evita que un cliente con mala conexión bloquee todo el broadcast
				client.SetWriteDeadline(time.Now().Add(5 * time.Second))
				err := client.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					client.Close()
					delete(h.Clients, client)
				}
			}
			h.Mu.Unlock()
		}
	}
}

func (h *Hub) GetClientCount() int {
	h.Mu.Lock()
	defer h.Mu.Unlock()
	return len(h.Clients)
}
