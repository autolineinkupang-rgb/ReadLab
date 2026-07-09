# ReadLab — Build Specification v1.0

**Date:** 2026-07-09
**Project:** ReadLab — Full-stack web novel platform

---

## 1. System Overview

Web novel platform dengan arsitektur **monolith frontend + backend terpisah**. Backend Go (Gin + GORM) serve REST API, frontend Next.js 16 (App Router) render client-side dengan data dari API. Database PostgreSQL 16.

```
Browser → Next.js 16 (:3000) → Go Gin API (:8080) → PostgreSQL 16 (:5432)
                                 → Consumet API (novel import)
                                 → Scraper Engine (NovelUpdates, RoyalRoad, NovelBin)
```

---

## 2. Database Model (13 tables)

### 2.1 Core Content

**`novels`**
| Field | Type | Constraints |
|---|---|---|
| gorm.Model | (ID, CreatedAt, UpdatedAt, DeletedAt) | |
| Title | string(500) | not null, index |
| AltTitle | string(500) | |
| Slug | string(500) | uniqueIndex, not null |
| Author | string(200) | |
| AuthorSlug | string(500) | |
| Status | string(20) | default:ongoing, index |
| Views | uint64 | default:0 |
| Rating | float64 | default:0 |
| RatingCount | uint | default:0 |
| Votes | uint | default:0 |
| Chapters | int | default:0 |
| Readers | int | default:0 |
| Chars | string(20) | |
| AIPercent | string(10) | |
| Description | text | |
| CoverURL | string(1000) | |
| SourceURL | string(1000) | |
| RequestedBy | string(200) | |
| ReleasedBy | string(200) | |
| AddedAt | time | autoCreateTime |
| WriterID | *uint | index, FK→users |
| Genres | []Genre | m2m: novel_genres |

**`novel_genres`** — NovelID(uint,pk), GenreID(uint,pk)

**`genres`** — gorm.Model, Slug(unique,50), Name(100)

**`chapters`** — gorm.Model, NovelID(uint,idx,FK→novels), Number(int,idx), Title(500), Content(text), IsLocked(bool,def:false), TicketCost(int,def:0)

### 2.2 Users & Auth

**`users`**
| Field | Type | Constraints |
|---|---|---|
| Username | string(100) | uniqueIndex, not null |
| Email | string(255) | uniqueIndex, not null |
| PasswordHash | string(255) | json:"-" |
| DisplayName | string(100) | |
| AvatarURL | string(1000) | |
| Tickets | float64 | default:0 |
| XP | int64 | default:0 |
| Role | string(20) | default:member |
| LastDailyClaim | *time.Time | index |
| AITranslateProvider | string(50) | default:openrouter |
| AITranslateModel | string(100) | default:google/gemini-2.0-flash-exp:free |
| AITranslateKey | string(500) | json:"-" |
| AITranslateEndpoint | string(500) | default:https://openrouter.ai/api/v1/chat/completions |
| TranslateTargetLang | string(10) | default:id-ID |
| AITranslateInstruction | string(2000) | |

### 2.3 Reviews & Interaction

**`reviews`** — gorm.Model, UserID(uint,idx_user_novel_parent), NovelID(uint,idx_user_novel_parent), Rating(uint,chk:0-5), Content(text), EditCount(uint,def:0), ParentID(*uint,idx,idx_user_novel_parent,FK→reviews), Replies←[]Review

**`votes`** — gorm.Model, UserID(uint,idx_user_novel), NovelID(uint,idx_user_novel)

**`requests`** — gorm.Model, UserID(uint,idx), NovelTitle(500), NovelURL(1000), Source(100), Status(20,def:pending), Votes(uint,def:0)

### 2.4 Reading & Library

**`reading_histories`** — gorm.Model, UserID(uint,idx_user_novel_chapter), NovelID(uint,idx_user_novel_chapter), ChapterID(uint,idx_user_novel_chapter), Progress(float64,def:0), XpClaimed(bool,def:false)

**`novel_follows`** — gorm.Model, UserID(uint,idx_user_novel_follow), NovelID(uint,idx_user_novel_follow)

### 2.5 Economy

**`ticket_transactions`** — gorm.Model, UserID(uint,idx), Amount(float64), Type(20: purchase/spend/reward), RefType(50: daily/novel_contribution/monthly_xp/upgrade_edit/upgrade_gate/upgrade_duplicate), RefID(uint), Note(500), Date(auto)

**`ticket_configs`** — gorm.Model, Key(unique,100), Value(float64,def:0), Label(255)

| Key | Default | Purpose |
|---|---|---|
| daily_reward | 2 | Daily login tickets |
| novel_contribution | 100 | Reward for creating novel |
| monthly_leaderboard | 50 | Monthly XP → tickets |
| edit_reset_cost | 20 | Reset review edit limit |
| gate_bypass_cost | 50 | Skip review gate |
| replace_review_cost | 100 | Replace existing review |

### 2.6 Content & Social

**`news`** — gorm.Model, Title(500), Content(text), Type(50,idx: news/changelog), Slug(unique,500)

**`shares`** — gorm.Model, UserID(uint,idx_user_novel), NovelID(uint,idx_user_novel), Platform(50)

---

## 3. Backend Architecture

### 3.1 Directory Layout

```
backend/
├── cmd/
│   ├── server/main.go          # Entry point
│   └── seed/main.go            # Sample data seeder
├── internal/
│   ├── config/config.go        # ENV loader
│   ├── handler/                # 27 handler files, 33 route handlers
│   ├── middleware/              # auth, role, cors, ratelimit, logger, security
│   ├── model/                  # 13 GORM models
│   ├── router/router.go        # Route definitions (189 lines)
│   ├── importer/               # Consumet API client
│   ├── scraper/                # NovelUpdates, RoyalRoad, NovelBin
│   └── lncrawl/                # LN Crawl HTTP client
└── migrations/                 # 12 SQL reference files (AutoMigrate in use)
```

### 3.2 Global Middleware Chain

```
SecurityHeaders() → CORS(frontendURL) → Logger()
```

### 3.3 Route Groups & Auth Tiers

| Group | Prefix | Middleware | Count |
|---|---|---|---|
| `api` | `/api/v1` | Global | 18 public |
| `authGroup` | `/auth` | + RateLimiter(10/min) | 3 public |
| `authMeGroup` | `/auth` | + AuthRequired | 1 protected |
| `protected` | `/api/v1` | + AuthRequired | 17 protected |
| `writerGroup` | `/api/v1` | + RequireRole(writer,admin) | 7 writer |
| `adminGroup` | `/api/v1` | + RequireRole(admin) | 20 admin |

### 3.4 Complete Route Table

**Public (no auth):**
```
GET    /health
GET    /novels                          # List with 14 filters
GET    /novels/trending                 # Top 20 by views
GET    /novels/recommendations          # Top 12 by rating
GET    /novels/random                   # ?limit=N
GET    /novels/:id                      # Detail + genres
GET    /novels/:id/chapters             # Chapter list
GET    /novels/:id/chapters/:num        # Chapter (optional auth)
GET    /chapters/:id                    # Chapter by ID
GET    /search                          # ?q=&page=&limit=
GET    /genres                          # All genres
GET    /ranking/:period                 # daily|weekly|monthly|all_time
GET    /updates                         # Recent chapter updates
GET    /news                            # ?type=&page=&limit=
GET    /news/:id
GET    /stats                           # Platform stats
GET    /leaderboard                     # ?sort=xp|tickets
GET    /author/:name/novels
GET    /profile/:id                     # Public profile
GET    /config/upgrade-costs            # Current ticket costs
POST   /auth/register                   # Rate-limited: 10/min
POST   /auth/login                      # Rate-limited: 10/min
POST   /auth/logout
POST   /translate                       # Legacy Google Translate
GET    /novels/import/search            # Rate-limited: 30/min
```

**Protected (auth required):**
```
GET    /auth/me                         # User profile + daily_reward status
PUT    /auth/password                   # Change password
POST   /votes
POST   /requests
GET    /requests
GET    /library                         # Follows + history
POST   /novels/:id/reviews
PUT    /novels/:id/reviews/:reviewId
GET    /novels/:id/reviews
POST   /novels/:id/chapters/:num/read   # Track read
POST   /novels/:id/chapters/:num/xp     # Claim XP
GET    /novels/:id/my-progress
POST   /novels/:id/share
POST   /rewards/daily
GET    /rewards/status
GET    /user/ai-settings
PUT    /user/ai-settings
POST   /translate/ai
```

**Writer (writer/admin):**
```
POST   /novels
PUT    /novels/:id
DELETE /novels/:id
POST   /admin/novels/:id/chapters
PUT    /admin/novels/:id/chapters/:chapterID
GET    /admin/novels/:id/chapters
GET    /admin/chapters/:id
```

**Admin (admin only):**
```
GET    /admin/users
GET    /admin/users/:id
PUT    /admin/users/:id
DELETE /admin/users/:id
POST   /admin/users/admin
GET    /admin/stats
GET    /admin/reviews
DELETE /admin/reviews/:id
GET    /admin/requests
PUT    /requests/:id
POST   /admin/news
PUT    /admin/news/:id
DELETE /admin/news/:id
DELETE /admin/chapters/:id
POST   /admin/rewards/monthly
GET    /admin/config/tickets
PUT    /admin/config/tickets
POST   /novels/import
POST   /novels/scrape
POST   /novels/scrape/import
POST   /novels/lncrawl
```

### 3.5 Auth Flow

```
Register/Login → JWT (HS256, 7d expiry, claims: user_id, role)
  → Set cookie: auth_token (HttpOnly, Secure?, SameSite=Strict, Path=/)
  → Client reads /auth/me → UserData + daily_reward status
  → Password change: PUT /auth/password (validates current + rules)

Middleware chain:
  OptionalAuth(jwtSecret, DB) → sets user_id + role if token valid
  AuthRequired(jwtSecret, DB) → 401 if no valid token
  RequireRole("admin") → 403 if role not in allowed list
```

---

## 4. Frontend Architecture

### 4.1 Directory Layout

```
src/
├── app/
│   ├── layout.tsx               # Root layout
│   ├── page.tsx                 # Landing page (7 API sections)
│   └── en/
│       ├── (main)/layout.tsx    # App shell: Navbar + Sidebar + Footer
│       ├── novel/[id]/[slug]/page.tsx  # Novel detail + reviews (1152 lines)
│       └── ... (40+ pages)
├── components/
│   ├── Navbar.tsx, Footer.tsx, Sidebar.tsx
│   ├── ChapterReader.tsx        # Reader: font, theme, TTS, translation
│   ├── NovelCard.tsx / NovelCardSmall.tsx
│   ├── UpdateItem.tsx, SectionHeader.tsx, RequireRole.tsx
│   └── ui/                      # Card, Button, Input, Breadcrumb, GenreTag, Pagination, ToggleGroup, Icons
├── lib/
│   ├── api.ts                   # 432 lines — all API client functions
│   ├── AuthContext.tsx           # User state + auth methods
│   ├── navigation.ts            # Nav link definitions
│   ├── utils.ts                 # stripHtml, formatViews
│   ├── mockData.ts              # UNUSED (all pages dynamic)
│   └── types.ts                 # Genre, Novel, Chapter, ProfileData
```

### 4.2 Page Routing (40+ pages)

**`/en/(main)` layout — sidebar + navbar:**
```
/                             → redirect to /en
/en                           → English landing
/en/novel-finder              → Advanced search (14 filters)
/en/novel-list                → Filterable grid
/en/ranking/[period]          → Daily/weekly/monthly
/en/leaderboard               → Top users
/en/trending                  → Trending novels
/en/recommendation            → Top rated
/en/random-novels             → Random picker
/en/library                   → Follows + history
/en/novel/[id]/[slug]         → Detail + reviews + upgrade dialog
/en/profile/[id]              → Profile + password change + AI settings
/en/profile/vote-serie        → Vote novels
/en/profile/request-serie     → Request novels
/en/news                      → News list
/en/news/[id]                 → News detail
/en/public-stats              → Platform stats
/en/author/[name]             → Author's novels
/en/about-us, /en/contact-us, /en/privacy-policy, /en/terms-of-use, /en/cookie-policy, /en/dmca

**Writer:** /en/writer
**Admin:**  /en/admin, /en/admin/users, /en/admin/novels, /en/admin/novels/[id]/chapters,
            /en/admin/requests, /en/admin/reviews, /en/admin/news, /en/admin/import,
            /en/admin/ticket-config
```

**`/en/novel` layout — reader shell:**
```
/en/novel/[id]/[slug]/chapter-[num]  → Chapter reader + TOC + TTS + translation
```

### 4.3 Shared State (AuthContext)

```
AuthProvider wraps (main) layout
  state: user (UserData | null), loading (boolean)
  methods: login(), register(), logout(), refresh()
  user.daily_reward: { can_claim, reward }
```

### 4.4 Key Data Flows

```
1. Landing → 7 concurrent API calls (stats, novels, ranking, recommendations,
   updates, spenders, news)

2. Auth → login → set cookie → AuthContext.refresh() → GET /auth/me
   → stores UserData → navbar shows tickets + claim button

3. Review upgrade flow:
   POST/PUT → 403 { upgrade_available, upgrade_cost, upgrade_type }
   → catch error → show Upgrade Dialog with cost from API
   → user confirms → retry with { upgrade: true }
   → backend deducts via Config.Get("edit_reset_cost")

4. Daily reward flow:
   Navbar shows "Claim X Tickets" from AuthContext.user.daily_reward
   → POST /rewards/daily → check LastDailyClaim (Makassar TZ)
   → award via Config.Get("daily_reward") → refresh()

5. Upgrade costs:
   Page mount → GET /config/upgrade-costs → all "Pay N Tickets" dynamic

6. Admin config change:
   Save → PUT /admin/config/tickets → DB update → Reload() cache
   → refresh() → Navbar updated
```

---

## 5. TicketConfigService (In-Memory Cache)

```
NewTicketConfigService(db) → Load all rows → cache map[string]float64
  Reload(): DB.Find() → replace cache (error ignored)
  Get(key): RLock → read cache → 0 if missing
  Used by: Auth, Reward, Review, Novel handlers
  Updated by: TicketConfigHandler.Update → DB → Reload()

⚠ Limitation: Per-instance cache. Multiple backend instances = stale reads.
```

---

## 6. Tech Stack

### Backend (Go 1.25)
```
gin-gonic/gin v1.12          → HTTP
golang-jwt/jwt v5.3.1       → JWT
gorm.io/gorm v1.31.2        → ORM
gorm.io/driver/postgres v1.6.0 → PG
golang.org/x/crypto         → bcrypt
joho/godotenv v1.5.1        → .env
PuerkitoBio/goquery         → HTML parser
```

### Frontend (Node 22+, Next.js 16)
```
next 16.2.10 (Turbopack)
react 19.2.4 / react-dom 19.2.4
tailwindcss 4
@tiptap/react + extensions  → Rich text editor
dompurify / isomorphic-dompurify → HTML sanitization
@consumet/extensions        → NovelUpdates API
mammoth                     → DOCX import
typescript 5
```

### Infrastructure
```
PostgreSQL 16 Alpine
Docker Compose (3 services: db, backend, frontend)
No CI/CD
```

---

## 7. Deployment Topology

```yaml
services:
  db:       postgres:16-alpine  → :5432 (host: 5433)
  backend:  Go binary (built from Dockerfile)  → :8080
  frontend: Next.js standalone (built from Dockerfile)  → :3000
```

Network: `backend → db:5432`, `frontend → backend:8080` (Next.js rewrites: `/api/*` → `http://backend:8080/api/v1/*`)

Backend config loading: `godotenv.Load()` → env vars with fallbacks.

---

## 8. Key Business Logic Rules

| Rule | Enforcement | Location |
|---|---|---|
| One review per user per novel | Unique index `idx_user_novel_parent` | model/review.go |
| Max 5 edits per review | EditCount >= 5 → 403 upgrade_available | review_handler.go |
| Review gate: need 5 chapters | chapterCount < 5 → 403 upgrade_available | review_handler.go |
| One vote per user per novel | Unique index `idx_user_novel` | model/vote.go |
| Daily reward: once per day | LastDailyClaim < today midnight (Makassar TZ) | reward_handler.go |
| Novel contribution reward | After Create() transaction succeeds | novel_handler.go |
| Monthly reward | Admin-triggered, resets all XP to 0 | reward_handler.go |
| Chapter XP | One per chapter (XpClaimed flag) | reading_handler.go |
| Password strength | 8+ chars, uppercase, lowercase, digit, special | auth_handler.go |
| Login rate limit | 10/min per IP | router.go |
| Writer ownership | WriterID check in admin chapter handlers | admin_chapter_handler.go |

---

## 9. Identified Gaps & Known Issues

### Architecture
- No service layer — all business logic in handlers
- In-memory TicketConfigService cache — not suitable for multi-instance
- No graceful shutdown — `r.Run()` blocks without signal handling

### Security
- SQL injection in ranking_handler.go:26 — raw string interpolation
- No token revocation — JWT valid 7 days, no blacklist
- No CSRF protection — SameSite=Strict mitigates partially
- DOMPurify without ALLOWED_ATTR — style/event attributes allowed
- Type assertion without guard — 15+ places use `userID.(uint)` can panic

### Testing
- Zero frontend tests
- Backend tests use SQLite in-memory (not PostgreSQL)
- No CI/CD pipeline

### Operations
- Docker containers run as root
- No health check or restart policy in docker-compose
- No database backup
- No metrics, no structured logging, no error reporting
- No connection pool limits configured

### Features
- Follow/unfollow novel — model exists, UI is local state toggle only, no API call
- No notifications system at all
- "Continue Reading" tracked but not surfaced effectively
- No password reset flow
- No monetization — tickets cannot be purchased
- Search autocomplete missing
- Dark mode not persisted between sessions
