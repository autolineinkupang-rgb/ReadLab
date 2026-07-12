package scraper

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type NovelfireChapterItem struct {
	Number int
	Title  string
}

type NovelfireNovelItem struct {
	Title string
	Slug  string
	Cover string
}

func (s *Scraper) ScrapeNovelfireList() ([]NovelfireNovelItem, error) {
	doc, err := s.fetchDoc("https://novelfire.net/ranking")
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var items []NovelfireNovelItem

	doc.Find("a[href*='/book/']").Each(func(i int, a *goquery.Selection) {
		href, exists := a.Attr("href")
		if !exists || href == "" {
			return
		}
		slug := strings.TrimPrefix(href, "/book/")
		slug = strings.TrimSuffix(slug, "/")
		if slug == "" || seen[slug] {
			return
		}
		title := strings.TrimSpace(a.Text())
		if title == "" {
			title = a.AttrOr("title", "")
		}
		if title == "" {
			return
		}

		cover := ""
		a.Find("img").Each(func(j int, img *goquery.Selection) {
			src := img.AttrOr("data-src", "")
			if src == "" {
				src = img.AttrOr("src", "")
			}
			if src != "" && cover == "" && strings.Contains(src, "/server-") {
				if strings.HasPrefix(src, "//") {
					src = "https:" + src
				}
				if strings.HasPrefix(src, "/") {
					src = "https://novelfire.net" + src
				}
				cover = src
			}
		})

		seen[slug] = true
		items = append(items, NovelfireNovelItem{
			Title: title,
			Slug:  slug,
			Cover: cover,
		})
	})

	if len(items) == 0 {
		return nil, fmt.Errorf("no novels found on novelfire ranking page")
	}

	if len(items) > 60 {
		items = items[:60]
	}

	return items, nil
}

func (s *Scraper) scrapeNovelFire(urlStr string) (*ScrapedNovel, error) {
	doc, err := s.fetchDoc(urlStr)
	if err != nil {
		return nil, err
	}

	novel := &ScrapedNovel{
		SourceURL: urlStr,
	}

	novel.Title = strings.TrimSpace(doc.Find("h1.novel-title").First().Text())
	if novel.Title == "" {
		novel.Title = strings.TrimSpace(doc.Find("h1[itemprop='name']").First().Text())
	}

	doc.Find("span[itemprop='author']").Each(func(i int, sel *goquery.Selection) {
		if novel.Author == "" {
			novel.Author = strings.TrimSpace(sel.Text())
		}
	})

	doc.Find("meta[itemprop='description']").Each(func(i int, meta *goquery.Selection) {
		content, exists := meta.Attr("content")
		if exists && content != "" && novel.Description == "" {
			novel.Description = content
		}
	})

	doc.Find("meta[property='og:image']").Each(func(i int, meta *goquery.Selection) {
		content, exists := meta.Attr("content")
		if exists && content != "" && novel.CoverURL == "" {
			novel.CoverURL = content
		}
	})

	if novel.CoverURL == "" {
		doc.Find("figure.cover img").Each(func(i int, img *goquery.Selection) {
			src := img.AttrOr("src", "")
			if src != "" && novel.CoverURL == "" {
				if strings.HasPrefix(src, "//") {
					src = "https:" + src
				}
				if strings.HasPrefix(src, "/") {
					src = "https://novelfire.net" + src
				}
				novel.CoverURL = src
			}
		})
	}

	doc.Find("meta[itemprop='keywords']").Each(func(i int, meta *goquery.Selection) {
		content, exists := meta.Attr("content")
		if exists && content != "" {
			parts := strings.Split(content, ",")
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p != "" && !strings.EqualFold(p, "Novel") && !strings.EqualFold(p, "Webnovel") {
					novel.Genres = append(novel.Genres, p)
				}
			}
		}
	})

	if len(novel.Genres) == 0 {
		doc.Find(".categories ul li a").Each(func(i int, a *goquery.Selection) {
			g := strings.TrimSpace(a.Text())
			if g != "" {
				novel.Genres = append(novel.Genres, g)
			}
		})
	}

	statusEl := doc.Find(".header-stats").First().Text()
	lower := strings.ToLower(statusEl)
	switch {
	case strings.Contains(lower, "ongoing"):
		novel.Status = "ongoing"
	case strings.Contains(lower, "completed"):
		novel.Status = "completed"
	case strings.Contains(lower, "hiatus"):
		novel.Status = "hiatus"
	default:
		novel.Status = "ongoing"
	}

	if novel.Title == "" {
		return nil, fmt.Errorf("could not extract novel info from %s", urlStr)
	}

	return novel, nil
}

func (s *Scraper) ScrapeNovelfireChapterContent(slug string, chapterNum int) (string, error) {
	url := fmt.Sprintf("https://novelfire.net/book/%s/chapter-%d", slug, chapterNum)
	doc, err := s.fetchDoc(url)
	if err != nil {
		return "", err
	}

	content := ""
	contentDiv := doc.Find("#content")
	if contentDiv.Length() > 0 {
		contentDiv.Find("script, iframe, ins, .nf-ads, .report-container, .hidden").Remove()
		content, err = contentDiv.Html()
		if err != nil {
			return "", err
		}
		content = strings.TrimSpace(content)
	}

	if content == "" {
		container := doc.Find("#chapter-container")
		if container.Length() > 0 {
			container.Find("script, iframe, ins, .nf-ads, .report-container, .hidden, #restore-scroll-btn, .text-center.pb-1").Remove()
			content, err = container.Html()
			if err != nil {
				return "", err
			}
			content = strings.TrimSpace(content)
		}
	}

	return content, nil
}

func (s *Scraper) ScrapeNovelfireChapters(slug string) ([]NovelfireChapterItem, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	var chapters []NovelfireChapterItem
	page := 1

	for {
		url := fmt.Sprintf("https://novelfire.net/book/%s/chapters?page=%d", slug, page)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36")
		req.Header.Set("Accept", "text/html,application/xhtml+xml")
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
		req.Header.Set("Referer", "https://novelfire.net/book/"+slug)

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("HTTP %d on page %d", resp.StatusCode, page)
		}

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		var count int
		doc.Find("span.chapter-no").Each(func(i int, sel *goquery.Selection) {
			numStr := strings.TrimSpace(sel.Text())
			num, err := strconv.Atoi(numStr)
			if err != nil {
				return
			}
			title := ""
			sel.Parent().Find("strong.chapter-title").Each(func(j int, titleSel *goquery.Selection) {
				title = strings.TrimSpace(titleSel.Text())
			})
			if title == "" {
				title = sel.Parent().AttrOr("title", "")
			}
			if title == "" {
				title = "Chapter " + numStr
			}

			chapters = append(chapters, NovelfireChapterItem{
				Number: num,
				Title:  title,
			})
			count++
		})

		if count == 0 {
			break
		}

		page++
		time.Sleep(200 * time.Millisecond)
	}

	return chapters, nil
}
