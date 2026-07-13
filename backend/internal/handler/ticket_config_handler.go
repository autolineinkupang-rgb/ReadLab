package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"readlab/backend/internal/ticket"
)

type TicketConfigHandler struct {
	Config *ticket.Config
}

func NewTicketConfigHandler(cfg *ticket.Config) *TicketConfigHandler {
	return &TicketConfigHandler{Config: cfg}
}

func (h *TicketConfigHandler) List(c *gin.Context) {
	configs := h.Config.List()
	c.JSON(http.StatusOK, gin.H{"data": configs})
}

func (h *TicketConfigHandler) Update(c *gin.Context) {
	var req struct {
		Key   string  `json:"key"`
		Value float64 `json:"value"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.Config.Update(req.Key, req.Value); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "config updated"})
}
