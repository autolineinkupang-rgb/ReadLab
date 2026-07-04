package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"wtr-lab-clone/backend/internal/model"
)

type NovelHandler struct {
	DB *gorm.DB
}

func NewNovelHandler(db *gorm.DB) *NovelHandler {
	return &NovelHandler{DB: db}
}

func (h *NovelHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")
	genre := c.Query("genre")
	sort := c.DefaultQuery("sort", "created_at")
	order := c.DefaultQuery("order", "desc")
	q := c.Query("q")
	minChapters, _ := strconv.Atoi(c.DefaultQuery("min_chapters", "0"))
	minRating, _ := strconv.ParseFloat(c.DefaultQuery("min_rating", "0"), 64)
	minReviews, _ := strconv.Atoi(c.DefaultQuery("min_reviews", "0"))
	genresParam := c.Query("genres")
	genreMode := c.DefaultQuery("genre_mode", "or")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	query := h.DB.Model(&model.Novel{}).Preload("Genres")

	if q != "" {
		like := "%" + q + "%"
		query = query.Where("title ILIKE ? OR alt_title ILIKE ? OR description ILIKE ?", like, like, like)
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if genre != "" {
		query = query.Joins("JOIN novel_genres ON novel_genres.novel_id = novels.id").
			Joins("JOIN genres ON genres.id = novel_genres.genre_id").
			Where("genres.slug = ?", genre)
	}

	if genresParam != "" {
		genreSlugs := strings.Split(genresParam, ",")
		if genreMode == "and" {
			for _, slug := range genreSlugs {
				slug = strings.TrimSpace(slug)
				if slug == "" {
					continue
				}
				subQuery := h.DB.Table("novel_genres").
					Select("novel_id").
					Joins("JOIN genres ON genres.id = novel_genres.genre_id").
					Where("genres.slug = ?", slug)
				query = query.Where("novels.id IN (?)", subQuery)
			}
		} else {
			query = query.Joins("JOIN novel_genres ON novel_genres.novel_id = novels.id").
				Joins("JOIN genres ON genres.id = novel_genres.genre_id").
				Where("genres.slug IN ?", genreSlugs)
		}
	}

	if minChapters > 0 {
		query = query.Where("chapters >= ?", minChapters)
	}
	if minRating > 0 {
		query = query.Where("rating >= ?", minRating)
	}
	if minReviews > 0 {
		query = query.Where("rating_count >= ?", minReviews)
	}

	sortMap := map[string]string{
		"created_at": "novels.created_at",
		"title":      "novels.title",
		"views":      "novels.views",
		"chapters":   "novels.chapters",
		"rating":     "novels.rating",
		"readers":    "novels.readers",
		"reviews":    "novels.rating_count",
	}
	if col, ok := sortMap[sort]; ok {
		query = query.Order(col + " " + order)
	} else {
		query = query.Order("novels.created_at DESC")
	}

	var total int64
	query.Count(&total)

	var novels []model.Novel
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Find(&novels).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        novels,
		"page":        page,
		"limit":       limit,
		"total":       total,
		"total_pages": totalPages,
	})
}

func (h *NovelHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var novel model.Novel
	if err := h.DB.Preload("Genres").First(&novel, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "novel not found"})
		return
	}

	c.JSON(http.StatusOK, novel)
}

func (h *NovelHandler) Chapters(c *gin.Context) {
	novelID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 50
	}

	var total int64
	h.DB.Model(&model.Chapter{}).Where("novel_id = ?", novelID).Count(&total)

	var chapters []model.Chapter
	offset := (page - 1) * limit
	if err := h.DB.Where("novel_id = ?", novelID).
		Order("number ASC").
		Offset(offset).Limit(limit).
		Find(&chapters).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":    chapters,
		"page":    page,
		"limit":   limit,
		"total":   total,
	})
}

func (h *NovelHandler) Random(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 || limit > 50 {
		limit = 10
	}

	var novels []model.Novel
	if err := h.DB.Preload("Genres").
		Order("RANDOM()").
		Limit(limit).
		Find(&novels).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": novels})
}

func (h *NovelHandler) Trending(c *gin.Context) {
	var novels []model.Novel
	if err := h.DB.Preload("Genres").
		Order("views DESC").
		Limit(20).
		Find(&novels).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": novels})
}

func (h *NovelHandler) Recommendations(c *gin.Context) {
	var novels []model.Novel
	if err := h.DB.Preload("Genres").
		Order("rating DESC, views DESC").
		Limit(12).
		Find(&novels).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": novels})
}

type CreateNovelRequest struct {
	Title       string   `json:"title" binding:"required"`
	AltTitle    string   `json:"alt_title"`
	Author      string   `json:"author"`
	Status      string   `json:"status"`
	Description string   `json:"description"`
	CoverURL    string   `json:"cover_url"`
	GenreIDs    []uint   `json:"genre_ids"`
}

func (h *NovelHandler) Create(c *gin.Context) {
	var req CreateNovelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	slug := generateSlug(req.Title)

	novel := model.Novel{
		Title:       req.Title,
		AltTitle:    req.AltTitle,
		Slug:        slug,
		Author:      req.Author,
		AuthorSlug:  generateSlug(req.Author),
		Status:      req.Status,
		Description: req.Description,
		CoverURL:    req.CoverURL,
	}

	if novel.Status == "" {
		novel.Status = "ongoing"
	}

	var genres []model.Genre
	if len(req.GenreIDs) > 0 {
		h.DB.Where("id IN ?", req.GenreIDs).Find(&genres)
	}

	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&novel).Error; err != nil {
			return err
		}
		if len(genres) > 0 {
			if err := tx.Model(&novel).Association("Genres").Append(genres); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.DB.Preload("Genres").First(&novel, novel.ID)
	c.JSON(http.StatusCreated, gin.H{"data": novel})
}

type UpdateNovelRequest struct {
	Title       string   `json:"title"`
	AltTitle    string   `json:"alt_title"`
	Author      string   `json:"author"`
	Status      string   `json:"status"`
	Description string   `json:"description"`
	CoverURL    string   `json:"cover_url"`
	GenreIDs    []uint   `json:"genre_ids"`
}

func (h *NovelHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var novel model.Novel
	if err := h.DB.Preload("Genres").First(&novel, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "novel not found"})
		return
	}

	var req UpdateNovelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Title != "" {
		updates["title"] = req.Title
		updates["slug"] = generateSlug(req.Title)
	}
	if req.AltTitle != "" {
		updates["alt_title"] = req.AltTitle
	}
	if req.Author != "" {
		updates["author"] = req.Author
		updates["author_slug"] = generateSlug(req.Author)
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.CoverURL != "" {
		updates["cover_url"] = req.CoverURL
	}

	err = h.DB.Transaction(func(tx *gorm.DB) error {
		if len(updates) > 0 {
			if err := tx.Model(&novel).Updates(updates).Error; err != nil {
				return err
			}
		}
		if req.GenreIDs != nil {
			var genres []model.Genre
			if len(req.GenreIDs) > 0 {
				tx.Where("id IN ?", req.GenreIDs).Find(&genres)
			}
			if err := tx.Model(&novel).Association("Genres").Replace(genres); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.DB.Preload("Genres").First(&novel, novel.ID)
	c.JSON(http.StatusOK, gin.H{"data": novel})
}

func (h *NovelHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var novel model.Novel
	if err := h.DB.First(&novel, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "novel not found"})
		return
	}

	err = h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&novel).Association("Genres").Clear(); err != nil {
			return err
		}
		if err := tx.Delete(&novel).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "novel deleted"})
}

func generateSlug(s string) string {
	slug := strings.ToLower(s)
	slug = strings.ReplaceAll(slug, " ", "-")
	replacer := strings.NewReplacer(
		".", "", ",", "", "!", "", "?", "", "'", "", "\"", "",
		":", "", ";", "", "(", "", ")", "", "[", "", "]", "",
		"{", "", "}", "", "/", "-", "&", "and",
	)
	slug = replacer.Replace(slug)
	slug = strings.Trim(slug, "-")
	if len(slug) > 200 {
		slug = slug[:200]
	}
	slug = strings.TrimSuffix(slug, "-")
	return slug
}
