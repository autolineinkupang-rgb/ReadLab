package mdimport

import (
	"sort"
	"strings"
)

type ParsedChapter struct {
	Number      int    `json:"number"`
	Title       string `json:"title"`
	ContentMD   string `json:"content_md"`
	ContentHTML string `json:"content_html"`
	Exists      bool   `json:"exists"`
}

type ParseResult struct {
	NovelTitle string          `json:"novel_title,omitempty"`
	Chapters   []ParsedChapter `json:"chapters"`
	Warnings   []string        `json:"warnings"`
}

// ParseSingleMD parses a single .md file containing multiple chapters.
// Chapters are delimited by ## or # headings.
// The first heading before any chapter heading is treated as novel title.
func ParseSingleMD(input string) *ParseResult {
	result := &ParseResult{}
	input = CleanMarkdown(input)
	lines := strings.Split(input, "\n")

	var chapters []ParsedChapter
	var currentTitle string
	var currentBody []string
	var novelTitleSet bool

	flushChapter := func() {
		if currentTitle != "" || len(currentBody) > 0 {
			md := strings.TrimSpace(strings.Join(currentBody, "\n"))
			html := ToHTML(md)
			chapters = append(chapters, ParsedChapter{
				Number:      len(chapters) + 1,
				Title:       currentTitle,
				ContentMD:   md,
				ContentHTML: html,
			})
			currentTitle = ""
			currentBody = nil
		}
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") && !novelTitleSet {
			result.NovelTitle = strings.TrimPrefix(trimmed, "# ")
			novelTitleSet = true
			continue
		}
		if strings.HasPrefix(trimmed, "## ") {
			flushChapter()
			currentTitle = strings.TrimPrefix(trimmed, "## ")
			continue
		}
		if strings.HasPrefix(trimmed, "# ") && novelTitleSet {
			flushChapter()
			currentTitle = strings.TrimPrefix(trimmed, "# ")
			continue
		}
		currentBody = append(currentBody, line)
	}
	flushChapter()

	for i := range chapters {
		chapters[i].Number = i + 1
	}

	result.Chapters = chapters
	return result
}

// ParseChapterMD parses a single chapter from a single .md file (no heading splitting).
func ParseChapterMD(input, defaultTitle string) *ParsedChapter {
	input = CleanMarkdown(input)
	html := ToHTML(input)
	return &ParsedChapter{
		Number:      0,
		Title:       defaultTitle,
		ContentMD:   input,
		ContentHTML: html,
	}
}

// naturalLess compares two strings for natural sort order.
func naturalLess(a, b string) bool {
	return strings.Compare(a, b) < 0
}

// SortFilenames sorts filenames naturally.
func SortFilenames(names []string) {
	sort.Slice(names, func(i, j int) bool {
		return naturalLess(names[i], names[j])
	})
}
