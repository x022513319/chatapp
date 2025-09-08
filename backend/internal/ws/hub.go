package ws

import (
	"backend/internal/models"
	"sync"
	"time"
)

type Client interface {
	WriteJSON(v any) error
	Close() error
}

type Hub struct {
	mu      sync.RWMutex
	clients map[int64]map[Client]struct{} // roomID -> set
}

func NewHub() *Hub {
	return &Hub{clients: make(map[int64]map[Client]struct{})}
}

type WSMessage struct {
	Type string    `json:"type"`
	Data any       `json:"data,omitempty"`
	TS   time.Time `json:"ts"`
}

func (h *Hub) Join(roomID int64, c Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[roomID] == nil {
		h.clients[roomID] = make(map[Client]struct{})
	}
	h.clients[roomID][c] = struct{}{}
}

func (h *Hub) Leave(roomID int64, c Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if set := h.clients[roomID]; set != nil {
		delete(set, c)
		if len(set) == 0 {
			delete(h.clients, roomID)
		}
	}
}

func (h *Hub) BroadcastMessage(m models.Message) {
	payload := map[string]any{
		"id":         m.ID,
		"room_id":    m.RoomID,
		"user_id":    m.UserID,
		"content":    m.Content,
		"created_at": m.CreatedAt.UTC().Format(time.RFC3339),
	}
	msg := WSMessage{Type: "message.create", Data: payload, TS: time.Now().UTC()}

	h.mu.RLock()
	defer h.mu.RUnlock()
	for cli := range h.clients[m.RoomID] {
		_ = cli.WriteJSON(msg)
	}
}
