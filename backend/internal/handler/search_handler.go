package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"wtr-lab-clone/backend/internal/service"
)

type SearchHandler struct {
	NovelSvc *service.NovelService
}

func NewSearchHandler(db *gorm.DB, novelSvc *service.NovelService) *SearchHandler {
	return &SearchHandler{NovelSvc: novelSvc}
}

func (h *SearchHandler) Search(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query required"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	novels, total, err := h.NovelSvc.Search(q, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  novels,
		"page":  page,
		"limit": limit,
		"total": total,
	})
}

func (h *SearchHandler) Autocomplete(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusOK, gin.H{"data": []service.AutocompleteResult{}})
		return
	}

	results, err := h.NovelSvc.Autocomplete(q, 5)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": results})
}
