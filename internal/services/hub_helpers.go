package services

import (
	"github.com/gorilla/websocket"
)

func (h *Hub) RegisterClient(client *websocket.Conn) {
	h.Register <- client
}

func (h *Hub) UnregisterClient(client *websocket.Conn) {
	h.Unregister <- client
}

func (h *Hub) BroadcastMessage(message []byte) {
	h.Broadcast <- message
}
