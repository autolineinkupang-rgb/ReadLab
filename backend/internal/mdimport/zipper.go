package mdimport

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

const (
	maxZipEntrySize  = 5 << 20 // 5 MB per entry
	maxTotalDecompressed = 50 << 20 // 50 MB total
)

func ExtractMDsFromZip(data []byte) (map[string]string, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}

	files := make(map[string]string)
	var names []string
	var totalDecompressed int64

	for _, f := range reader.File {
		if f.FileInfo().IsDir() {
			continue
		}
		if strings.HasPrefix(filepath.Base(f.Name), "._") {
			continue
		}
		if strings.Contains(f.Name, "__MACOSX") {
			continue
		}
		ext := strings.ToLower(filepath.Ext(f.Name))
		if ext != ".md" && ext != ".markdown" {
			continue
		}

		if f.FileInfo().Size() > maxZipEntrySize {
			return nil, fmt.Errorf("entry %s exceeds max size of 5 MB", f.Name)
		}

		rc, err := f.Open()
		if err != nil {
			continue
		}
		data, err := io.ReadAll(io.LimitReader(rc, maxZipEntrySize))
		rc.Close()
		if err != nil {
			continue
		}

		totalDecompressed += int64(len(data))
		if totalDecompressed > maxTotalDecompressed {
			return nil, fmt.Errorf("total decompressed size exceeds max of 50 MB")
		}

		files[f.Name] = string(data)
		names = append(names, f.Name)
	}

	if len(files) == 0 {
		return nil, nil
	}

	SortFilenames(names)
	sorted := make(map[string]string, len(files))
	for _, name := range names {
		sorted[name] = files[name]
	}

	return sorted, nil
}
