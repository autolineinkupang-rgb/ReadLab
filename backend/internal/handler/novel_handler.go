package handler

import (
        "log/slog"
        "net/http"
        "strconv"
        "strings"

        "github.com/gin-gonic/gin"
        "gorm.io/gorm"
        "readlab/backend/internal/model"
        "readlab/backend/internal/scraper"
        "readlab/backend/internal/service"
        "readlab/backend/internal/ticket"
)

type NovelHandler struct {
        DB       *gorm.DB
        Config   *ticket.Config
        NovelSvc *service.NovelService
        Scraper  *scraper.Scraper
}

func NewNovelHandler(db *gorm.DB, cfg *ticket.Config, novelSvc *service.NovelService) *NovelHandler {
        return &NovelHandler{DB: db, Config: cfg, NovelSvc: novelSvc, Scraper: scraper.New()}
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
        writerIDStr := c.Query("writer_id")

        if page < 1 {
                page = 1
        }
        if limit < 1 || limit > 100 {
                limit = 20
        }

        query := h.DB.Model(&model.Novel{}).Preload("Genres").Preload("Tags")

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
        if writerIDStr != "" {
                writerID, err := strconv.Atoi(writerIDStr)
                if err == nil {
                        query = query.Where("writer_id = ?", writerID)
                }
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
        if err := h.DB.Preload("Genres").Preload("Tags").First(&novel, id).Error; err != nil {
                c.JSON(http.StatusNotFound, gin.H{"error": "novel not found"})
                return
        }

        c.JSON(http.StatusOK, novel)
}

type CreateChapterItem struct {
        Number  int    `json:"number"`
        Title   string `json:"title"`
        Content string `json:"content"`
}

type CreateNovelRequest struct {
        Title       string              `json:"title" binding:"required"`
        AltTitle    string              `json:"alt_title"`
        Author      string              `json:"author"`
        Status      string              `json:"status"`
        Description string              `json:"description"`
        CoverURL    string              `json:"cover_url"`
        Chars       string              `json:"chars"`
        AIPercent   string              `json:"ai_percent"`
        Rating      float64             `json:"rating"`
        SourceURL   string              `json:"source_url"`
        GenreIDs    []uint              `json:"genre_ids"`
        TagIDs      []uint              `json:"tag_ids"`
        Chapters    []CreateChapterItem `json:"chapters"`
}

func (h *NovelHandler) Create(c *gin.Context) {
        var req CreateNovelRequest
        if err := c.ShouldBindJSON(&req); err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
                return
        }

        slug := generateSlug(req.Title)

        userID, _ := c.Get("user_id")
        role, _ := c.Get("role")

        novel := model.Novel{
                Title:       req.Title,
                AltTitle:    req.AltTitle,
                Slug:        slug,
                Author:      req.Author,
                AuthorSlug:  generateSlug(req.Author),
                Status:      req.Status,
                Description: req.Description,
                CoverURL:    req.CoverURL,
                SourceURL:   req.SourceURL,
                Chars:       req.Chars,
                AIPercent:   req.AIPercent,
                Rating:      req.Rating,
                Chapters:    len(req.Chapters),
        }

        if novel.Status == "" {
                novel.Status = "ongoing"
        }

        if role == "writer" {
                uid, ok := userID.(uint)
                if !ok {
                        c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
                        return
                }
                novel.WriterID = &uid
        }

        var genres []model.Genre
        if len(req.GenreIDs) > 0 {
                h.DB.Where("id IN ?", req.GenreIDs).Find(&genres)
        }
        var tags []model.Tag
        if len(req.TagIDs) > 0 {
                h.DB.Where("id IN ?", req.TagIDs).Find(&tags)
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
                if len(tags) > 0 {
                        if err := tx.Model(&novel).Association("Tags").Append(tags); err != nil {
                                return err
                        }
                }
                for _, ch := range req.Chapters {
                        chapter := model.Chapter{
                                NovelID: novel.ID,
                                Number:  ch.Number,
                                Title:   ch.Title,
                                Content: ch.Content,
                        }
                        if err := tx.Create(&chapter).Error; err != nil {
                                return err
                        }
                }
                return nil
        })
        if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
        }

        h.DB.Preload("Genres").Preload("Tags").First(&novel, novel.ID)

        uid := userID.(uint)
        reward := h.Config.Get("novel_contribution")
        if err := h.DB.Transaction(func(tx *gorm.DB) error {
                tx.Create(&model.TicketUnit{
                        Serial: model.NewSerial(),
                        UserID: uid,
                        Amount: reward,
                        Status: "active",
                })
                tx.Create(&model.TicketTransaction{
                        UserID:  uid,
                        Amount:  reward,
                        Type:    "reward",
                        RefType: "novel_contribution",
                        Note:    "Novel contribution reward",
                })
                var sum float64
                tx.Model(&model.TicketUnit{}).
                        Where("user_id = ? AND status = 'active'", uid).
                        Select("COALESCE(SUM(amount), 0)").Scan(&sum)
                tx.Model(&model.User{}).Where("id = ?", uid).Update("tickets", sum)
                return nil
        }); err != nil {
                slog.Error("failed to award novel contribution tickets", "error", err)
        }

        c.JSON(http.StatusCreated, gin.H{"data": novel})
}

type UpdateNovelRequest struct {
        Title       string   `json:"title"`
        AltTitle    string   `json:"alt_title"`
        Author      string   `json:"author"`
        Status      string   `json:"status"`
        Description string   `json:"description"`
        CoverURL    string   `json:"cover_url"`
        Chars       string   `json:"chars"`
        AIPercent   string   `json:"ai_percent"`
        Rating      *float64 `json:"rating"`
        SourceURL   string   `json:"source_url"`
        GenreIDs    []uint   `json:"genre_ids"`
        TagIDs      []uint   `json:"tag_ids"`
}

func (h *NovelHandler) Update(c *gin.Context) {
        id, err := strconv.ParseUint(c.Param("id"), 10, 32)
        if err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
                return
        }

        var novel model.Novel
        if err := h.DB.Preload("Genres").Preload("Tags").First(&novel, id).Error; err != nil {
                c.JSON(http.StatusNotFound, gin.H{"error": "novel not found"})
                return
        }

        role, _ := c.Get("role")
        userID, _ := c.Get("user_id")
        uid, ok := userID.(uint)
        if !ok {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
                return
        }
        if role != "admin" && (novel.WriterID == nil || *novel.WriterID != uid) {
                c.JSON(http.StatusForbidden, gin.H{"error": "you do not have permission to edit this novel"})
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
        if req.Chars != "" {
                updates["chars"] = req.Chars
        }
        if req.AIPercent != "" {
                updates["ai_percent"] = req.AIPercent
        }
        if req.Rating != nil {
                updates["rating"] = *req.Rating
        }
        if req.SourceURL != "" {
                updates["source_url"] = req.SourceURL
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
                if req.TagIDs != nil {
                        var tags []model.Tag
                        if len(req.TagIDs) > 0 {
                                tx.Where("id IN ?", req.TagIDs).Find(&tags)
                        }
                        if err := tx.Model(&novel).Association("Tags").Replace(tags); err != nil {
                                return err
                        }
                }
                return nil
        })
        if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
        }

        h.DB.Preload("Genres").Preload("Tags").First(&novel, novel.ID)
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

        role, _ := c.Get("role")
        userID, _ := c.Get("user_id")
        uid, ok := userID.(uint)
        if !ok {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
                return
        }
        if role != "admin" && (novel.WriterID == nil || *novel.WriterID != uid) {
                c.JSON(http.StatusForbidden, gin.H{"error": "you do not permission to delete this novel"})
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