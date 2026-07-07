# Review Feature Design

## Overview
Add a user review system to the novel detail page. Users can rate a novel (1–5 stars) and write a text review, provided they have read at least 5 unique chapters of that novel. Reading progress is tracked automatically when a user views a chapter.

## Backend

### Models

**Review** (`backend/internal/model/review.go`)
| Field | Type | Notes |
|-------|------|-------|
| ID | uint (gorm.Model) | Auto-increment PK |
| UserID | uint | FK → User, unique per novel |
| NovelID | uint | FK → Novel |
| Rating | uint | 1–5 |
| Content | string | Text review, max 2000 chars |
| CreatedAt / UpdatedAt | time.Time | Via gorm.Model |

Unique constraint: `(user_id, novel_id)` — one review per user per novel.

**ReadingHistory** (already exists at `backend/internal/model/reading_history.go`)
Will be reused as-is for tracking. One record per user per novel per chapter.

### Handlers

**ReviewHandler** (`backend/internal/handler/review_handler.go`)
- `List(c *gin.Context)` — `GET /novels/:id/reviews`
  - No auth required (public)
  - Returns: `{ data: Review[], rating_summary: { average, count, distribution } }`
  - Joins User to get username/display_name/avatar_url
  - Ordered by newest first

- `Create(c *gin.Context)` — `POST /novels/:id/reviews`
  - Auth required
  - Body: `{ rating: uint, content: string }`
  - Validates: rating 1–5, content 10–2000 chars
  - Checks: 5+ unique ReadingHistory records for this user+novel
  - If check fails → 403 `{ error: "need to read at least 5 chapters" }`
  - Checks: existing review → 409 `{ error: "already reviewed" }`
  - On success: creates Review, recalculates Novel.Rating + Novel.RatingCount
  - Returns 201 with review data

### ReadingHandler (separate) — `backend/internal/handler/reading_handler.go`

- `TrackRead(c *gin.Context)` — `POST /novels/:id/chapters/:num/read`
  - Auth required
  - Upserts ReadingHistory for (user_id, novel_id, chapter_id)
  - Returns 200

- `GetProgress(c *gin.Context)` — `GET /novels/:id/my-progress`
  - Auth required
  - Returns: `{ chapter_count: uint, can_review: bool, my_review: Review | null }`
  - Counts distinct ReadingHistory records for this user+novel

### Routes
Add to `router.go`:
```
GET    /novels/:id/reviews              → reviewHandler.List
POST   /novels/:id/reviews              → reviewHandler.Create  (auth)
POST   /novels/:id/chapters/:num/read   → readingHandler.Track  (auth)
GET    /novels/:id/my-progress          → readingHandler.Progress  (auth)
```

### Novel Rating Recalculation
After creating a review:
1. Count all reviews for the novel
2. Average all ratings
3. Update `Novel.Rating` and `Novel.RatingCount`

## Frontend

### API Client (`frontend/src/lib/api.ts`)

```typescript
export const reviews = {
  list: (novelId: number) =>
    fetcher<{ data: Review[]; rating_summary: RatingSummary }>(`/novels/${novelId}/reviews`),
  create: (novelId: number, rating: number, content: string) =>
    fetcher<{ data: Review }>(`/novels/${novelId}/reviews`, {
      method: "POST",
      body: JSON.stringify({ rating, content }),
    }),
};

export const reading = {
  track: (novelId: number, chapterNum: number) =>
    fetcher<{ message: string }>(`/novels/${novelId}/chapters/${chapterNum}/read`, {
      method: "POST",
    }),
  progress: (novelId: number) =>
    fetcher<{ chapter_count: number; can_review: boolean; my_review: Review | null }>(
      `/novels/${novelId}/my-progress`
    ),
};
```

### Reading Tracking
In `ChapterReader.tsx` or chapter page:
- On mount (or when chapter loads), if user is logged in, call `reading.track(novelId, chapterNum)`
- Fire-and-forget (no UI feedback needed)

### Novel Detail Page — Reviews Tab

The existing tab system will be updated:

**State:**
- `reviews: Review[]` — list of reviews
- `ratingSummary: { average, count, distribution }` — aggregate stats
- `myReview: Review | null` — current user's review (if any)
- `formRating: number` — 1-5 stars (hover + selected state)
- `formContent: string` — text input
- `formSubmitting: boolean`
- `formError: string`
- `userChapterCount: number` — how many chapters the user has read
- `canReview: boolean` — chapterCount >= 5

**Layout:**
```
┌─────────────────────────────────────┐
│  Rating Summary (stars, avg, count) │
│  Distribution bars (5★, 4★, ...)    │
│                                     │
│  ──── Write Your Review ────        │  ← only if logged in
│  [★★★★★] (clickable stars)         │
│  [textarea: Your review...]         │
│  [Submit Review]                    │
│                                     │
│  ──── Reviews ────                  │
│  [ReviewCard]                       │
│  [ReviewCard]                       │
│  ...                                │
└─────────────────────────────────────┘
```

**If not logged in:** Show "Login to leave a review"

**If logged in but < 5 chapters:** Show "Read at least 5 chapters to leave a review (X/5)"

**If already reviewed:** Show "Your review" card + edit/delete option (future)

### ReviewCard Component
- Avatar initial + username
- Star rating (visual stars)
- Content text
- Relative time (e.g., "2 days ago")

## Files Changed/Created

### Backend (Go)
| Action | File |
|--------|------|
| Create | `backend/internal/model/review.go` |
| Create | `backend/internal/handler/review_handler.go` |
| Modify | `backend/internal/handler/chapter_handler.go` (add TrackRead) |
| Create | `backend/internal/handler/reading_handler.go` |
| Modify | `backend/internal/router/router.go` |
| Modify | `backend/cmd/server/main.go` (AutoMigrate Review) |

### Frontend (TSX/TS)
| Action | File |
|--------|------|
| Modify | `frontend/src/lib/api.ts` |
| Modify | `frontend/src/app/en/(main)/novel/[id]/[slug]/page.tsx` |
| Modify | `frontend/src/components/ChapterReader.tsx` (track reading) |

## Edge Cases
- **Self-review submission:** Prevent duplicate (one per user). 409 if exists.
- **Rating bounds:** Backend clamps to 1-5 if out of range.
- **Concurrent reviews:** DB unique constraint handles race conditions.
- **Chapter count:** Counts distinct chapters only (not repeated reads).
- **Novel not found:** 404 on review create.
- **User deleted:** Reviews remain with "Deleted User" display.
