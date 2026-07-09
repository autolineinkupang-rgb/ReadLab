# Markdown Chapter Import — Design Spec

**Date:** 2026-07-09
**Project:** ReadLab
**Status:** Draft

---

## 1. Overview

Writer/Admin dapat mengupload file `.md` (single file berisi banyak chapter) atau `.zip` (berisi banyak `.md` file) untuk di-import sebagai章节 chapter ke novel yang sudah ada. Backend melakukan parsing, auto-formatting cleanup, dan conversion markdown → HTML. Hasil parsing ditampilkan sebagai preview, writer mereview lalu menyimpan.

---

## 2. Model Changes

### Chapter — tambah field `ContentMD`

```
ContentMD string `gorm:"type:text" json:"content_md"`
```

Field ini menyimpan markdown mentah dari hasil upload. Field `Content` tetap berisi HTML hasil konversi, digunakan untuk rendering di reader.

Migration: `db.AutoMigrate(&model.Chapter{})` — GORM otomatis add column.

---

## 3. Backend: Markdown Parser Package

`backend/internal/mdimport/`

### Parser

`parser.go` — Parse single .md content → extract chapters:

```
type ParsedChapter struct {
    Number    int
    Title     string
    ContentMD string   // raw markdown
    ContentHTML string // converted to HTML
}

type ParseResult struct {
    Title    string
    Chapters []ParsedChapter
}
```

**Splitting strategy:**
- Single .md: `##` atau `#` headings sebagai delimiter chapter. Heading level pertama di file dianggap sebagai novel title (jika sebelum chapter pertama). Chapter dimulai dari heading level yang sama atau lebih rendah dari heading chapter pertama.
- ZIP: setiap file .md diurutkan berdasarkan nama file (natural sort), masing-masing jadi satu chapter. Nama file (tanpa ekstensi) jadi judul chapter default.

### Auto-Formatting Cleanup

`cleaner.go` — Fungsi-fungsi cleanup:

```
func CleanMarkdown(input string) string
```

Rules:
1. **Heading spacing** — pastikan ada satu blank line sebelum heading (`\n## Title`), hapus extra spacing
2. **List normalization** — `*` → `-` untuk unordered list, pastikan konsisten
3. **Blank lines** — collapse multiple blank lines (>2) jadi maks 2
4. **Trailing whitespace** — hapus trailing spaces per line
5. **HR normalization** — `***`, `---`, `___` → `---`
6. **Encoding** — pastikan UTF-8, handle BOM
7. **Numbering validation** — validasi chapter numbers tidak duplikat, detect gap

### Markdown → HTML Converter

`converter.go` — Convert markdown ke HTML:

Gunakan library Go `github.com/yuin/goldmark` atau `github.com/gomarkdown/markdown` untuk convert. Pilih `goldmark` karena lebih aktif, extensible, dan aman (no raw HTML by default).

Config:
- Enable typographer (smart quotes, ellipsis, etc.)
- Enable task list
- Tabel support
- Strikethrough support
- Raw HTML = tidak diizinkan (dihapus/escape)
- Link = `target="_blank" rel="noopener noreferrer"`

### Upload Handler

`backend/internal/handler/mdimport_handler.go`

**Endpoint:** `POST /admin/novels/:id/chapters/import-md`

**Middleware:** AuthRequired + RequireRole("writer", "admin")
- Writer: hanya bisa upload ke novel miliknya (check `writer_id`)
- Admin: bisa upload ke novel manapun

**Request:** `multipart/form-data`
- Field `file`: file `.md` atau `.zip` (max 50MB)
- Field `mode`: `"preview"` (default) atau `"save"`

**Response (mode=preview):**

```json
{
  "chapters": [
    {
      "number": 1,
      "title": "The Beginning",
      "content_md": "# The Beginning\\n\\n...",
      "content_html": "<h1>The Beginning</h1><p>...</p>",
      "exists": false
    }
  ],
  "warnings": [
    "Chapter 3: duplicate number, will be auto-renumbered"
  ]
}
```

**Response (mode=save):**

```json
{
  "message": "5 chapters imported successfully",
  "imported": 5,
  "skipped": 0
}
```

**Save logic:**
- Auto-number: jika file .md punya numbering sendiri, gunakan itu. Jika ada duplikat dengan chapter existing, skip atau beri warning.
- Jika dari ZIP, numbering berdasarkan urutan file + offset chapter terakhir.
- Jika `number` sudah ada di DB → skip (tidak overwrite).
- Update denormalized `Chapters` count di novel.

### ZIP Handler

`zipper.go` — Extract ZIP:
- Baca semua file `.md` dari ZIP (rekursif)
- Urutkan secara natural (1.md, 2.md, ..., 10.md — bukan lexicographic)
- Abaikan file non-.md dan folder __MACOSX

---

## 4. Frontend

### Lokasi

Tombol "Import .md/ZIP" di halaman `/en/admin/novels/[id]/chapters/page.tsx`, ditaruh di atas tabel chapters, sebelah kanan.

### Flow

1. Klik tombol → modal terbuka
2. Pilih file (.md atau .zip) via `<input type="file">`
3. Upload ke endpoint `/admin/novels/:id/chapters/import-md` dengan `mode=preview`
4. Tampilkan preview table: nomor, judul, warning (jika ada), checkbox untuk memilih chapter mana yang diimport
5. Tombol "Import Selected" → kirim ulang dengan `mode=save` + selected chapters
6. Loading state + success/error message
7. Refresh chapter list

### Component

Modal komponen: `ChapterMdImportModal.tsx`
- File dropzone / picker
- Preview table dengan checkbox per row
- Warning messages
- Progress indicator

### API

```typescript
export const chapters = {
  // ... existing
  importMd: (novelId: number, file: File, mode: "preview" | "save", selected?: number[]) =>
    fetcherFormData<ImportMdResponse>(`/admin/novels/${novelId}/chapters/import-md`, { file, mode, selected }),
};
```

Butuh helper `fetcherFormData` di api.ts yang pakai `FormData` tanpa `Content-Type` header (biar browser set boundary).

---

## 5. Error Handling

| Case | Response |
|---|---|
| No file uploaded | 400 "file required" |
| Invalid file type | 400 "only .md and .zip files are supported" |
| File too large (>50MB) | 413 "file too large" |
| Novel not found | 404 |
| Writer tidak punya akses | 403 |
| ZIP corrupted | 400 "invalid zip file" |
| No valid .md in ZIP | 400 "no markdown files found in zip" |
| All chapters skipped (duplicate) | 200 with warnings |

---

## 6. Testing

- Backend unit test: parser, cleaner, converter (test table with known md → html)
- Backend integration: upload .md → preview → save → verify chapter created
- ZIP dengan berbagai struktur folder
- Edge cases: empty file, no headings, special chars in filename
- Frontend: modal open → upload → preview renders → save → list refreshed
