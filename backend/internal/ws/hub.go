package ws

import (
	"backend/internal/models"
	"sync"
	"time"
)

type Client interface {
	WriteJSON(v any) error
	Close() error
	UserID() int64
	RoomID() int64
}

type Hub struct {
	mu     sync.RWMutex
	byRoom map[int64]map[Client]struct{} // roomID -> set
}

func NewHub() *Hub {
	return &Hub{byRoom: make(map[int64]map[Client]struct{})}
}

type WSMessage struct {
	Type string    `json:"type"`
	Data any       `json:"data,omitempty"`
	TS   time.Time `json:"ts"`
}

func (h *Hub) Join(c Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	rid := c.RoomID()
	if h.byRoom[rid] == nil {
		h.byRoom[rid] = make(map[Client]struct{})
	}
	h.byRoom[rid][c] = struct{}{}
}

func (h *Hub) Leave(c Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	rid := c.RoomID()
	if set := h.byRoom[rid]; set != nil {
		delete(set, c)
		if len(set) == 0 {
			delete(h.byRoom, rid)
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
	for cli := range h.byRoom[m.RoomID] {
		_ = cli.WriteJSON(msg)
	}
}
