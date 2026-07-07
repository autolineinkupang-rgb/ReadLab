# Review Feature Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Allow users to rate (1-5 stars) and review novels after reading at least 5 unique chapters.

**Architecture:** Backend adds Review model, ReviewHandler, and ReadingHandler; frontend adds reviews tab with inline form + reading tracking in ChapterReader.

**Tech Stack:** Go (gin/gorm), Next.js 16, TypeScript

---

### Task 1: Backend — Review Model

**Files:**
- Create: `backend/internal/model/review.go`

- [ ] **Create Review model**

```go
package model

import "gorm.io/gorm"

type Review struct {
	gorm.Model
	UserID  uint   `gorm:"uniqueIndex:idx_user_novel;not null"`
	NovelID uint   `gorm:"uniqueIndex:idx_user_novel;not null"`
	Rating  uint   `gorm:"not null;check:rating >= 1 AND rating <= 5"`
	Content string `gorm:"type:text;not null"`

	User  User  `gorm:"foreignKey:UserID"`
	Novel Novel `gorm:"foreignKey:NovelID"`
}
```

- [ ] **AutoMigrate Review** — add to `backend/cmd/server/main.go`

Find the `AutoMigrate` call and add `&model.Review{},`:

```go
db.AutoMigrate(
    &model.User{},
    &model.Novel{},
    &model.Chapter{},
    &model.Genre{},
    &model.Vote{},
    &model.ReadingHistory{},
    &model.NovelFollow{},
    &model.Request{},
    &model.News{},
    &model.Ticket{},
    &model.Review{},   // <-- add
)
```

- [ ] **Commit**

```bash
git add backend/internal/model/review.go backend/cmd/server/main.go
git commit -m "feat: add Review model"
```

---

### Task 2: Backend — ReviewHandler (List + Create)

**Files:**
- Create: `backend/internal/handler/review_handler.go`

- [ ] **Create ReviewHandler**

```go
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"wtr-lab-clone/backend/internal/model"
)

type ReviewHandler struct {
	DB *gorm.DB
}

func NewReviewHandler(db *gorm.DB) *ReviewHandler {
	return &ReviewHandler{DB: db}
}

type CreateReviewRequest struct {
	Rating  uint   `json:"rating" binding:"required,min=1,max=5"`
	Content string `json:"content" binding:"required,min=10,max=2000"`
}

type reviewResponse struct {
	ID        uint   `json:"id"`
	Rating    uint   `json:"rating"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	User      struct {
		ID          uint   `json:"id"`
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
		AvatarURL   string `json:"avatar_url"`
	} `json:"user"`
}

func toReviewResponse(r model.Review) reviewResponse {
	var resp reviewResponse
	resp.ID = r.ID
	resp.Rating = r.Rating
	resp.Content = r.Content
	resp.CreatedAt = r.CreatedAt.Format("2006-01-02T15:04:05Z")
	resp.User.ID = r.User.ID
	resp.User.Username = r.User.Username
	resp.User.DisplayName = r.User.DisplayName
	resp.User.AvatarURL = r.User.AvatarURL
	return resp
}

func (h *ReviewHandler) List(c *gin.Context) {
	novelID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid novel id"})
		return
	}

	var reviews []model.Review
	h.DB.Where("novel_id = ?", novelID).
		Preload("User").
		Order("created_at DESC").
		Find(&reviews)

	var totalCount int64
	h.DB.Model(&model.Review{}).Where("novel_id = ?", novelID).Count(&totalCount)

	distribution := map[uint]int64{1: 0, 2: 0, 3: 0, 4: 0, 5: 0}
	type ratingCount struct {
		Rating uint
		Count  int64
	}
	var counts []ratingCount
	h.DB.Model(&model.Review{}).Select("rating, count(*) as count").
		Where("novel_id = ?", novelID).
		Group("rating").Find(&counts)
	for _, rc := range counts {
		distribution[rc.Rating] = rc.Count
	}

	var avg float64
	if totalCount > 0 {
		h.DB.Model(&model.Review{}).
			Select("avg(rating)").
			Where("novel_id = ?", novelID).
			Scan(&avg)
	}

	items := make([]reviewResponse, len(reviews))
	for i, r := range reviews {
		items[i] = toReviewResponse(r)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": items,
		"rating_summary": gin.H{
			"average":      avg,
			"count":        totalCount,
			"distribution": distribution,
		},
	})
}

func (h *ReviewHandler) Create(c *gin.Context) {
	userID, _ := c.Get("user_id")

	novelID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid novel id"})
		return
	}

	var novel model.Novel
	if err := h.DB.First(&novel, novelID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "novel not found"})
		return
	}

	var req CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing model.Review
	if err := h.DB.Where("user_id = ? AND novel_id = ?", userID, novelID).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "you have already reviewed this novel"})
		return
	}

	var chapterCount int64
	h.DB.Model(&model.ReadingHistory{}).
		Where("user_id = ? AND novel_id = ?", userID, novelID).
		Count(&chapterCount)

	if chapterCount < 5 {
		c.JSON(http.StatusForbidden, gin.H{
			"error":         "you need to read at least 5 chapters before reviewing",
			"chapter_count": chapterCount,
		})
		return
	}

	review := model.Review{
		UserID:  userID.(uint),
		NovelID: uint(novelID),
		Rating:  req.Rating,
		Content: req.Content,
	}

	if err := h.DB.Create(&review).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create review"})
		return
	}

	var avg float64
	var count int64
	h.DB.Model(&model.Review{}).
		Select("avg(rating)").
		Where("novel_id = ?", novelID).
		Scan(&avg)
	h.DB.Model(&model.Review{}).
		Where("novel_id = ?", novelID).
		Count(&count)

	h.DB.Model(&novel).Updates(map[string]interface{}{
		"Rating":      avg,
		"RatingCount": count,
	})

	h.DB.Preload("User").First(&review, review.ID)

	c.JSON(http.StatusCreated, gin.H{"data": toReviewResponse(review)})
}
```

- [ ] **Commit**

```bash
git add backend/internal/handler/review_handler.go
git commit -m "feat: add ReviewHandler (List + Create) with 5-chapter check"
```

---

### Task 3: Backend — ReadingHandler (TrackRead + Progress)

**Files:**
- Create: `backend/internal/handler/reading_handler.go`

- [ ] **Create ReadingHandler**

```go
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"wtr-lab-clone/backend/internal/model"
)

type ReadingHandler struct {
	DB *gorm.DB
}

func NewReadingHandler(db *gorm.DB) *ReadingHandler {
	return &ReadingHandler{DB: db}
}

func (h *ReadingHandler) TrackRead(c *gin.Context) {
	userID, _ := c.Get("user_id")

	novelID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid novel id"})
		return
	}

	chapterNum, err := strconv.ParseUint(c.Param("num"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chapter number"})
		return
	}

	var chapter model.Chapter
	if err := h.DB.Where("novel_id = ? AND number = ?", novelID, chapterNum).First(&chapter).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "chapter not found"})
		return
	}

	var existing model.ReadingHistory
	result := h.DB.Where("user_id = ? AND novel_id = ? AND chapter_id = ?", userID, novelID, chapter.ID).First(&existing)

	if result.Error != nil {
		entry := model.ReadingHistory{
			UserID:    userID.(uint),
			NovelID:   uint(novelID),
			ChapterID: chapter.ID,
		}
		h.DB.Create(&entry)
	} else {
		h.DB.Model(&existing).Update("progress", 100)
	}

	c.JSON(http.StatusOK, gin.H{"message": "reading progress recorded"})
}

func (h *ReadingHandler) Progress(c *gin.Context) {
	userID, _ := c.Get("user_id")

	novelID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid novel id"})
		return
	}

	var chapterCount int64
	h.DB.Model(&model.ReadingHistory{}).
		Where("user_id = ? AND novel_id = ?", userID, novelID).
		Count(&chapterCount)

	var myReview model.Review
	reviewResult := h.DB.Where("user_id = ? AND novel_id = ?", userID, novelID).Preload("User").First(&myReview)

	var myReviewResp *reviewResponse
	if reviewResult.Error == nil {
		resp := toReviewResponse(myReview)
		myReviewResp = &resp
	}

	c.JSON(http.StatusOK, gin.H{
		"chapter_count": chapterCount,
		"can_review":    chapterCount >= 5,
		"my_review":     myReviewResp,
	})
}
```

- [ ] **Commit**

```bash
git add backend/internal/handler/reading_handler.go
git commit -m "feat: add ReadingHandler (TrackRead + Progress)"
```

---

### Task 4: Backend — Wire Routes

**Files:**
- Modify: `backend/internal/router/router.go`

- [ ] **Add imports and handler registrations**

Add imports:
```go
reviewHandler := handler.NewReviewHandler(db)
readingHandler := handler.NewReadingHandler(db)
```

At the bottom of `Setup()` before `return r`, add:
```go
api.GET("/novels/:id/reviews", reviewHandler.List)
api.POST("/novels/:id/reviews", middleware.AuthRequired(jwtSecret), reviewHandler.Create)
api.POST("/novels/:id/chapters/:num/read", middleware.AuthRequired(jwtSecret), readingHandler.TrackRead)
api.GET("/novels/:id/my-progress", middleware.AuthRequired(jwtSecret), readingHandler.Progress)
```

- [ ] **Commit**

```bash
git add backend/internal/router/router.go
git commit -m "feat: wire review + reading routes"
```

---

### Task 5: Frontend — API Client (reviews + reading)

**Files:**
- Modify: `frontend/src/lib/api.ts`

- [ ] **Add review + reading API functions to api.ts**

After `export const adminRequests` block (end of file), add:

```typescript
// Reviews
export const reviews = {
  list: (novelId: number) =>
    fetcher<{ data: ReviewResponse[]; rating_summary: RatingSummary }>(`/novels/${novelId}/reviews`),
  create: (novelId: number, rating: number, content: string) =>
    fetcher<{ data: ReviewResponse }>(`/novels/${novelId}/reviews`, {
      method: "POST",
      body: JSON.stringify({ rating, content }),
    }),
};

// Reading tracking
export const reading = {
  track: (novelId: number, chapterNum: number) =>
    fetcher<{ message: string }>(`/novels/${novelId}/chapters/${chapterNum}/read`, {
      method: "POST",
    }),
  progress: (novelId: number) =>
    fetcher<{ chapter_count: number; can_review: boolean; my_review: ReviewResponse | null }>(
      `/novels/${novelId}/my-progress`
    ),
};

// Types for reviews
export interface ReviewResponse {
  id: number;
  rating: number;
  content: string;
  created_at: string;
  user: {
    id: number;
    username: string;
    display_name: string;
    avatar_url: string;
  };
}

export interface RatingSummary {
  average: number;
  count: number;
  distribution: Record<number, number>;
}
```

- [ ] **Commit**

```bash
git add frontend/src/lib/api.ts
git commit -m "feat: add reviews + reading API client"
```

---

### Task 6: Frontend — Track Reading in ChapterReader

**Files:**
- Modify: `frontend/src/components/ChapterReader.tsx`

- [ ] **Add reading.track call when chapter loads**

Find the `useEffect` that runs when `chapter?.number` changes (around line 93-98):

```tsx
useEffect(() => {
    if (!chapter) return;
    window.scrollTo({ top: 0, behavior: "instant" as ScrollBehavior });
    setShowToc(false);
}, [chapter?.number]);
```

Extend it to also track reading (fire-and-forget):

```tsx
useEffect(() => {
    if (!chapter) return;
    window.scrollTo({ top: 0, behavior: "instant" as ScrollBehavior });
    setShowToc(false);

    if (user && novel?.id && chapter.number) {
      import("@/lib/api").then(({ reading }) => {
        reading.track(novel.id, chapter.number).catch(() => {});
      });
    }
}, [chapter?.number, user, novel?.id]);
```

Also ensure `user` from `useAuth()` is available in scope (already added in previous tasks).

- [ ] **Commit**

```bash
git add frontend/src/components/ChapterReader.tsx
git commit -m "feat: track reading progress when viewing chapter"
```

---

### Task 7: Frontend — Novel Detail Page (Reviews Tab)

**Files:**
- Modify: `frontend/src/app/en/(main)/novel/[id]/[slug]/page.tsx`

- [ ] **Add imports at top**

```typescript
import { reviews, reading } from "@/lib/api";
import { useAuth } from "@/lib/AuthContext";
import type { ReviewResponse, RatingSummary } from "@/lib/api";
```

- [ ] **Add state variables** (after existing state, before `useEffect`)

```typescript
const [reviewsData, setReviewsData] = useState<ReviewResponse[]>([]);
const [ratingSummary, setRatingSummary] = useState<RatingSummary | null>(null);
const [myReview, setMyReview] = useState<ReviewResponse | null>(null);
const [chapterCount, setChapterCount] = useState(0);
const [formRating, setFormRating] = useState(0);
const [formHoverRating, setFormHoverRating] = useState(0);
const [formContent, setFormContent] = useState("");
const [formSubmitting, setFormSubmitting] = useState(false);
const [formError, setFormError] = useState("");
const [reviewsLoading, setReviewsLoading] = useState(false);
const { user } = useAuth();
```

- [ ] **Add useEffect to fetch reviews + progress** (after existing useEffects)

```typescript
useEffect(() => {
    if (!id) return;
    setReviewsLoading(true);
    Promise.all([
      reviews.list(parseInt(id)),
      user ? reading.progress(parseInt(id)) : Promise.resolve(null),
    ])
      .then(([reviewsRes, progressRes]) => {
        setReviewsData(reviewsRes.data);
        setRatingSummary(reviewsRes.rating_summary);
        if (progressRes) {
          setChapterCount(progressRes.chapter_count);
          setMyReview(progressRes.my_review);
        }
      })
      .catch(() => {})
      .finally(() => setReviewsLoading(false));
}, [id, user]);
```

- [ ] **Replace the existing reviews tab content** (lines 279-283)

Old placeholder:
```tsx
{activeTab === "reviews" && (
    <div className="bg-card border border-line rounded-xl p-6 text-center text-sm text-gray-500">
      No reviews yet. Be the first to review!
    </div>
)}
```

Replace with:

```tsx
{activeTab === "reviews" && (
    <div className="space-y-6">
      {/* Rating Summary */}
      {ratingSummary && ratingSummary.count > 0 && (
        <div className="bg-card border border-line rounded-xl p-6">
          <div className="flex flex-col sm:flex-row items-center gap-6">
            <div className="text-center">
              <div className="text-5xl font-bold text-yellow-400">{ratingSummary.average.toFixed(1)}</div>
              <div className="text-sm text-gray-500 mt-1">{ratingSummary.count} review{ratingSummary.count !== 1 ? "s" : ""}</div>
            </div>
            <div className="flex-1 w-full space-y-1.5">
              {[5, 4, 3, 2, 1].map((star) => {
                const pct = ratingSummary.count > 0
                  ? ((ratingSummary.distribution[star] || 0) / ratingSummary.count) * 100
                  : 0;
                return (
                  <div key={star} className="flex items-center gap-2 text-sm">
                    <span className="text-yellow-400 w-6 text-right">{star}</span>
                    <svg className="w-3.5 h-3.5 text-yellow-400" fill="currentColor" viewBox="0 0 20 20">
                      <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
                    </svg>
                    <div className="flex-1 h-2 bg-card-hover rounded-full overflow-hidden">
                      <div className="h-full bg-yellow-400 rounded-full" style={{ width: `${pct}%` }} />
                    </div>
                    <span className="text-gray-500 w-6 text-right">{ratingSummary.distribution[star] || 0}</span>
                  </div>
                );
              })}
            </div>
          </div>
        </div>
      )}

      {/* Write Review Form */}
      {user ? (
        myReview ? (
          <div className="bg-card border border-line rounded-xl p-6">
            <p className="text-sm text-green-400 mb-2">You have reviewed this novel.</p>
            <ReviewCard review={myReview} />
          </div>
        ) : chapterCount < 5 ? (
          <div className="bg-card border border-line rounded-xl p-6 text-center">
            <p className="text-sm text-gray-500">
              Read <strong className="text-yellow-400">{chapterCount}/5</strong> chapters to unlock the review feature.
              <span className="block mt-1 text-xs text-gray-600">Continue reading to share your thoughts!</span>
            </p>
          </div>
        ) : (
          <div className="bg-card border border-line rounded-xl p-6">
            <h3 className="text-sm font-medium text-gray-200 mb-4">Write Your Review</h3>
            <form onSubmit={async (e) => {
              e.preventDefault();
              if (formRating === 0) { setFormError("Please select a rating"); return; }
              if (formContent.trim().length < 10) { setFormError("Review must be at least 10 characters"); return; }
              setFormSubmitting(true);
              setFormError("");
              try {
                const res = await reviews.create(parseInt(id), formRating, formContent);
                setMyReview(res.data);
                setReviewsData((prev) => [res.data, ...prev]);
                if (ratingSummary) {
                  const newCount = ratingSummary.count + 1;
                  const newAvg = ((ratingSummary.average * ratingSummary.count) + formRating) / newCount;
                  const newDist = { ...ratingSummary.distribution };
                  newDist[formRating] = (newDist[formRating] || 0) + 1;
                  setRatingSummary({ average: newAvg, count: newCount, distribution: newDist });
                }
                setFormRating(0);
                setFormContent("");
              } catch (err) {
                setFormError(err instanceof Error ? err.message : "Failed to submit review");
              } finally {
                setFormSubmitting(false);
              }
            }}>
              {/* Star Rating */}
              <div className="flex items-center gap-1 mb-4">
                {[1, 2, 3, 4, 5].map((star) => (
                  <button
                    key={star}
                    type="button"
                    onClick={() => setFormRating(star)}
                    onMouseEnter={() => setFormHoverRating(star)}
                    onMouseLeave={() => setFormHoverRating(0)}
                    className="p-0.5 transition-transform hover:scale-110"
                  >
                    <svg
                      className={`w-7 h-7 ${
                        (formHoverRating || formRating) >= star
                          ? "text-yellow-400"
                          : "text-gray-600"
                      }`}
                      fill="currentColor"
                      viewBox="0 0 20 20"
                    >
                      <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
                    </svg>
                  </button>
                ))}
                {formRating > 0 && (
                  <span className="text-sm text-yellow-400 ml-2">
                    {formRating === 1 ? "Poor" : formRating === 2 ? "Fair" : formRating === 3 ? "Good" : formRating === 4 ? "Very Good" : "Excellent"}
                  </span>
                )}
              </div>

              {/* Textarea */}
              <textarea
                value={formContent}
                onChange={(e) => setFormContent(e.target.value)}
                placeholder="Share your thoughts about this novel (min. 10 characters)..."
                rows={4}
                maxLength={2000}
                className="w-full bg-card-hover border border-line-light rounded-lg px-4 py-3 text-sm text-gray-200 outline-none focus:border-accent transition-colors resize-none"
              />
              <p className="text-xs text-gray-600 mt-1 text-right">{formContent.length}/2000</p>

              {formError && (
                <p className="text-xs text-red-400 mt-2">{formError}</p>
              )}

              <button
                type="submit"
                disabled={formSubmitting}
                className="mt-3 px-6 py-2 bg-violet-600 hover:bg-violet-700 disabled:opacity-50 text-white text-sm font-medium rounded-lg transition-colors"
              >
                {formSubmitting ? "Submitting..." : "Submit Review"}
              </button>
            </form>
          </div>
        )
      ) : (
        <div className="bg-card border border-line rounded-xl p-6 text-center">
          <p className="text-sm text-gray-500">
            <Link href="/en/login" className="text-violet-400 hover:text-violet-300 transition-colors">Login</Link> to leave a review
          </p>
        </div>
      )}

      {/* Reviews List */}
      <div className="space-y-3">
        {reviewsLoading ? (
          <div className="text-center text-sm text-gray-500 py-8">Loading reviews...</div>
        ) : reviewsData.length === 0 ? (
          <div className="bg-card border border-line rounded-xl p-6 text-center text-sm text-gray-500">
            No reviews yet. Be the first to review!
          </div>
        ) : (
          reviewsData.map((review) => (
            <ReviewCard key={review.id} review={review} />
          ))
        )}
      </div>
    </div>
)}
```

- [ ] **Add ReviewCard component** at the bottom of the file (after the closing `}` of NovelDetailPage)

```tsx
function ReviewCard({ review }: { review: ReviewResponse }) {
  return (
    <div className="bg-card border border-line rounded-xl p-4">
      <div className="flex items-start gap-3">
        <div className="w-9 h-9 rounded-full bg-accent flex items-center justify-center text-white text-sm font-bold shrink-0">
          {(review.user.display_name || review.user.username)[0].toUpperCase()}
        </div>
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 flex-wrap">
            <span className="text-sm font-medium text-gray-200">{review.user.display_name || review.user.username}</span>
            <div className="flex items-center gap-0.5">
              {[1, 2, 3, 4, 5].map((star) => (
                <svg
                  key={star}
                  className={`w-3.5 h-3.5 ${star <= review.rating ? "text-yellow-400" : "text-gray-600"}`}
                  fill="currentColor"
                  viewBox="0 0 20 20"
                >
                  <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
                </svg>
              ))}
            </div>
          </div>
          <p className="text-sm text-gray-300 mt-2 leading-relaxed">{review.content}</p>
          <p className="text-xs text-gray-600 mt-2">{timeAgo(review.created_at)}</p>
        </div>
      </div>
    </div>
  );
}

function timeAgo(dateStr: string) {
  const now = Date.now();
  const date = new Date(dateStr).getTime();
  const diff = now - date;
  const minutes = Math.floor(diff / 60000);
  if (minutes < 1) return "just now";
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  if (days < 30) return `${days}d ago`;
  const months = Math.floor(days / 30);
  return `${months}mo ago`;
}
```

- [ ] **Commit**

```bash
git add frontend/src/app/en/\(main\)/novel/\[id\]/\[slug\]/page.tsx
git commit -m "feat: add reviews tab with inline form + review list"
```

---

### Task 8: Verify — Build + Test

- [ ] **Build backend**

```bash
cd backend && go build ./...
```

- [ ] **Build frontend**

```bash
cd frontend && npx next build
```

- [ ] **Final commit** (if fixes needed)

```bash
git add -A && git commit -m "fix: address build issues"
```
