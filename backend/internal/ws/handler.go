package ws

import (
	"backend/internal/store"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Handler struct {
	Store *store.Store
	Hub   *Hub
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // MVP
}

type wsConn struct{ *websocket.Conn }

func (c wsConn) WriteJSON(v any) error { return c.Conn.WriteJSON(v) }

func (h *Handler) Handle(c *gin.Context) {
	uidAny, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no user", "code": "UNAUTHORIZED"})
		return
	}
	userID := uidAny.(int64)

	roomIDStr := c.Query("room_id")
	roomID, err := strconv.ParseInt(roomIDStr, 10, 64)
	if err != nil || roomID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room_id", "code": "VALIDATION_FAILED"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	ws := wsConn{Conn: conn}
	h.Hub.Join(roomID, ws)
	defer func() {
		h.Hub.Leave(roomID, ws)
		_ = conn.Close()
	}()

	_ = ws.WriteJSON(WSMessage{Type: "system.ready", TS: time.Now().UTC()})

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			break
		}
		var packet struct {
			Type string          `json:"type"`
			Data json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(data, &packet); err != nil {
			continue
		}
		switch packet.Type {
		case "message.create":
			var p struct {
				Content string `json:"content"`
			}
			if err := json.Unmarshal(packet.Data, &p); err != nil {
				continue
			}
			if l := len([]rune(p.Content)); l < 1 || l > 2000 {
				_ = ws.WriteJSON(gin.H{"type": "system.ack", "error": "content length 1..2000"})
				continue
			}
			msg, err := h.Store.CreateMessage(c.Request.Context(), roomID, userID, p.Content)
			if err != nil {
				_ = ws.WriteJSON(gin.H{"type": "system.ack", "error": "insert error"})
				continue
			}
			h.Hub.BroadcastMessage(msg)
		default:
		}
	}
}
