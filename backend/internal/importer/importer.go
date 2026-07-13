package importer

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gorm.io/gorm"
	"readlab/backend/internal/model"
)

const consumetBase = "https://api.consumet.org/light-novels/novelupdates"

type SearchResult struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
	Image string `json:"image"`
}

type SearchResponse struct {
	CurrentPage int            `json:"currentPage"`
	HasNextPage bool           `json:"hasNextPage"`
	Results     []SearchResult `json:"results"`
}

type ChapterInfo struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

type NovelInfo struct {
	ID          string        `json:"id"`
	Title       string        `json:"title"`
	URL         string        `json:"url"`
	Image       string        `json:"image"`
	Description string        `json:"description"`
	Genres      []string      `json:"genres"`
	Status      string        `json:"status"`
	Authors     []string      `json:"authors"`
	Chapters    []ChapterInfo `json:"chapters"`
}

type Importer struct {
	DB        *gorm.DB
	HTTPClient *http.Client
}

func New(db *gorm.DB) *Importer {
	return &Importer{
		DB: db,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (imp *Importer) Search(query string) ([]SearchResult, error) {
	encoded := url.PathEscape(query)
	resp, err := imp.HTTPClient.Get(consumetBase + "/" + encoded)
	if err != nil {
		return nil, fmt.Errorf("consumet search request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("consumet search read failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("consumet search returned status %d: %s", resp.StatusCode, string(body))
	}

	var result SearchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("consumet search parse failed: %w", err)
	}

	return result.Results, nil
}

func (imp *Importer) FetchNovelInfo(sourceID string) (*NovelInfo, error) {
	resp, err := imp.HTTPClient.Get(consumetBase + "/info?id=" + url.QueryEscape(sourceID))
	if err != nil {
		return nil, fmt.Errorf("consumet info request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("consumet info read failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("consumet info returned status %d: %s", resp.StatusCode, string(body))
	}

	var info NovelInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("consumet info parse failed: %w", err)
	}

	return &info, nil
}

func (imp *Importer) ImportNovel(sourceID string, withChapters bool) (*model.Novel, error) {
	info, err := imp.FetchNovelInfo(sourceID)
	if err != nil {
		return nil, err
	}

	if info.Title == "" {
		return nil, fmt.Errorf("novel not found for id: %s", sourceID)
	}

	slug := generateSlug(info.Title)
	author := ""
	if len(info.Authors) > 0 {
		author = info.Authors[0]
	}

	status := "ongoing"
	switch strings.ToLower(info.Status) {
	case "completed":
		status = "completed"
	case "hiatus":
		status = "hiatus"
	case "dropped":
		status = "dropped"
	}

	novel := model.Novel{
		Title:       info.Title,
		Slug:        slug,
		Author:      author,
		AuthorSlug:  generateSlug(author),
		Status:      status,
		Description: info.Description,
		CoverURL:    info.Image,
		Chapters:    len(info.Chapters),
		AddedAt:     time.Now(),
	}

	var matchedGenres []model.Genre
	for _, gName := range info.Genres {
		slug := strings.ToLower(strings.ReplaceAll(gName, " ", "-"))
		var genre model.Genre
		err := imp.DB.Where("LOWER(name) = ? OR slug = ?", strings.ToLower(gName), slug).First(&genre).Error
		if err != nil {
			genre = model.Genre{Slug: slug, Name: gName}
			if err := imp.DB.Create(&genre).Error; err != nil {
				continue
			}
		}
		matchedGenres = append(matchedGenres, genre)
	}

	err = imp.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&novel).Error; err != nil {
			return err
		}
		if len(matchedGenres) > 0 {
			if err := tx.Model(&novel).Association("Genres").Append(matchedGenres); err != nil {
				return err
			}
		}
		if withChapters && len(info.Chapters) > 0 {
			chapters := make([]model.Chapter, 0, len(info.Chapters))
			for i, ch := range info.Chapters {
				chapters = append(chapters, model.Chapter{
					NovelID: novel.ID,
					Number:  i + 1,
					Title:   ch.Title,
				})
			}
			if err := tx.Create(&chapters).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("import transaction failed: %w", err)
	}

	imp.DB.Preload("Genres").First(&novel, novel.ID)
	return &novel, nil
}

func generateSlug(title string) string {
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.NewReplacer(
		".", "", ",", "", "!", "", "?", "", "'", "", "\"", "",
		":", "", ";", "", "(", "", ")", "", "[", "", "]", "",
		"{", "", "}", "", "/", "-", "&", "and",
	).Replace(slug)
	slug = strings.Trim(slug, "-")
	if len(slug) > 200 {
		slug = slug[:200]
	}
	slug = strings.TrimSuffix(slug, "-")
	return slug
}
