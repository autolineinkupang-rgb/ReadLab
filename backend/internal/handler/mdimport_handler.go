package handler

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"wtr-lab-clone/backend/internal/mdimport"
	"wtr-lab-clone/backend/internal/model"
)

type MdImportHandler struct {
	DB *gorm.DB
}

func NewMdImportHandler(db *gorm.DB) *MdImportHandler {
	return &MdImportHandler{DB: db}
}

type MdImportPreviewResp struct {
	Chapters []mdimport.ParsedChapter `json:"chapters"`
	Warnings []string                 `json:"warnings"`
}

type MdImportSaveResp struct {
	Message  string `json:"message"`
	Imported int    `json:"imported"`
	Skipped  int    `json:"skipped"`
}

func (h *MdImportHandler) Import(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	uid := userID.(uint)

	novelID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid novel id"})
		return
	}

	var novel model.Novel
	if err := h.DB.First(&novel, novelID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "novel not found"})
		return
	}
	role, _ := c.Get("role")
	if role != "admin" && novel.WriterID != nil && *novel.WriterID != uid {
		c.JSON(http.StatusForbidden, gin.H{"error": "you do not have access to this novel"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
		return
	}

	if len(data) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty file"})
		return
	}

	mode := c.DefaultPostForm("mode", "preview")
	filename := strings.ToLower(header.Filename)
	contentType := header.Header.Get("Content-Type")

	var chapters []mdimport.ParsedChapter
	var warnings []string

	if strings.HasSuffix(filename, ".zip") || contentType == "application/zip" {
		files, err := mdimport.ExtractMDsFromZip(data)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid zip file"})
			return
		}
		if files == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no markdown files found in zip"})
			return
		}
		offset := getLastChapterNumber(h.DB, uint(novelID))
		i := 0
		for name, content := range files {
			title := strings.TrimSuffix(name, ".md")
			title = strings.TrimSuffix(title, ".markdown")
			ch := mdimport.ParseChapterMD(content, title)
			i++
			ch.Number = offset + i
			ch.Exists = chapterExists(h.DB, uint(novelID), ch.Number)
			chapters = append(chapters, *ch)
		}
	} else if strings.HasSuffix(filename, ".md") || strings.HasSuffix(filename, ".markdown") {
		result := mdimport.ParseSingleMD(string(data))
		chapters = result.Chapters
		warnings = result.Warnings

		offset := getLastChapterNumber(h.DB, uint(novelID))
		for i := range chapters {
			chapters[i].Number = offset + i + 1
			chapters[i].Exists = chapterExists(h.DB, uint(novelID), chapters[i].Number)
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only .md, .markdown, and .zip files are supported"})
		return
	}

	if mode == "save" {
		imported := 0
		skipped := 0
		for _, ch := range chapters {
			if ch.Exists {
				skipped++
				continue
			}
			chapter := model.Chapter{
				NovelID:   uint(novelID),
				Number:    ch.Number,
				Title:     ch.Title,
				Content:   ch.ContentHTML,
				ContentMD: ch.ContentMD,
			}
			if err := h.DB.Create(&chapter).Error; err != nil {
				warnings = append(warnings, "failed to import chapter "+strconv.Itoa(ch.Number)+": "+err.Error())
				skipped++
				continue
			}
			imported++
		}

		var count int64
		h.DB.Model(&model.Chapter{}).Where("novel_id = ?", novelID).Count(&count)
		h.DB.Model(&model.Novel{}).Where("id = ?", novelID).Update("chapters", count)

		c.JSON(http.StatusOK, MdImportSaveResp{
			Message:  strconv.Itoa(imported) + " chapters imported successfully",
			Imported: imported,
			Skipped:  skipped,
		})
		return
	}

	c.JSON(http.StatusOK, MdImportPreviewResp{
		Chapters: chapters,
		Warnings: warnings,
	})
}

func getLastChapterNumber(db *gorm.DB, novelID uint) int {
	var max struct{ Max int }
	db.Model(&model.Chapter{}).Select("COALESCE(MAX(number), 0) as max").Where("novel_id = ?", novelID).Scan(&max)
	return max.Max
}

func chapterExists(db *gorm.DB, novelID uint, number int) bool {
	var count int64
	db.Model(&model.Chapter{}).Where("novel_id = ? AND number = ?", novelID, number).Count(&count)
	return count > 0
}
