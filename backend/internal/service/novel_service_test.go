package service

import (
	"fmt"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"readlab/backend/internal/model"
)

func setupNovelServiceTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	db.AutoMigrate(
		&model.Novel{}, &model.Genre{}, &model.NovelGenre{},
		&model.Chapter{},
	)
	return db
}

func seedNovels(t *testing.T, db *gorm.DB, count int) []model.Novel {
	t.Helper()
	novels := make([]model.Novel, count)
	for i := 0; i < count; i++ {
		n := model.Novel{
			Title:  fmt.Sprintf("Novel %d", i+1),
			Slug:   fmt.Sprintf("novel-%d", i+1),
			Author: fmt.Sprintf("Author %d", i+1),
			Status: "ongoing",
			Views:  uint64(i * 100),
			Rating: float64(i % 5),
		}
		if err := db.Create(&n).Error; err != nil {
			t.Fatalf("failed to seed novel %d: %v", i+1, err)
		}
		novels[i] = n
	}
	return novels
}

func newNovelService(t *testing.T) (*NovelService, *gorm.DB) {
	db := setupNovelServiceTestDB(t)
	return NewNovelService(db), db
}

// ── List ──

func TestNovelList_EmptyDatabase(t *testing.T) {
	svc, _ := newNovelService(t)

	page, err := svc.List(NovelFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Total != 0 {
		t.Errorf("expected total 0, got %d", page.Total)
	}
	if len(page.Data) != 0 {
		t.Errorf("expected 0 novels, got %d", len(page.Data))
	}
	if page.TotalPages != 0 {
		t.Errorf("expected 0 total pages, got %d", page.TotalPages)
	}
}

func TestNovelList_Pagination(t *testing.T) {
	svc, _ := newNovelService(t)
	seedNovels(t, svc.DB, 5)

	page, err := svc.List(NovelFilter{Page: 1, Limit: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Total != 5 {
		t.Errorf("expected total 5, got %d", page.Total)
	}
	if len(page.Data) != 2 {
		t.Fatalf("expected 2 novels, got %d", len(page.Data))
	}
	if page.TotalPages != 3 { // ceil(5/2) = 3
		t.Errorf("expected 3 total pages, got %d", page.TotalPages)
	}
	if page.Limit != 2 {
		t.Errorf("expected limit 2, got %d", page.Limit)
	}
}

func TestNovelList_DefaultPagination(t *testing.T) {
	svc, _ := newNovelService(t)
	seedNovels(t, svc.DB, 3)

	// Page 0 should be clamped to 1
	page, err := svc.List(NovelFilter{Page: 0, Limit: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Page != 1 {
		t.Errorf("expected page 1, got %d", page.Page)
	}
	if page.Limit != 20 { // default
		t.Errorf("expected limit 20, got %d", page.Limit)
	}
}

func TestNovelList_SortAndOrder(t *testing.T) {
	svc, _ := newNovelService(t)
	seedNovels(t, svc.DB, 5)

	// Test valid sort
	page, err := svc.List(NovelFilter{Sort: "views", Order: "asc"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(page.Data) < 2 {
		t.Fatal("need at least 2 novels to check sort order")
	}
	if page.Data[0].Views >= page.Data[1].Views {
		t.Errorf("expected ascending views, got %d then %d", page.Data[0].Views, page.Data[1].Views)
	}

	// Test invalid sort falls back to created_at
	page2, _ := svc.List(NovelFilter{Sort: "invalid_column", Order: "desc"})
	if len(page2.Data) == 0 {
		t.Error("expected novels even with invalid sort")
	}

	// Test invalid order falls back to desc
	page3, _ := svc.List(NovelFilter{Sort: "views", Order: "invalid"})
	if len(page3.Data) == 0 {
		t.Error("expected novels even with invalid order")
	}
}

func TestNovelList_StatusFilter(t *testing.T) {
	svc, db := newNovelService(t)
	db.Create(&model.Novel{Title: "Ongoing A", Slug: "ongoing-a", Status: "ongoing"})
	db.Create(&model.Novel{Title: "Completed B", Slug: "completed-b", Status: "completed"})
	db.Create(&model.Novel{Title: "Ongoing C", Slug: "ongoing-c", Status: "ongoing"})

	page, err := svc.List(NovelFilter{Status: "ongoing"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Total != 2 {
		t.Errorf("expected 2 ongoing novels, got %d", page.Total)
	}

	page2, _ := svc.List(NovelFilter{Status: "completed"})
	if page2.Total != 1 {
		t.Errorf("expected 1 completed novel, got %d", page2.Total)
	}
}

func TestNovelList_SearchFilter(t *testing.T) {
	svc, db := newNovelService(t)
	db.Create(&model.Novel{Title: "Dragon Slayer", Slug: "dragon-slayer", Author: "John"})
	db.Create(&model.Novel{Title: "Wolf Warrior", Slug: "wolf-warrior", Author: "Jane"})
	db.Create(&model.Novel{Title: "Dragon Tamer", Slug: "dragon-tamer", Author: "Bob"})

	page, err := svc.List(NovelFilter{Search: "dragon"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Total != 2 {
		t.Errorf("expected 2 novels matching 'dragon', got %d", page.Total)
	}

	page2, _ := svc.List(NovelFilter{Search: "Jane"})
	if page2.Total != 1 {
		t.Errorf("expected 1 novel matching author 'Jane', got %d", page2.Total)
	}
}

// ── Trending ──

func TestTrending_RespectsLimit(t *testing.T) {
	svc, _ := newNovelService(t)
	seedNovels(t, svc.DB, 10)

	novels, err := svc.Trending(3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(novels) != 3 {
		t.Fatalf("expected 3 novels, got %d", len(novels))
	}

	// Should be ordered by views DESC
	for i := 1; i < len(novels); i++ {
		if novels[i-1].Views < novels[i].Views {
			t.Errorf("trending not ordered by views DESC: %d < %d at index %d",
				novels[i-1].Views, novels[i].Views, i)
		}
	}
}

func TestTrending_ClampsLimit(t *testing.T) {
	svc, _ := newNovelService(t)
	seedNovels(t, svc.DB, 5)

	// Limit > 100 should be clamped to 20
	novels, err := svc.Trending(200)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(novels) != 5 {
		t.Errorf("expected 5 novels (all), got %d", len(novels))
	}
}

func TestTrending_EmptyDatabase(t *testing.T) {
	svc, _ := newNovelService(t)

	novels, err := svc.Trending(10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(novels) != 0 {
		t.Errorf("expected 0 novels, got %d", len(novels))
	}
}

// ── Recommendations ──

func TestRecommendations_FiltersByRating(t *testing.T) {
	svc, db := newNovelService(t)

	// Create novels with varying ratings
	db.Create(&model.Novel{Title: "High Rated", Slug: "high-rated", Rating: 4.5, RatingCount: 10})
	db.Create(&model.Novel{Title: "Low Rated", Slug: "low-rated", Rating: 2.0, RatingCount: 5})
	db.Create(&model.Novel{Title: "Zero Rated", Slug: "zero-rated", Rating: 0, RatingCount: 0})
	db.Create(&model.Novel{Title: "Medium Rated", Slug: "medium-rated", Rating: 3.5, RatingCount: 8})

	novels, err := svc.Recommendations(10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should not include zero-rated novel
	for _, n := range novels {
		if n.Rating <= 0 {
			t.Errorf("recommendations should filter rating > 0, got novel '%s' with rating %v", n.Title, n.Rating)
		}
	}
	if len(novels) != 3 {
		t.Errorf("expected 3 novels with rating > 0, got %d", len(novels))
	}

	// Should be ordered by rating DESC
	for i := 1; i < len(novels); i++ {
		if novels[i-1].Rating < novels[i].Rating {
			t.Errorf("recommendations not ordered by rating DESC: %v < %v",
				novels[i-1].Rating, novels[i].Rating)
		}
	}
}

func TestRecommendations_ClampsLimit(t *testing.T) {
	svc, _ := newNovelService(t)
	// Default limit should be 12
	novels, err := svc.Recommendations(200)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(novels) != 0 {
		t.Errorf("expected 0 novels for empty db, got %d", len(novels))
	}
}

// ── GetByID ──

func TestGetByID_Found(t *testing.T) {
	svc, db := newNovelService(t)
	novel := model.Novel{
		Title:  "Test Novel",
		Slug:   "test-novel",
		Author: "Test Author",
	}
	db.Create(&novel)

	found, err := svc.GetByID(novel.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found.Title != "Test Novel" {
		t.Errorf("expected title 'Test Novel', got '%s'", found.Title)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	svc, _ := newNovelService(t)

	_, err := svc.GetByID(99999)
	if err == nil {
		t.Fatal("expected error for missing novel")
	}
}

// ── Search ──

func TestSearch_MatchingNovels(t *testing.T) {
	svc, db := newNovelService(t)
	db.Create(&model.Novel{Title: "Dragon Quest", Slug: "dragon-quest", Author: "Author A", Views: 1000})
	db.Create(&model.Novel{Title: "Dragon Soul", Slug: "dragon-soul", Author: "Author B", Views: 500})
	db.Create(&model.Novel{Title: "Phoenix Rise", Slug: "phoenix-rise", Author: "Author C", Views: 200})

	novels, total, err := svc.Search("dragon", 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 2 {
		t.Errorf("expected 2 results, got %d", total)
	}
	if len(novels) != 2 {
		t.Fatalf("expected 2 novels, got %d", len(novels))
	}

	// Should be ordered by views DESC
	if novels[0].Title != "Dragon Quest" {
		t.Errorf("expected 'Dragon Quest' first (higher views), got '%s'", novels[0].Title)
	}
}

func TestSearch_NoResults(t *testing.T) {
	svc, _ := newNovelService(t)

	novels, total, err := svc.Search("nonexistent", 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 0 {
		t.Errorf("expected 0 results, got %d", total)
	}
	if len(novels) != 0 {
		t.Errorf("expected 0 novels, got %d", len(novels))
	}
}

func TestSearch_Pagination(t *testing.T) {
	svc, db := newNovelService(t)
	for i := 0; i < 5; i++ {
		db.Create(&model.Novel{
			Title: fmt.Sprintf("Match %d", i),
			Slug:  fmt.Sprintf("match-%d", i),
			Views: uint64(i * 10),
		})
	}

	// Page 1, limit 2
	novels, total, err := svc.Search("Match", 1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}
	if len(novels) != 2 {
		t.Fatalf("expected 2 novels, got %d", len(novels))
	}
}

// ── Autocomplete ──

func TestAutocomplete_LimitedResults(t *testing.T) {
	svc, db := newNovelService(t)
	for i := 0; i < 10; i++ {
		db.Create(&model.Novel{
			Title: fmt.Sprintf("Fantasy World %d", i),
			Slug:  fmt.Sprintf("fantasy-world-%d", i),
			Views: uint64(i * 10),
		})
	}

	results, err := svc.Autocomplete("Fantasy", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	// Results should have ID, Slug, Title populated
	for _, r := range results {
		if r.Title == "" {
			t.Error("expected non-empty Title")
		}
		if r.Slug == "" {
			t.Error("expected non-empty Slug")
		}
	}
}

func TestAutocomplete_DefaultLimit(t *testing.T) {
	svc, db := newNovelService(t)
	for i := 0; i < 10; i++ {
		db.Create(&model.Novel{
			Title: fmt.Sprintf("SciFi Story %d", i),
			Slug:  fmt.Sprintf("scifi-story-%d", i),
			Views: uint64(i * 10),
		})
	}

	// Limit > 20 should be clamped to 5
	results, err := svc.Autocomplete("SciFi", 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 5 {
		t.Fatalf("expected 5 results (default limit), got %d", len(results))
	}
}

func TestAutocomplete_NoMatch(t *testing.T) {
	svc, _ := newNovelService(t)

	results, err := svc.Autocomplete("nonexistent", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestAutocomplete_OnlyMatchesTitle(t *testing.T) {
	svc, db := newNovelService(t)
	db.Create(&model.Novel{Title: "Dragon Fire", Slug: "dragon-fire", Author: "John Author"})
	db.Create(&model.Novel{Title: "Unique Title", Slug: "unique-title", Author: "Dragon Writer"})

	// Autocomplete only searches title, not author
	results, err := svc.Autocomplete("Dragon", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result (title match only), got %d", len(results))
	}
	if results[0].Title != "Dragon Fire" {
		t.Errorf("expected 'Dragon Fire', got '%s'", results[0].Title)
	}
}