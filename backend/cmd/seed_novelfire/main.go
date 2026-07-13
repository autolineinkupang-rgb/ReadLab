package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"readlab/backend/internal/config"
	"readlab/backend/internal/model"
	"readlab/backend/internal/scraper"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var coversDir string

func main() {
	cfg := config.Load()

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	cwd, _ := os.Getwd()
	coversDir = filepath.Join(cwd, "uploads", "covers")
	os.MkdirAll(coversDir, 0755)

	log.Println("connected to database")
	log.Println("covers directory:", coversDir)

	log.Println("truncating existing novels, chapters, and novel_genres...")
	db.Exec("DELETE FROM novel_genres")
	db.Exec("DELETE FROM chapters")
	db.Exec("DELETE FROM novels")

	s := scraper.New()

	log.Println("scraping novel list from novelfire.net/ranking...")
	items, err := s.ScrapeNovelfireList()
	if err != nil {
		log.Fatalf("failed to scrape novel list: %v", err)
	}
	log.Printf("found %d novels\n", len(items))

	for i, item := range items {
		url := "https://novelfire.net/book/" + item.Slug
		log.Printf("[%d/%d] scraping: %s\n", i+1, len(items), item.Title)

		novel, err := s.ScrapeNovel(url)
		if err != nil {
			log.Printf("  failed to scrape %s: %v\n", item.Title, err)
			continue
		}

		chapters := scrapeChapterCount(url)

		slug := generateSlug(novel.Title)
		coverURL := ""

		if novel.CoverURL != "" {
			ext := ".jpg"
			coverPath := filepath.Join(coversDir, slug+ext)
			if err := downloadImage(novel.CoverURL, coverPath); err != nil {
				log.Printf("  failed to download cover: %v\n", err)
				coverURL = novel.CoverURL
			} else {
				coverURL = "/api/novel-covers/" + slug + ext
				log.Printf("  cover saved: %s\n", coverPath)
			}
		}

		status := novel.Status
		if status == "" {
			status = "ongoing"
		}

		authorSlug := generateSlug(novel.Author)

		dbNovel := model.Novel{
			Title:       novel.Title,
			AltTitle:    novel.AltTitle,
			Slug:        slug,
			Author:      novel.Author,
			AuthorSlug:  authorSlug,
			Status:      status,
			Views:       uint64(1000 + i*500),
			Rating:      0,
			RatingCount: 0,
			Chapters:    chapters,
			Readers:     10 + i*5,
			Chars:       "",
			AIPercent:   "",
			Description: novel.Description,
			CoverURL:    coverURL,
			SourceURL:   url,
		}

		if err := db.Create(&dbNovel).Error; err != nil {
			if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique") {
				log.Printf("  skipped (duplicate): %s\n", novel.Title)
				continue
			}
			log.Printf("  failed to create novel: %v\n", err)
			continue
		}

		for _, gName := range novel.Genres {
			genreSlug := generateSlug(gName)
			var genre model.Genre
			err := db.Where("LOWER(name) = ? OR slug = ?", strings.ToLower(gName), genreSlug).First(&genre).Error
			if err != nil {
				genre = model.Genre{Slug: genreSlug, Name: gName}
				if err := db.Create(&genre).Error; err != nil {
					continue
				}
			}
			db.Create(&model.NovelGenre{NovelID: dbNovel.ID, GenreID: genre.ID})
		}

		log.Printf("  created: %s (id=%d, chapters=%d)\n", novel.Title, dbNovel.ID, chapters)

		time.Sleep(500 * time.Millisecond)
	}

	log.Println("seeding users...")
	seedUsers(db)

	log.Println("seeding news...")
	seedNews(db)

	log.Println("seed novelfire completed!")
}

func scrapeChapterCount(url string) int {
	client := &http.Client{Timeout: 15 * time.Second}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	resp, err := client.Do(req)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return 0
	}

	count := 0
	doc.Find(".header-stats span strong").Each(func(i int, strong *goquery.Selection) {
		if i == 0 {
			text := strings.TrimSpace(strong.Text())
			text = strings.ReplaceAll(text, ",", "")
			text = strings.ReplaceAll(text, ".", "")
			if n, err := strconv.Atoi(text); err == nil {
				count = n
			}
		}
	})

	return count
}

func downloadImage(url, path string) error {
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36")
	req.Header.Set("Referer", "https://novelfire.net/")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
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

func seedUsers(db *gorm.DB) {
	users := []struct {
		Username string
		Email    string
		Password string
		Tickets  float64
		IsAdmin  bool
	}{
		{"Mega_bells", "mega@example.com", "password", 3569.76, false},
		{"StandardCrystal", "crystal@example.com", "password", 2907.17, false},
		{"Alpha2", "alpha2@example.com", "password", 2693.07, false},
		{"reader1", "reader1@example.com", "password", 100, false},
	}

	for _, u := range users {
		var existing model.User
		if err := db.Where("email = ?", u.Email).First(&existing).Error; err == nil {
			continue
		}

		hash, _ := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		role := "member"
		if u.IsAdmin {
			role = "admin"
		}
		user := model.User{
			Username:     u.Username,
			Email:        u.Email,
			PasswordHash: string(hash),
			DisplayName:  u.Username,
			Tickets:      u.Tickets,
			Role:         role,
		}
		db.Create(&user)
		fmt.Printf("seeded user: %s\n", u.Username)
	}
}

func seedNews(db *gorm.DB) {
	news := []struct {
		Title   string
		Content string
		Type    string
		Slug    string
	}{
		{
			Title:   "New Novels Added from NovelFire",
			Content: "We have updated our library with the latest popular novels from NovelFire. Enjoy reading!",
			Type:    "news",
			Slug:    "new-novels-from-novelfire",
		},
		{
			Title:   "Welcome to ReadLab!",
			Content: "ReadLab is now live with a fresh collection of web novels and light novels.",
			Type:    "news",
			Slug:    "welcome-to-readlab",
		},
	}

	for _, n := range news {
		var existing model.News
		if err := db.Where("slug = ?", n.Slug).First(&existing).Error; err == nil {
			continue
		}
		db.Create(&model.News{
			Title:   n.Title,
			Content: n.Content,
			Type:    n.Type,
			Slug:    n.Slug,
		})
		fmt.Printf("seeded news: %s\n", n.Title)
	}
}
