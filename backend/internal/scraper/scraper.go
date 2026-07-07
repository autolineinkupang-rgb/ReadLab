package scraper

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var htmlTagRe = regexp.MustCompile(`<[^>]*>`)

func stripHTML(s string) string {
	s = htmlTagRe.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&quot;", "\"")
	return strings.TrimSpace(s)
}

type ScrapedNovel struct {
	Title       string
	AltTitle    string
	Author      string
	Status      string
	Description string
	CoverURL    string
	Genres      []string
	Chapters    []ScrapedChapter
	SourceURL   string
}

type ScrapedChapter struct {
	Number  int
	Title   string
	URL     string
	Content string
}

type Result struct {
	Novel  *ScrapedNovel
	Error  error
}

type Scraper struct {
	client *http.Client
}

func New() *Scraper {
	return &Scraper{
		client: &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) > 5 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
	}
}

func (s *Scraper) fetchDoc(url string) (*goquery.Document, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse failed: %w", err)
	}

	return doc, nil
}

func (s *Scraper) ScrapeNovel(url string) (*ScrapedNovel, error) {
	switch {
	case strings.Contains(url, "royalroad"):
		return s.scrapeRoyalRoad(url)
	case strings.Contains(url, "novelbin") || strings.Contains(url, "novelbin.net"):
		return s.scrapeNovelBin(url)
	case strings.Contains(url, "freewebnovel"):
		return s.scrapeFreeWebNovel(url)
	case strings.Contains(url, "novelupdates"):
		return s.scrapeNovelUpdates(url)
	default:
		return nil, fmt.Errorf("unsupported site: %s", url)
	}
}

func (s *Scraper) ScrapeChapter(url string) (*ScrapedChapter, error) {
	switch {
	case strings.Contains(url, "royalroad"):
		return s.scrapeRoyalRoadChapter(url)
	case strings.Contains(url, "novelbin") || strings.Contains(url, "novelbin.net"):
		return s.scrapeNovelBinChapter(url)
	case strings.Contains(url, "freewebnovel"):
		return s.scrapeFreeWebNovelChapter(url)
	default:
		return nil, fmt.Errorf("unsupported site: %s", url)
	}
}
