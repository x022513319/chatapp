package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ListRooms(c *gin.Context) {
	rooms, err := h.Store.ListRooms(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error", "code": "INTERNAL"})
		return
	}
	c.JSON(http.StatusOK, rooms)
}
