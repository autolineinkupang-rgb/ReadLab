package service

import (
        "fmt"
        "testing"

        "gorm.io/driver/sqlite"
        "gorm.io/gorm"
        "readlab/backend/internal/model"
)

func setupReviewServiceTestDB(t *testing.T) *gorm.DB {
        db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
        if err != nil {
                t.Fatalf("failed to open test db: %v", err)
        }
        db.AutoMigrate(&model.User{}, &model.Novel{}, &model.Chapter{}, &model.Review{})
        return db
}

func newReviewService(t *testing.T) (*ReviewService, *gorm.DB) {
        db := setupReviewServiceTestDB(t)
        return NewReviewService(db), db
}

func createReviewUser(t *testing.T, db *gorm.DB) model.User {
        t.Helper()
        user := model.User{
                Username:     fmt.Sprintf("revuser%d", len(db.Tables)),
                Email:        fmt.Sprintf("revuser%d@test.com", len(db.Tables)),
                PasswordHash: "$2a$10$abcdefghijklmnopqrstuvwxABCDEFGHIJ",
                DisplayName:  "Review User",
        }
        if err := db.Create(&user).Error; err != nil {
                t.Fatalf("failed to create user: %v", err)
        }
        return user
}

func createReviewNovel(t *testing.T, db *gorm.DB, chapterCount int) model.Novel {
        t.Helper()
        novel := model.Novel{
                Title: "Test Novel",
                Slug:  fmt.Sprintf("test-novel-%d", chapterCount),
        }
        if err := db.Create(&novel).Error; err != nil {
                t.Fatalf("failed to create novel: %v", err)
        }

        for i := 1; i <= chapterCount; i++ {
                db.Create(&model.Chapter{
                        NovelID: novel.ID,
                        Number:  i,
                        Title:   fmt.Sprintf("Chapter %d", i),
                })
        }
        return novel
}

// ── CheckReviewGate ──

func TestCheckReviewGate_FewerThan5Chapters(t *testing.T) {
        svc, _ := newReviewService(t)

        tests := []struct {
                name  string
                count int
        }{
                {"0 chapters", 0},
                {"1 chapter", 1},
                {"4 chapters", 4},
        }

        for _, tt := range tests {
                t.Run(tt.name, func(t *testing.T) {
                        n := createReviewNovel(t, svc.DB, tt.count)
                        err := svc.CheckReviewGate(n.ID)
                        if err == nil {
                                t.Error("expected error for fewer than 5 chapters")
                        }
                })
        }
}

func TestCheckReviewGate_Exactly5Chapters(t *testing.T) {
        svc, _ := newReviewService(t)
        novel := createReviewNovel(t, svc.DB, 5)

        err := svc.CheckReviewGate(novel.ID)
        if err != nil {
                t.Errorf("expected no error for 5 chapters, got %v", err)
        }
}

func TestCheckReviewGate_MoreThan5Chapters(t *testing.T) {
        svc, _ := newReviewService(t)
        novel := createReviewNovel(t, svc.DB, 10)

        err := svc.CheckReviewGate(novel.ID)
        if err != nil {
                t.Errorf("expected no error for 10 chapters, got %v", err)
        }
}

func TestCheckReviewGate_NonexistentNovel(t *testing.T) {
        svc, _ := newReviewService(t)

        err := svc.CheckReviewGate(99999)
        // No chapters means 0 < 5, so it should return an error
        if err == nil {
                t.Error("expected error for nonexistent novel (0 chapters)")
        }
}

// ── CheckEditLimit ──

func TestCheckEditLimit_UnderLimit(t *testing.T) {
        svc, db := newReviewService(t)
        user := createReviewUser(t, db)
        novel := createReviewNovel(t, db, 10)

        review := model.Review{
                UserID:    user.ID,
                NovelID:   novel.ID,
                Rating:    4,
                Content:   "Great novel!",
                EditCount: 2,
        }
        db.Create(&review)

        count, err := svc.CheckEditLimit(review.ID)
        if err != nil {
                t.Fatalf("unexpected error: %v", err)
        }
        if count != 2 {
                t.Errorf("expected edit count 2, got %d", count)
        }
}

func TestCheckEditLimit_AtLimit(t *testing.T) {
        svc, db := newReviewService(t)
        user := createReviewUser(t, db)
        novel := createReviewNovel(t, db, 10)

        review := model.Review{
                UserID:    user.ID,
                NovelID:   novel.ID,
                Rating:    4,
                Content:   "Great novel!",
                EditCount: 5,
        }
        db.Create(&review)

        count, err := svc.CheckEditLimit(review.ID)
        if err == nil {
                t.Fatal("expected error when edit count >= 5")
        }
        if count != 5 {
                t.Errorf("expected edit count 5, got %d", count)
        }
}

func TestCheckEditLimit_OverLimit(t *testing.T) {
        svc, db := newReviewService(t)
        user := createReviewUser(t, db)
        novel := createReviewNovel(t, db, 10)

        review := model.Review{
                UserID:    user.ID,
                NovelID:   novel.ID,
                Rating:    3,
                Content:   "Good novel!",
                EditCount: 8,
        }
        db.Create(&review)

        count, err := svc.CheckEditLimit(review.ID)
        if err == nil {
                t.Fatal("expected error when edit count > 5")
        }
        if count != 8 {
                t.Errorf("expected edit count 8, got %d", count)
        }
}

func TestCheckEditLimit_NonexistentReview(t *testing.T) {
        svc, _ := newReviewService(t)

        _, err := svc.CheckEditLimit(99999)
        if err == nil {
                t.Fatal("expected error for nonexistent review")
        }
}

func TestCheckEditLimit_Edges(t *testing.T) {
        svc, db := newReviewService(t)
        user := createReviewUser(t, db)
        novel := createReviewNovel(t, db, 10)

        tests := []struct {
                name      string
                editCount uint
                wantErr   bool
        }{
                {"0 edits", 0, false},
                {"4 edits", 4, false},
                {"5 edits", 5, true},
                {"6 edits", 6, true},
        }

        for _, tt := range tests {
                t.Run(tt.name, func(t *testing.T) {
                        review := model.Review{
                                UserID:    user.ID,
                                NovelID:   novel.ID,
                                Rating:    4,
                                Content:   "Review content",
                                EditCount: tt.editCount,
                        }
                        db.Create(&review)

                        _, err := svc.CheckEditLimit(review.ID)
                        if tt.wantErr && err == nil {
                                t.Error("expected error")
                        }
                        if !tt.wantErr && err != nil {
                                t.Errorf("unexpected error: %v", err)
                        }
                })
        }
}

// ── GetNovelReviews ──

func TestGetNovelReviews_Pagination(t *testing.T) {
        svc, db := newReviewService(t)
        user := createReviewUser(t, db)
        novel := createReviewNovel(t, db, 10)

        // Create 5 reviews (different users for each to avoid unique constraint)
        for i := 0; i < 5; i++ {
                u := model.User{
                        Username:     fmt.Sprintf("reviewer%d", i),
                        Email:        fmt.Sprintf("reviewer%d@test.com", i),
                        PasswordHash: "$2a$10$abcdefghijklmnopqrstuvwxABCDEFGHIJ",
                        DisplayName:  fmt.Sprintf("Reviewer %d", i),
                }
                db.Create(&u)

                db.Create(&model.Review{
                        UserID:  u.ID,
                        NovelID: novel.ID,
                        Rating:  uint((i % 5) + 1),
                        Content: fmt.Sprintf("Review %d content here", i),
                })
        }

        // Page 1, limit 2
        reviews, total, err := svc.GetNovelReviews(novel.ID, 1, 2)
        if err != nil {
                t.Fatalf("unexpected error: %v", err)
        }
        if total != 5 {
                t.Errorf("expected total 5, got %d", total)
        }
        if len(reviews) != 2 {
                t.Fatalf("expected 2 reviews on page 1, got %d", len(reviews))
        }

        // Page 3, limit 2 (1 remaining)
        reviews3, _, err := svc.GetNovelReviews(novel.ID, 3, 2)
        if err != nil {
                t.Fatalf("unexpected error: %v", err)
        }
        if len(reviews3) != 1 {
                t.Fatalf("expected 1 review on page 3, got %d", len(reviews3))
        }
}

func TestGetNovelReviews_Empty(t *testing.T) {
        svc, _ := newReviewService(t)
        novel := createReviewNovel(t, svc.DB, 10)

        reviews, total, err := svc.GetNovelReviews(novel.ID, 1, 10)
        if err != nil {
                t.Fatalf("unexpected error: %v", err)
        }
        if total != 0 {
                t.Errorf("expected total 0, got %d", total)
        }
        if len(reviews) != 0 {
                t.Errorf("expected 0 reviews, got %d", len(reviews))
        }
}

func TestGetNovelReviews_ExcludesReplies(t *testing.T) {
        svc, db := newReviewService(t)
        user := createReviewUser(t, db)
        novel := createReviewNovel(t, svc.DB, 10)

        // Create a parent review
        parent := model.Review{
                UserID:  user.ID,
                NovelID: novel.ID,
                Rating:  4,
                Content: "Parent review",
        }
        db.Create(&parent)

        // Create a reply (child review)
        replier := model.User{
                Username:     "replier",
                Email:        "replier@test.com",
                PasswordHash: "$2a$10$abcdefghijklmnopqrstuvwxABCDEFGHIJ",
                DisplayName:  "Replier",
        }
        db.Create(&replier)
        db.Create(&model.Review{
                UserID:    replier.ID,
                NovelID:   novel.ID,
                Rating:    3,
                Content:   "Reply content",
                ParentID:  &parent.ID,
        })

        reviews, total, err := svc.GetNovelReviews(novel.ID, 1, 10)
        if err != nil {
                t.Fatalf("unexpected error: %v", err)
        }
        if total != 1 {
                t.Errorf("expected total 1 (excluding replies), got %d", total)
        }
        if len(reviews) != 1 {
                t.Fatalf("expected 1 review, got %d", len(reviews))
        }
        if reviews[0].ID != parent.ID {
                t.Errorf("expected parent review, got review ID %d", reviews[0].ID)
        }
}

func TestGetNovelReviews_DefaultPagination(t *testing.T) {
        svc, _ := newReviewService(t)

        // Page 0 → clamped to 1, limit 0 → clamped to 10
        reviews, _, err := svc.GetNovelReviews(99999, 0, 0)
        if err != nil {
                t.Fatalf("unexpected error: %v", err)
        }
        if reviews == nil {
                t.Error("expected non-nil result")
        }
}

// ── GetUserReview ──

func TestGetUserReview_Found(t *testing.T) {
        svc, db := newReviewService(t)
        user := createReviewUser(t, db)
        novel := createReviewNovel(t, db, 10)

        review := model.Review{
                UserID:  user.ID,
                NovelID: novel.ID,
                Rating:  5,
                Content: "Amazing!",
        }
        db.Create(&review)

        found, err := svc.GetUserReview(user.ID, novel.ID)
        if err != nil {
                t.Fatalf("unexpected error: %v", err)
        }
        if found.Content != "Amazing!" {
                t.Errorf("expected content 'Amazing!', got '%s'", found.Content)
        }
}

func TestGetUserReview_NotFound(t *testing.T) {
        svc, _ := newReviewService(t)

        _, err := svc.GetUserReview(1, 1)
        if err == nil {
                t.Fatal("expected error for missing review")
        }
}

// ── CreateReview ──

func TestCreateReview_Success(t *testing.T) {
        svc, db := newReviewService(t)
        user := createReviewUser(t, db)
        novel := createReviewNovel(t, db, 10)

        review, err := svc.CreateReview(user.ID, novel.ID, 4, "Great story!")
        if err != nil {
                t.Fatalf("unexpected error: %v", err)
        }
        if review.Rating != 4 {
                t.Errorf("expected rating 4, got %d", review.Rating)
        }
        if review.Content != "Great story!" {
                t.Errorf("expected content 'Great story!', got '%s'", review.Content)
        }
        if review.ID == 0 {
                t.Error("expected non-zero ID")
        }
}

func TestCreateReview_DuplicateFails(t *testing.T) {
        svc, db := newReviewService(t)
        user := createReviewUser(t, db)
        novel := createReviewNovel(t, db, 10)

        _, err := svc.CreateReview(user.ID, novel.ID, 4, "First review")
        if err != nil {
                t.Fatalf("first review failed: %v", err)
        }

        // Second review should fail due to unique constraint
        _, err = svc.CreateReview(user.ID, novel.ID, 3, "Second review")
        if err == nil {
                t.Fatal("expected error for duplicate review")
        }
}