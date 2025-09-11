package api

import (
	"backend/internal/store"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Store *store.Store
}

func Mount(g *gin.RouterGroup, h *Handler, authMW gin.HandlerFunc) {
	g.GET("/rooms", h.ListRooms)
	g.GET("/rooms/:id/messages", h.ListMessages)
	g.POST("/rooms/:id/messages", authMW, h.CreateMessage)
}
