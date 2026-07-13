package main

import (
	"log"
	"path/filepath"
	"strings"
	"time"

	"readlab/backend/internal/config"
	"readlab/backend/internal/model"
	"readlab/backend/internal/scraper"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	var novels []model.Novel
	db.Order("id ASC").Find(&novels)
	log.Printf("found %d novels\n", len(novels))

	s := scraper.New()
	totalChapters := 0

	for i, novel := range novels {
		if novel.SourceURL == "" || !strings.Contains(novel.SourceURL, "novelfire.net") {
			log.Printf("[%d/%d] skipping %s (no novelfire source)\n", i+1, len(novels), novel.Title)
			continue
		}

		slug := filepath.Base(novel.SourceURL)
		log.Printf("[%d/%d] scraping chapters for: %s (slug=%s)\n", i+1, len(novels), novel.Title, slug)

		chapters, err := s.ScrapeNovelfireChapters(slug)
		if err != nil {
			log.Printf("  failed: %v\n", err)
			continue
		}

		if len(chapters) == 0 {
			log.Printf("  no chapters found\n")
			continue
		}

		toInsert := make([]model.Chapter, 0, len(chapters))
		for _, ch := range chapters {
			toInsert = append(toInsert, model.Chapter{
				NovelID:   novel.ID,
				Number:    ch.Number,
				Title:     ch.Title,
				Content:   "",
				ContentMD: "",
				IsLocked:  false,
				TicketCost: 0,
			})
		}

		err = db.CreateInBatches(toInsert, 500).Error
		if err != nil {
			log.Printf("  db insert failed: %v\n", err)
			continue
		}

		chapterCount := len(toInsert)
		db.Model(&novel).Update("chapters", chapterCount)

		totalChapters += chapterCount
		log.Printf("  inserted %d chapters (total so far: %d)\n", chapterCount, totalChapters)

		time.Sleep(200 * time.Millisecond)
	}

	log.Printf("done! total chapters inserted: %d\n", totalChapters)
}
