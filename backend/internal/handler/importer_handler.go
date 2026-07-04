package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"wtr-lab-clone/backend/internal/importer"
)

type ImporterHandler struct {
	DB       *gorm.DB
	Importer *importer.Importer
}

func NewImporterHandler(db *gorm.DB) *ImporterHandler {
	return &ImporterHandler{
		DB:       db,
		Importer: importer.New(db),
	}
}

type ImportRequest struct {
	SourceID     string `json:"source_id" binding:"required"`
	WithChapters bool   `json:"with_chapters"`
}

func (h *ImporterHandler) Import(c *gin.Context) {
	var req ImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	novel, err := h.Importer.ImportNovel(req.SourceID, req.WithChapters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": novel})
}

func (h *ImporterHandler) Search(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query required"})
		return
	}

	results, err := h.Importer.Search(q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": results})
}
