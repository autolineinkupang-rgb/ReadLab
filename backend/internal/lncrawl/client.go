package lncrawl

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/klauspost/compress/zstd"
)

type chapterJSON struct {
	Title       string `json:"title"`
	Serial      int    `json:"serial"`
	VolumeID    string `json:"volume_id"`
	NovelID     string `json:"novel_id"`
	URL         string `json:"url"`
	Body        string `json:"body"`
	IsAvailable bool   `json:"is_available"`
	IsDone      bool   `json:"is_done"`
	ContentFile string `json:"content_file"`
}

type ChapterContent struct {
	Number  int
	Title   string
	Content string
}

type Result struct {
	Title    string
	Author   string
	URL      string
	CoverURL string
	Chapters []ChapterContent
	Total    int
}

func RunCrawl(novelURL string, maxChapters int) (*Result, error) {
	venvPython := "/tmp/scraper-venv/bin/python3"
	lncrawlPath := "/tmp/scraper-venv/bin/lncrawl"

	if _, err := os.Stat(lncrawlPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("lncrawl not found at %s", lncrawlPath)
	}

	tmpDir, err := os.MkdirTemp("", "lncrawl-*")
	if err != nil {
		return nil, fmt.Errorf("cannot create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	args := []string{lncrawlPath, "crawl", novelURL, "--noin", "--format", "json"}
	if maxChapters > 0 {
		args = append(args, "--first", fmt.Sprintf("%d", maxChapters))
	}

	cmd := exec.Command(venvPython, args...)
	cmd.Dir = tmpDir
	cmd.Env = append(os.Environ(),
		"PATH=/tmp/scraper-venv/bin:"+os.Getenv("PATH"),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("lncrawl failed: %w\nOutput: %s", err, string(output))
	}

	stdout := string(output)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot get home dir: %w", err)
	}

	result, err := parseOutput(stdout, homeDir)
	if err != nil {
		return nil, fmt.Errorf("parse failed: %w", err)
	}

	return result, nil
}

func parseOutput(stdout, homeDir string) (*Result, error) {
	meta := extractMetaFromStdout(stdout)
	zipPath := findZipPath(stdout)

	if zipPath == "" {
		return nil, fmt.Errorf("cannot find zip output in lncrawl output")
	}
	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("zip not found at %s", zipPath)
	}

	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("cannot open zip: %w", err)
	}
	defer reader.Close()

	lncrawlDir := filepath.Join(homeDir, ".lncrawl")

	var chapters []ChapterContent
	var allSerial []int
	chMap := make(map[int]*chapterJSON)
	novelID := ""

	for _, f := range reader.File {
		if f.FileInfo().IsDir() || !strings.HasSuffix(f.Name, ".json") {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			continue
		}
		var ch chapterJSON
		if err := json.NewDecoder(rc).Decode(&ch); err != nil {
			rc.Close()
			continue
		}
		rc.Close()

		if ch.NovelID != "" {
			novelID = ch.NovelID
		}
		chMap[ch.Serial] = &ch
		allSerial = append(allSerial, ch.Serial)
	}

	sort.Ints(allSerial)

	for _, serial := range allSerial {
		ch := chMap[serial]
		if !ch.IsAvailable || !ch.IsDone || ch.ContentFile == "" {
			continue
		}

		contentPath := filepath.Join(lncrawlDir, ch.ContentFile)
		content, err := decompressZST(contentPath)
		if err != nil {
			continue
		}

		chapters = append(chapters, ChapterContent{
			Number:  ch.Serial,
			Title:   ch.Title,
			Content: content,
		})
	}

	coverURL := ""
	if novelID != "" {
		coverPath := filepath.Join(lncrawlDir, "novels", novelID, "cover.jpg")
		if _, err := os.Stat(coverPath); err == nil {
			coverURL = "/api/covers/" + novelID + "/cover.jpg"
		}
	}

	return &Result{
		Title:    meta.Title,
		Author:   meta.Author,
		URL:      meta.URL,
		CoverURL: coverURL,
		Chapters: chapters,
		Total:    len(reader.File),
	}, nil
}

type stdoutMeta struct {
	Title  string
	Author string
	URL    string
}

func extractMetaFromStdout(output string) stdoutMeta {
	var m stdoutMeta
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "│") {
			continue
		}
		content := strings.TrimSpace(trimmed)
		content = strings.TrimPrefix(content, "│")
		content = strings.TrimSuffix(content, "│")
		content = strings.TrimSpace(content)

		if content == "" || strings.Contains(content, "volumes") {
			continue
		}

		if strings.HasPrefix(content, "http") {
			if m.URL == "" {
				m.URL = content
			}
		} else if m.Title == "" {
			m.Title = content
		} else if m.Author == "" {
			m.Author = content
		}
	}
	return m
}

func findZipPath(output string) string {
	lines := strings.Split(output, "\n")
	var buf strings.Builder
	collecting := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasSuffix(trimmed, ".json.zip") {
			if collecting {
				buf.WriteString(trimmed)
			} else {
				buf.Reset()
				buf.WriteString(trimmed)
			}
			return buf.String()
		}
		if collecting {
			buf.WriteString(trimmed)
			continue
		}
		if strings.Contains(trimmed, ".json.zip") {
			buf.Reset()
			collecting = true
			buf.WriteString(trimmed)
			trimmed2 := strings.TrimSuffix(trimmed, ".json.zip")
			if trimmed2 == trimmed {
				continue
			}
			return buf.String()
		}
		if strings.HasPrefix(trimmed, "/") && strings.Count(trimmed, "/") >= 3 {
			buf.Reset()
			collecting = true
			buf.WriteString(trimmed)
		}
	}

	result := buf.String()
	if strings.Contains(result, ".json.zip") {
		return result
	}
	return ""
}

func decompressZST(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	decoder, err := zstd.NewReader(nil)
	if err != nil {
		return "", err
	}
	defer decoder.Close()
	decompressed, err := decoder.DecodeAll(data, nil)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(decompressed)), nil
}
