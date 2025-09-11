package api

import (
	"backend/internal/models"
	"log"
	"net/http"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ListMessages(c *gin.Context) {
	roomID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || roomID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room id", "code": "VALIDATION_FAILED"})
		return
	}

	limit := 50
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}

	var (
		beforeTS *time.Time
		beforeID *int64
	)

	if ts := c.Query("before_ts"); ts != "" {
		if t, err := time.Parse(time.RFC3339, ts); err == nil {
			beforeTS = &t
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid before_ts", "code": "VALIDATION_FAILED"})
			return
		}
	}
	if idStr := c.Query("before_id"); idStr != "" {
		if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
			beforeID = &id
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid before_ts", "code": "VALIDATION_FAILED"})
			return
		}
	}
	if (beforeTS == nil) != (beforeID == nil) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "before_ts & before_id must both present", "code": "VALIDATION_FAILED"})
		return
	}

	items, err := h.Store.ListMessages(c.Request.Context(), roomID, limit, beforeTS, beforeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error", "code": "INTERNAL"})
		return
	}

	resp := models.MessageResp{Items: items, PageInfo: models.PageInfo{HasMore: false}}
	if len(items) == limit {
		last := items[len(items)-1]
		resp.PageInfo.NextBeforeTS = &last.CreatedAt
		resp.PageInfo.NextBeforeID = &last.ID
		if hasMore, err := h.Store.HasOlderMessages(c.Request.Context(), roomID, last.CreatedAt, last.ID); err == nil {
			resp.PageInfo.HasMore = hasMore
		}
	}
	c.JSON(http.StatusOK, resp)
}

type createMsgReq struct {
	Content string `json:"content"`
}

func (h *Handler) CreateMessage(c *gin.Context) {
	roomID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || roomID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room id", "code": "VALIDATION_FAILED"})
		return
	}
	var req createMsgReq
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json", "code": "VALIDATION_FAILED"})
		return
	}
	req.Content = trimEdges(req.Content)
	if l := utf8.RuneCountInString(req.Content); l < 1 || l > 2000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content length 1..2000", "code": "VALIDATION_FAILED"})
		return
	}

	uidAny, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no user", "code": "UNAUTHORIZED"})
		return
	}
	userID := uidAny.(int64)

	m, err := h.Store.CreateMessage(c.Request.Context(), roomID, userID, req.Content)
	if err != nil {
		log.Printf("[CreateMessage] insert error room_id=%d user_id=%d: %v", roomID, userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "insert error", "code": "INTERNAL"})
		return
	}

	// WS hub broadcast
	// if hBroadcaster != nil { hBroadcaster.BroadcastMessage(m) }

	c.JSON(http.StatusCreated, m)
}

func trimEdges(s string) string {
	for len(s) > 0 {
		r, size := utf8.DecodeRuneInString(s)
		if r == ' ' || r == '\n' || r == '\r' || r == '\t' {
			s = s[size:]
		} else {
			break
		}
	}
	for len(s) > 0 {
		r, size := utf8.DecodeLastRuneInString(s)
		if r == ' ' || r == '\n' || r == '\r' || r == '\t' {
			s = s[:len(s)-size]
		} else {
			break
		}
	}
	return s
}
