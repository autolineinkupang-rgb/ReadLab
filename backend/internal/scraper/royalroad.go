package scraper

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func (s *Scraper) scrapeRoyalRoad(url string) (*ScrapedNovel, error) {
	doc, err := s.fetchDoc(url)
	if err != nil {
		return nil, err
	}

	novel := &ScrapedNovel{
		SourceURL: url,
	}

	novel.Title = strings.TrimSpace(doc.Find("h1.font-white, h1[property='name'], h1").First().Text())

	doc.Find("div.fic-header, .col-md-7, .fiction-info").First().Find("div.small, span, p").Each(func(i int, sel *goquery.Selection) {
		sel.Find("a").Each(func(j int, a *goquery.Selection) {
			href, exists := a.Attr("href")
			if exists && strings.Contains(href, "/profile/") {
				novel.Author = strings.TrimSpace(a.Text())
			}
		})
	})

	doc.Find("meta[property='books:author']").Each(func(i int, meta *goquery.Selection) {
		content, exists := meta.Attr("content")
		if exists && content != "" && novel.Author == "" {
			novel.Author = content
		}
	})

	doc.Find("meta[name='description'], meta[property='og:description']").Each(func(i int, meta *goquery.Selection) {
		content, exists := meta.Attr("content")
		if exists && content != "" && novel.Description == "" {
			novel.Description = content
		}
	})

	if novel.Description == "" {
		doc.Find("div.description, .description div, [property='description']").Each(func(i int, sel *goquery.Selection) {
			text := strings.TrimSpace(sel.Text())
			if len(text) > 100 {
				novel.Description = text
			}
		})
	}

	doc.Find("meta[property='og:image']").Each(func(i int, meta *goquery.Selection) {
		content, exists := meta.Attr("content")
		if exists && content != "" && novel.CoverURL == "" {
			novel.CoverURL = content
		}
	})

	if novel.CoverURL == "" {
		doc.Find("img.fic-image, img[alt*='cover'], .cover-image img").Each(func(i int, img *goquery.Selection) {
			src, exists := img.Attr("src")
			if exists && src != "" {
				novel.CoverURL = src
			}
		})
	}

	doc.Find("span.tags a, .tags a, .fic-header a[href*='/genre/'], a[href*='genre']").Each(func(i int, a *goquery.Selection) {
		g := strings.TrimSpace(a.Text())
		if g != "" {
			novel.Genres = append(novel.Genres, g)
		}
	})

	statusText := strings.TrimSpace(doc.Find("div.fic-status, span.fic-status, .status").First().Text())
	switch strings.ToLower(statusText) {
	case "ongoing", "active":
		novel.Status = "ongoing"
	case "completed", "complete":
		novel.Status = "completed"
	case "hiatus":
		novel.Status = "hiatus"
	default:
		novel.Status = "ongoing"
	}

	doc.Find("table.chapter-table tbody tr, .chapter-row, .chapter-list tr, table.table tbody tr").Each(func(i int, tr *goquery.Selection) {
		link := tr.Find("a").First()
		href, exists := link.Attr("href")
		if !exists || href == "" {
			return
		}
		chTitle := strings.TrimSpace(link.Text())
		if chTitle == "" {
			chTitle = fmt.Sprintf("Chapter %d", len(novel.Chapters)+1)
		}
		chURL := href
		if strings.HasPrefix(href, "/") {
			chURL = "https://www.royalroad.com" + href
		}

		novel.Chapters = append(novel.Chapters, ScrapedChapter{
			Number: len(novel.Chapters) + 1,
			Title:  chTitle,
			URL:    chURL,
		})
	})

	if len(novel.Chapters) == 0 {
		doc.Find("a[href*='/chapter/']").Each(func(i int, a *goquery.Selection) {
			href, exists := a.Attr("href")
			if !exists || href == "" {
				return
			}
			chURL := href
			if strings.HasPrefix(href, "/") {
				chURL = "https://www.royalroad.com" + href
			}
			novel.Chapters = append(novel.Chapters, ScrapedChapter{
				Number: len(novel.Chapters) + 1,
				Title:  strings.TrimSpace(a.Text()),
				URL:    chURL,
			})
		})
	}

	if novel.Title == "" {
		return nil, fmt.Errorf("could not extract novel info from %s", url)
	}

	return novel, nil
}

func (s *Scraper) scrapeRoyalRoadChapter(url string) (*ScrapedChapter, error) {
	doc, err := s.fetchDoc(url)
	if err != nil {
		return nil, err
	}

	content := ""
	doc.Find("div.chapter-content, .chapter-inner, div.description, .portlet-body").Each(func(i int, sel *goquery.Selection) {
		html, err := sel.Html()
		if err == nil {
			content = strings.TrimSpace(html)
		}
	})

	if content == "" {
		doc.Find("div#content, .content, .story-content").Each(func(i int, sel *goquery.Selection) {
			html, err := sel.Html()
			if err == nil {
				content = strings.TrimSpace(html)
			}
		})
	}

	title := strings.TrimSpace(doc.Find("h1, h2, .chapter-title, .title").First().Text())

	return &ScrapedChapter{
		Title:   title,
		URL:     url,
		Content: content,
	}, nil
}
