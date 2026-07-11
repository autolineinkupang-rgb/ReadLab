// Command import_real imports real novels from public sources using the
// existing scraper package. Meant to be run once to populate the DB with
// real content (covers, descriptions, tags, chapters).
//
// Usage:
//   go run ./cmd/import_real [--with-content] [--max-chapters=N]
//
// Sources: RoyalRoad (public, well-formed HTML, no anti-bot).
package main

import (
	"flag"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	"wtr-lab-clone/backend/internal/config"
	"wtr-lab-clone/backend/internal/model"
	"wtr-lab-clone/backend/internal/scraper"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Curated list of well-known public novels on RoyalRoad with good metadata.
var realNovels = []string{
	"https://www.royalroad.com/fiction/21220/mother-of-learning",
	"https://www.royalroad.com/fiction/25137/the-perfect-run",
	"https://www.royalroad.com/fiction/40373/beware-of-chicken",
	"https://www.royalroad.com/fiction/49033/super-supportive",
	"https://www.royalroad.com/fiction/36735/the-primal-hunter",
	"https://www.royalroad.com/fiction/26534/i-am-a-book",
	"https://www.royalroad.com/fiction/59450/vainqueur-the-dragon",
	"https://www.royalroad.com/fiction/45534/the-daily-grind",
	"https://www.royalroad.com/fiction/22518/millennial-mage-a-slice-of-life-progression",
	"https://www.royalroad.com/fiction/57457/necrotic-apocalypse",
	"https://www.royalroad.com/fiction/58965/vigor-mortis",
	"https://www.royalroad.com/fiction/49482/dungeon-crawler-carl",
}

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

func generateSlug(s string) string {
	s = strings.ToLower(s)
	s = slugRe.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 200 {
		s = s[:200]
	}
	return s
}

func main() {
	withContent := flag.Bool("with-content", false, "also scrape full chapter content (slow, expensive)")
	maxChapters := flag.Int("max-chapters", 20, "max chapters to import per novel")
	limitNovels := flag.Int("limit", 0, "only import first N novels (0 = all)")
	flag.Parse()

	cfg := config.Load()
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{Logger: nil})
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}

	sc := scraper.New()

	urls := realNovels
	if *limitNovels > 0 && *limitNovels < len(urls) {
		urls = urls[:*limitNovels]
	}

	log.Printf("importing %d real novels (with_content=%v, max_chapters=%d)", len(urls), *withContent, *maxChapters)

	success := 0
	failed := 0
	for i, url := range urls {
		fmt.Printf("\n[%d/%d] Scraping: %s\n", i+1, len(urls), url)
		start := time.Now()

		scraped, err := sc.ScrapeNovel(url)
		if err != nil {
			log.Printf("  ✗ scrape failed: %v", err)
			failed++
			continue
		}
		fmt.Printf("  ✓ %s (%d chapters found, cover=%v)\n",
			scraped.Title, len(scraped.Chapters), scraped.CoverURL != "")

		if err := importNovel(db, sc, scraped, url, *withContent, *maxChapters); err != nil {
			log.Printf("  ✗ import failed: %v", err)
			failed++
			continue
		}
		success++
		fmt.Printf("  ✓ imported in %s\n", time.Since(start).Round(time.Millisecond))
	}

	fmt.Printf("\n=== Done: %d success, %d failed ===\n", success, failed)
}

func importNovel(db *gorm.DB, sc *scraper.Scraper, novel *scraper.ScrapedNovel, srcURL string, withContent bool, maxChapters int) error {
	// Check for existing novel
	var existing model.Novel
	if err := db.Where("source_url = ? OR LOWER(title) = LOWER(?)", srcURL, novel.Title).First(&existing).Error; err == nil {
		fmt.Printf("  → already exists (id=%d), updating metadata + chapters\n", existing.ID)
		return updateExisting(db, sc, &existing, novel, withContent, maxChapters)
	}

	status := novel.Status
	if status == "" {
		status = "ongoing"
	}

	dbNovel := model.Novel{
		Title:       novel.Title,
		AltTitle:    novel.AltTitle,
		Slug:        generateSlug(novel.Title),
		Author:      novel.Author,
		AuthorSlug:  generateSlug(novel.Author),
		Status:      status,
		Description: novel.Description,
		CoverURL:    novel.CoverURL,
		SourceURL:   srcURL,
		Chapters:    0,
		Views:       uint64(1000 + i()), // small view count for ordering
		Rating:      4.0,
		RatingCount: 25,
	}

	if err := db.Create(&dbNovel).Error; err != nil {
		return fmt.Errorf("create novel: %w", err)
	}

	// Attach genres (create if not exist)
	genres := matchOrCreateGenres(db, novel.Genres)
	if len(genres) > 0 {
		db.Model(&dbNovel).Association("Genres").Append(genres)
	}

	// Insert chapters
	return insertChapters(db, sc, dbNovel.ID, novel.Chapters, withContent, maxChapters)
}

func updateExisting(db *gorm.DB, sc *scraper.Scraper, existing *model.Novel, novel *scraper.ScrapedNovel, withContent bool, maxChapters int) error {
	updates := map[string]interface{}{}
	if existing.CoverURL == "" && novel.CoverURL != "" {
		updates["cover_url"] = novel.CoverURL
	}
	if existing.Description == "" && novel.Description != "" {
		updates["description"] = novel.Description
	}
	if existing.Author == "" && novel.Author != "" {
		updates["author"] = novel.Author
		updates["author_slug"] = generateSlug(novel.Author)
	}
	if len(updates) > 0 {
		db.Model(existing).Updates(updates)
	}
	return insertChapters(db, sc, existing.ID, novel.Chapters, withContent, maxChapters)
}

func matchOrCreateGenres(db *gorm.DB, names []string) []model.Genre {
	seen := map[string]bool{}
	result := []model.Genre{}
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" || seen[strings.ToLower(name)] {
			continue
		}
		seen[strings.ToLower(name)] = true
		slug := generateSlug(name)
		var g model.Genre
		if err := db.Where("slug = ? OR LOWER(name) = ?", slug, strings.ToLower(name)).First(&g).Error; err != nil {
			g = model.Genre{Slug: slug, Name: name}
			if err := db.Create(&g).Error; err != nil {
				continue
			}
			fmt.Printf("    + new genre: %s\n", name)
		}
		result = append(result, g)
	}
	return result
}

func insertChapters(db *gorm.DB, sc *scraper.Scraper, novelID uint, chapters []scraper.ScrapedChapter, withContent bool, maxChapters int) error {
	if maxChapters > 0 && len(chapters) > maxChapters {
		chapters = chapters[:maxChapters]
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	created := 0

	sem := make(chan struct{}, 5) // concurrency limit for chapter content scrape

	for _, sc2 := range chapters {
		var count int64
		db.Model(&model.Chapter{}).Where("novel_id = ? AND number = ?", novelID, sc2.Number).Count(&count)
		if count > 0 {
			continue // already exists
		}

		content := ""
		isLocked := false

		if withContent && sc2.URL != "" {
			wg.Add(1)
			go func(ch scraper.ScrapedChapter) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				scraped, err := sc.ScrapeChapter(ch.URL)
				if err != nil || scraped == nil {
					return
				}
				// Prefer listing title (e.g. "1. Good Morning Brother") over
				// scraped page <h1> which is usually the novel title on RoyalRoad.
				title := ch.Title
				if title == "" {
					title = scraped.Title
				}
				mu.Lock()
				defer mu.Unlock()
				db.Create(&model.Chapter{
					NovelID:  novelID,
					Number:   ch.Number, // use listing's number (scraper.ScrapedChapter for individual page doesn't set Number)
					Title:    title,
					Content:  scraped.Content,
					IsLocked: false,
				})
				created++
			}(sc2)
		} else {
			db.Create(&model.Chapter{
				NovelID:  novelID,
				Number:   sc2.Number,
				Title:    sc2.Title,
				Content:  content,
				IsLocked: isLocked,
			})
			created++
		}
	}
	wg.Wait()

	// Update chapter count
	var totalChapters int64
	db.Model(&model.Chapter{}).Where("novel_id = ?", novelID).Count(&totalChapters)
	db.Model(&model.Novel{}).Where("id = ?", novelID).UpdateColumn("chapters", totalChapters)

	fmt.Printf("    + %d new chapters (total: %d)\n", created, totalChapters)
	return nil
}

var counter = 0

func i() int {
	counter += 1000
	return counter
}
