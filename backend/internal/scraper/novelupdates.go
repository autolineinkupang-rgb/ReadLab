package scraper

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func (s *Scraper) scrapeNovelUpdates(urlStr string) (*ScrapedNovel, error) {
	doc, err := s.fetchDoc(urlStr)
	if err != nil {
		return nil, err
	}

	check := strings.TrimSpace(doc.Find("title").Text())
	if strings.Contains(check, "Just a moment") || strings.Contains(check, "Cloudflare") {
		return nil, fmt.Errorf("blocked by Cloudflare - NovelUpdates uses anti-scraping protection")
	}

	novel := &ScrapedNovel{
		SourceURL: urlStr,
	}

	doc.Find("div.seriestitle, h1.series-title, h1.entry-title").Each(func(i int, sel *goquery.Selection) {
		if novel.Title == "" {
			novel.Title = strings.TrimSpace(sel.Text())
		}
	})

	if novel.Title == "" {
		doc.Find("h1").Each(func(i int, h1 *goquery.Selection) {
			text := strings.TrimSpace(h1.Text())
			if text != "" && novel.Title == "" && len(text) > 3 {
				novel.Title = text
			}
		})
	}

	doc.Find("div#shoutmix, div.shoutmix").Remove()

	doc.Find("div#editdescription, div.description, div.wpb_text_column").Each(func(i int, sel *goquery.Selection) {
		html, err := sel.Html()
		if err == nil {
			text := stripHTML(html)
			if len(text) > 100 && novel.Description == "" {
				novel.Description = text
			}
		}
	})

	if novel.Description == "" {
		doc.Find("div[itemprop='description'], .series-description, .desc").Each(func(i int, sel *goquery.Selection) {
			text := strings.TrimSpace(sel.Text())
			if len(text) > 80 && novel.Description == "" {
				novel.Description = text
			}
		})
	}

	doc.Find("div.seriestru, .series-info, table.series-info").Each(func(i int, sel *goquery.Selection) {
		sel.Find("tr").Each(func(j int, tr *goquery.Selection) {
			label := strings.TrimSpace(tr.Find("td:first-child, th").Text())
			value := strings.TrimSpace(tr.Find("td:last-child").Text())
			lower := strings.ToLower(label)

			switch {
			case strings.Contains(lower, "author"):
				novel.Author = value
			case strings.Contains(lower, "status"):
				s := strings.ToLower(value)
				switch {
				case strings.Contains(s, "ongoing"):
					novel.Status = "ongoing"
				case strings.Contains(s, "completed"):
					novel.Status = "completed"
				case strings.Contains(s, "hiatus"):
					novel.Status = "hiatus"
				case strings.Contains(s, "dropped"):
					novel.Status = "dropped"
				default:
					novel.Status = value
				}
			case strings.Contains(lower, "genre") || strings.Contains(lower, "tags"):
				parts := strings.Split(value, ",")
				for _, g := range parts {
					g = strings.TrimSpace(g)
					if g != "" {
						novel.Genres = append(novel.Genres, g)
					}
				}
			}
		})
	})

	doc.Find("a[href*='genre'], a[href*='genre'], .genre a, .tags a, .series-genre a, div.genre a").Each(func(i int, a *goquery.Selection) {
		g := strings.TrimSpace(a.Text())
		if g != "" {
			seen := false
			for _, existing := range novel.Genres {
				if strings.EqualFold(existing, g) {
					seen = true
					break
				}
			}
			if !seen {
				novel.Genres = append(novel.Genres, g)
			}
		}
	})

	doc.Find("img[src*='novelupdates'], .series-cover img, img.wp-post-image, img[alt*='cover']").Each(func(i int, img *goquery.Selection) {
		src, exists := img.Attr("src")
		if exists && src != "" && novel.CoverURL == "" {
			if strings.HasPrefix(src, "//") {
				src = "https:" + src
			}
			novel.CoverURL = src
		}
	})

	doc.Find("div#myTable table tbody tr, table#myTable tr, .chapter-list tr, table.chapter-table tr, .chapters tr").Each(func(i int, tr *goquery.Selection) {
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
			u, err := url.Parse(urlStr)
			if err == nil {
				chURL = u.Scheme + "://" + u.Host + href
			}
		}

		novel.Chapters = append(novel.Chapters, ScrapedChapter{
			Number: len(novel.Chapters) + 1,
			Title:  chTitle,
			URL:    chURL,
		})
	})

	if novel.Chapters == nil {
		doc.Find("div.chp_list a, .chapter-list a, ul.chapter-list li a, .chp-list a").Each(func(i int, a *goquery.Selection) {
			href, exists := a.Attr("href")
			if !exists || href == "" {
				return
			}
			title := strings.TrimSpace(a.Text())
			chURL := href
			if strings.HasPrefix(href, "/") {
				u, err := url.Parse(urlStr)
				if err == nil {
					chURL = u.Scheme + "://" + u.Host + href
				}
			}
			novel.Chapters = append(novel.Chapters, ScrapedChapter{
				Number: len(novel.Chapters) + 1,
				Title:  title,
				URL:    chURL,
			})
		})
	}

	if novel.Title == "" {
		return nil, fmt.Errorf("could not extract novel info - Cloudflare may have blocked the request")
	}

	return novel, nil
}
