# ReadLab — Product Requirements Document

## Original Problem Statement
> `https://github.com/autolineinkupang-rgb/ReadLab` — buat lebih sempurna dengan kode yang ada di repository tersebut.
> User choices:
> - Focus: **d. Semua** (bug fix, UI/UX, new features)
> - Clone repo: **a. Ya, clone dan gunakan sebagai basis**
> - Integration: **a. Tidak perlu integrasi baru**
> - Design preference: **b. Bebas — Anda tentukan gaya terbaik**
> - Deployment: **Docker (user deploys via Docker on own server — Option C from confirmation)**

## Tech Stack (UNCHANGED — repo original)
| Layer | Technology |
|---|---|
| Frontend | Next.js 16 (App Router), React 19, TypeScript, Tailwind CSS 4 |
| Backend | Go 1.25+, Gin, GORM |
| Database | PostgreSQL 16 |
| Auth | JWT + HTTP-only cookies + CSRF |
| Deploy | Docker Compose (user runs `docker compose up -d`) |

## Environment Constraints
- **Emergent platform does NOT support Docker/Go/PostgreSQL** — we can only edit code and perform static + Next.js-only verification.
- **User will deploy** the finished code via Docker on their own server/VPS.

## User Personas
1. **Reader** — browses novels, reads chapters, tracks progress, writes reviews
2. **Writer** — creates novels, manages chapters, uses lncrawl/import tools
3. **Admin** — manages users, ticket economy, requests, news, and site config

## Core Requirements (Static — inherited from repo)
- Novel discovery (14-filter finder, ranking, trending, recommendations, random)
- Reader with fontsize / font family / dark-light-sepia themes
- Ticket economy (daily rewards, XP leaderboard rewards, review upgrades)
- Reviews with rating 1–5, edit limits, gate-bypass via tickets
- User library (follows, reading history)
- AI translation (per-user OpenRouter endpoint config)
- Voting & series requests
- Import from Consumet, NovelUpdates, RoyalRoad, NovelBin, lncrawl
- Role-gated endpoints (member / writer / admin)

## What's Been Implemented (this session — 2026-01-11)

### v1.1.0 — Frontend polish pass
**Fixed:**
- NovelCard cover images now actually render (`image` prop was previously dead code)
- NovelCardSmall now shows optional thumbnail
- UpdateItem now shows cover thumbnails (was always placeholder icon)
- Layout flex-parent chain corrected so footer sticks to bottom on short pages
- Hardcoded "16th Giveaway" banner replaced with dynamic `HeroBanner` (news-driven, dismissible)

**Added:**
- `Skeleton` + `NovelCardSkeleton` + `NovelCardSkeletonRow` + `UpdateItemSkeleton` components
- `BackToTop` floating button (appears after 400 px scroll)
- `HeroBanner` — dynamic, dismissible per-news-item via localStorage
- Reading-progress bar prop on `NovelCard`
- Empty-state components with icon + message on every homepage section
- Staggered fade-in entrance animation utility (`.animate-stagger`)
- `prefers-reduced-motion` respect
- Custom scrollbar, custom selection color, global focus-visible outline
- Rank medal colors (gold/silver/bronze) on ranking list

**Verified:**
- `npx tsc --noEmit` → 0 errors
- `npx next build` → all 30+ pages compile
- Manual screenshot of `/en` and `/en/novel-finder` — renders cleanly with skeletons & new hero

## Prioritized Backlog (P0 → P2)

### P0 — Blocking / High Impact
- (none identified — original repo is production-ready per README)

### P1 — Nice-to-Have (deferred)
- Replace all `<img>` with `next/image` for automatic optimization (would break next.config domain whitelist without user config; deferred)
- Refactor `any` types in homepage → typed API responses (5 pre-existing lint errors)
- Wire "Recent Updates > Load More" to actual paginated API instead of navigation link
- Auto-carousel for HeroBanner across top 3 news items
- Real "Continue Reading" chapter-progress percentage (currently only visual affordance)

### P2 — Future Enhancements
- PWA manifest + service worker for offline chapter caching
- Push notifications for followed-novel updates
- Novel comparison view (side-by-side)
- Reading streak / gamification badges
- Advanced admin analytics dashboard

## Next Action Items (for user)
1. **Deploy**: `docker compose up -d --build` on your VPS — no code changes needed on Docker files
2. **Point DNS** to your server, update `FRONTEND_URL` in `.env` for CORS
3. **Seed database**: `docker compose exec backend go run ./cmd/seed/main.go` (creates admin/demo accounts per README)
4. (Optional) Consider running lint fixes on pre-existing `any` types when you next touch the homepage
