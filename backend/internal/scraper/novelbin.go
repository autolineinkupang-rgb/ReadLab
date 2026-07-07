package scraper

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func (s *Scraper) scrapeNovelBin(url string) (*ScrapedNovel, error) {
	doc, err := s.fetchDoc(url)
	if err != nil {
		return nil, err
	}

	novel := &ScrapedNovel{
		SourceURL: url,
	}

	novel.Title = strings.TrimSpace(doc.Find("h3.title").First().Text())
	if novel.Title == "" {
		novel.Title = strings.TrimSpace(doc.Find("h1").First().Text())
	}

	doc.Find("div.info-col, .col-info, .info").Each(func(i int, s *goquery.Selection) {
		s.Find("p").Each(func(j int, p *goquery.Selection) {
			text := strings.TrimSpace(p.Text())
			lower := strings.ToLower(text)
			switch {
			case strings.Contains(lower, "author"):
				parts := strings.SplitN(text, ":", 2)
				if len(parts) == 2 {
					novel.Author = strings.TrimSpace(parts[1])
				}
			case strings.Contains(lower, "status"):
				parts := strings.SplitN(text, ":", 2)
				if len(parts) == 2 {
					s := strings.TrimSpace(parts[1])
					switch strings.ToLower(s) {
					case "completed":
						novel.Status = "completed"
					case "ongoing":
						novel.Status = "ongoing"
					default:
						novel.Status = s
					}
				}
			case strings.Contains(lower, "alternative"):
				parts := strings.SplitN(text, ":", 2)
				if len(parts) == 2 {
					novel.AltTitle = strings.TrimSpace(parts[1])
				}
			}
		})
	})

	novel.Description = strings.TrimSpace(doc.Find("div.desc-text, .desc, .description, #description").First().Text())

	doc.Find("img.cover, .book img, .novel-cover img, img[alt*='cover']").Each(func(i int, img *goquery.Selection) {
		src, exists := img.Attr("src")
		if exists && src != "" && novel.CoverURL == "" {
			novel.CoverURL = src
		}
	})

	doc.Find("a[href*='genre'], a[href*='category'], .tags a, .genres a").Each(func(i int, a *goquery.Selection) {
		g := strings.TrimSpace(a.Text())
		if g != "" {
			novel.Genres = append(novel.Genres, g)
		}
	})

	doc.Find("#chapters a, .chapter-list a, .list-chapter a, ul.chapter-list li a").Each(func(i int, a *goquery.Selection) {
		href, exists := a.Attr("href")
		if !exists || href == "" {
			return
		}
		chURL := href
		if strings.HasPrefix(href, "/") {
			base := url
			if idx := strings.Index(base, "//"); idx >= 0 {
				if idx2 := strings.Index(base[idx+2:], "/"); idx2 >= 0 {
					base = base[:idx+2+idx2]
				}
			}
			chURL = base + href
		}
		title := strings.TrimSpace(a.Text())
		novel.Chapters = append(novel.Chapters, ScrapedChapter{
			Number: len(novel.Chapters) + 1,
			Title:  title,
			URL:    chURL,
		})
	})

	if novel.Title == "" {
		return nil, fmt.Errorf("could not extract novel info from %s", url)
	}

	return novel, nil
}

func (s *Scraper) scrapeNovelBinChapter(url string) (*ScrapedChapter, error) {
	doc, err := s.fetchDoc(url)
	if err != nil {
		return nil, err
	}

	content := ""
	doc.Find("#chapter-content, .chapter-content, #content, .content, .reading-content, .chapter-c") .Each(func(i int, sel *goquery.Selection) {
		html, err := sel.Html()
		if err == nil {
			content = strings.TrimSpace(html)
		}
	})

	if content == "" {
		doc.Find("p").Each(func(i int, p *goquery.Selection) {
			text := strings.TrimSpace(p.Text())
			if len(text) > 50 {
				content += text + "\n\n"
			}
		})
	}

	title := strings.TrimSpace(doc.Find("h1, h2, .chapter-title, .title").First().Text())

	num := 0
	parts := strings.Split(title, " ")
	for _, p := range parts {
		if n, err := strconv.Atoi(strings.TrimSuffix(p, ":")); err == nil {
			num = n
			break
		}
	}

	return &ScrapedChapter{
		Number:  num,
		Title:   title,
		URL:     url,
		Content: content,
	}, nil
}

func (s *Scraper) scrapeFreeWebNovel(url string) (*ScrapedNovel, error) {
	doc, err := s.fetchDoc(url)
	if err != nil {
		return nil, err
	}

	novel := &ScrapedNovel{
		SourceURL: url,
	}

	novel.Title = strings.TrimSpace(doc.Find("h1, .title, .novel-title").First().Text())

	doc.Find(".info, .novel-info, .detail").Each(func(i int, sel *goquery.Selection) {
		sel.Find("p, span, div").Each(func(j int, p *goquery.Selection) {
			text := strings.TrimSpace(p.Text())
			lower := strings.ToLower(text)
			if strings.HasPrefix(lower, "author") {
				parts := strings.SplitN(text, ":", 2)
				if len(parts) == 2 {
					novel.Author = strings.TrimSpace(parts[1])
				}
			}
			if strings.HasPrefix(lower, "status") {
				parts := strings.SplitN(text, ":", 2)
				if len(parts) == 2 {
					novel.Status = strings.TrimSpace(parts[1])
				}
			}
		})
	})

	novel.Description = strings.TrimSpace(doc.Find(".desc, .description, #description").First().Text())

	doc.Find("img.cover, img[alt*='cover'], .cover img").Each(func(i int, img *goquery.Selection) {
		src, exists := img.Attr("src")
		if exists && src != "" && novel.CoverURL == "" {
			novel.CoverURL = src
		}
	})

	doc.Find("a[href*='genre'], a[href*='category'], .tags a").Each(func(i int, a *goquery.Selection) {
		g := strings.TrimSpace(a.Text())
		if g != "" {
			novel.Genres = append(novel.Genres, g)
		}
	})

	doc.Find(".chapter-list a, #chapters a, .list-chapter a").Each(func(i int, a *goquery.Selection) {
		href, exists := a.Attr("href")
		if !exists || href == "" {
			return
		}
		chURL := href
		if strings.HasPrefix(href, "/") {
			base := url
			if idx := strings.Index(base, "//"); idx >= 0 {
				if idx2 := strings.Index(base[idx+2:], "/"); idx2 >= 0 {
					base = base[:idx+2+idx2]
				}
			}
			chURL = base + href
		}
		novel.Chapters = append(novel.Chapters, ScrapedChapter{
			Number: len(novel.Chapters) + 1,
			Title:  strings.TrimSpace(a.Text()),
			URL:    chURL,
		})
	})

	if novel.Title == "" {
		return nil, fmt.Errorf("could not extract novel info from %s", url)
	}

	return novel, nil
}

func (s *Scraper) scrapeFreeWebNovelChapter(url string) (*ScrapedChapter, error) {
	doc, err := s.fetchDoc(url)
	if err != nil {
		return nil, err
	}

	content := ""
	doc.Find("#chapter-content, .chapter-content, #content, .txt, .content").Each(func(i int, sel *goquery.Selection) {
		html, err := sel.Html()
		if err == nil {
			content = strings.TrimSpace(html)
		}
	})

	if content == "" {
		doc.Find("p").Each(func(i int, p *goquery.Selection) {
			text := strings.TrimSpace(p.Text())
			if len(text) > 50 {
				content += text + "\n\n"
			}
		})
	}

	title := strings.TrimSpace(doc.Find("h1, h2, .chapter-title").First().Text())

	return &ScrapedChapter{
		Title:   title,
		URL:     url,
		Content: content,
	}, nil
}
