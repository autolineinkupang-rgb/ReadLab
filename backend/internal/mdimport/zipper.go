package mdimport

import (
	"archive/zip"
	"bytes"
	"io"
	"path/filepath"
	"strings"
)

func ExtractMDsFromZip(data []byte) (map[string]string, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}

	files := make(map[string]string)
	var names []string

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

		rc, err := f.Open()
		if err != nil {
			continue
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			continue
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
