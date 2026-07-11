# Changelog & Build Specification

## Project Overview

Full-stack web novel platform (ReadLab) with:
- **Backend**: Go + Gin + GORM + SQLite
- **Frontend**: Next.js 16 + React 19 + Tailwind CSS 4 + Turbopack
- **Features**: Novel discovery, external API import (Consumet), admin management, responsive UI (sidebar desktop / navbar tablet / overlay mobile), reusable ChapterReader

---


### v1.1.0.1 — Repo hygiene cleanup
**Date:** 2026-07-11

**Fixed:**
- Removed `.playwright-mcp/` debug artifacts from git tracking (added to `.gitignore`)
- Removed `.gitconfig` and `.emergent/` agent-platform metadata from git tracking
- Restored `.env.example` and `backend/.env.example` as setup documentation (placeholder values only, no real secrets)

### v1.1.0 — UI/UX Polish & Bug Fixes
**Date:** 2026-01-11

**Fixed:**
- **NovelCard** now actually renders the cover image passed via the `image` prop (was previously declared but unused — always showed placeholder icon)
- **NovelCardSmall** now shows an optional thumbnail alongside the ranking
- **UpdateItem** now displays actual cover thumbnails when available (was showing generic book icon)
- **(main)/layout.tsx** — fixed flex parent chain so `flex-1`/`min-h-screen` behaves correctly with sidebar-off state; footer now consistently sticks to bottom
- **HeroBanner** — replaced hardcoded "16th Giveaway" banner with a dynamic hero pulling the latest news item; falls back to a "Welcome to ReadLab" card when no news exists; dismissible per-item via localStorage
- Reading progress bar (0–100 %) can now be rendered on any `NovelCard` via the new `progress` prop (enables "Continue Reading" cards to show completion visually)

**Added:**
- **`components/Skeleton.tsx`** — `Skeleton`, `NovelCardSkeleton`, `NovelCardSkeletonRow`, `UpdateItemSkeleton` for consistent loading states across the app
- **`components/BackToTop.tsx`** — floating "back to top" pill button; appears after 400 px scroll; smooth-scrolls with animation
- **`components/HeroBanner.tsx`** — dismissible dynamic hero banner (uses latest news or fallback welcome)
- Homepage now renders skeleton loaders while data loads, and empty-state cards (with icon) when a section returns no data
- Homepage sections use `.animate-stagger` for polished staggered entrance
- Rank badge colors on ranking list: gold (#1), silver (#2), bronze (#3)
- Every interactive novel card / small card / update item has a stable `data-testid` for QA
- Focus-visible outlines on all navigation cards (accessibility)

**Changed:**
- **`globals.css`** — added `fadeIn`, `fadeInUp`, `animate-stagger` (8-child staggered entrance), `prefers-reduced-motion` overrides, custom scrollbar styling, text selection color, global focus-visible ring
- **`Footer.tsx`** — copyright year is now dynamic (`new Date().getFullYear()`); version bumped to v1.1.0; added tagline
- **Recent Updates** section: "Load More" button is now a real link to `/en/novel-list?sort=updated_at`
- **Top Spenders** — added inline rank badge (#1/#2/#3) and 🎫 emoji for ticket count

**Files added:**
- `frontend/src/components/BackToTop.tsx`
- `frontend/src/components/HeroBanner.tsx`
- `frontend/src/components/Skeleton.tsx`

**Files modified:**
- `frontend/src/app/globals.css`
- `frontend/src/app/en/(main)/layout.tsx`
- `frontend/src/app/en/(main)/page.tsx`
- `frontend/src/components/NovelCard.tsx`
- `frontend/src/components/NovelCardSmall.tsx`
- `frontend/src/components/UpdateItem.tsx`
- `frontend/src/components/Footer.tsx`
- `frontend/src/types.ts` (added `LatestNewsItem`)

**Verification:**
- `npx tsc --noEmit` → 0 errors
- `npx next build` → all 30+ pages compile successfully
- ESLint: no new errors introduced (baseline of 9 pre-existing `any` errors from original code preserved)

---

## Build History

### Commit `f20ee9a` — Initial Project Scaffold
**Date:** 2026-07-03

**Added:**
- Backend scaffold (Go): models, handlers, migrations, router, middleware, seed data, config, Dockerfile, docker-compose
  - 10 SQL migration files (genres, novels, chapters, users, votes, requests, tickets, news, reading_history, novel_follows)
  - 16 handlers: auth, author, chapter, genre, health, leaderboard, library, news, novel, ranking, request, search, stats, update, user, vote
  - Middleware: JWT auth, CORS, request logging
  - Models for all entities
- Frontend scaffold (Next.js): 30+ pages, components, API client, config
  - Pages: novel-finder, novel-list, novel detail, chapter reader, ranking, leaderboard, library, login, news, author, profile, vote, request-serie, trending, recommendation, random-novels, public-stats, about-us, contact-us, DMCA, privacy-policy, terms-of-use, cookie-policy
  - Components: Navbar, Footer, NovelCard, NovelCardSmall, SectionHeader, UpdateItem
  - API client library with all CRUD operations

---

### Commit `7112c80` — Visual Redesign (WTR-LAB Theme → ReadLab)
**Date:** 2026-07-04
**Message:** "visual redesign for ReadLab rebranding"

**Added:**
- Navbar search bar with `useRouter` navigation to `/en/novel-finder?q=...`
- Google Fonts (Nunito Sans, JetBrains Mono) via `<link>` tag
- Full favicon set (SVG, PNG, ICO, apple-touch-icon, webmanifest)
- OpenGraph metadata in root layout
- `q` query param support in `novels.list()` API call
- `.hide` CSS utility for screen-reader-only content
- Antialiased font rendering CSS
- CSS custom properties (`--primary`, `--primary-dark`)

**Changed:**
- Color scheme from violet/purple (`#7c3aed`) to teal/cyan (`#2193b0` / `#6dd5ed`) across ALL components
- Gradient backgrounds from violet palette to teal/blue (`#0d1b2a`)
- Navbar logo gradient from violet-purple to teal-cyan
- Card hover shadow from violet glow to teal glow
- Font family to "Nunito Sans" with system fallback
- Homepage content: all novel lists, covers, featured novel, community section — replaced with Douluo-themed data and updated routes
- All route paths to include `/en/` prefix
- Button colors to teal throughout

**Deleted:**
- Trending covers grid (12-item carousel) — replaced with single featured novel card
- `MOCK_RECOMMENDATIONS` usage — replaced with `MOCK_NEW.slice(0, 6)`

---

### Commit `0ca1f1e` — Enhanced Novel Finder
**Date:** 2026-07-04
**Message:** "fitur finder diperluas"

**Added:**
- Description-only search toggle checkbox
- Release status filter (all, ongoing, completed, hiatus, dropped)
- Min chapters, min rating, min reviews filter dropdowns
- Genre match mode toggle (AND / OR)
- Tag system with include/exclude tags, 5 categories, searchable autocomplete dropdown
- Backend filter params: `genres`, `genre_mode`, `min_chapters`, `min_rating`, `min_reviews`
- `clientSideFilter()` function for comprehensive client-side filtering
- "Clear All" button to reset all filters
- Active filter count badge on "Show Filters" button
- Pagination with ellipsis for large page ranges
- Empty results state with icon and suggestion text
- `Author` field to Novel interface in novel-list

**Changed:**
- Search bar design (input-with-button → styled "Search" prefix + transparent input)
- Results display (horizontal card → vertical cover grid with hover zoom)
- Sort options expanded: added `created_at`, `reviews`; logic uses `AddedMinutesAgo`
- Results count text format
- `sort` type narrowed from `string` to `SortField` union type

---

### Commit `0106672` — Admin Panel, Importer, Documentation
**Date:** 2026-07-04
**Message:** "menambqhkan deskripsi dan readme.md"

**Added:**
- `docs/API.md` — 493-line API reference with curl examples, request/response schemas
- Backend Importer package (`backend/internal/importer/importer.go`) — Consumet API client, auto-slug generation, genre matching, transaction-based novel + chapters creation
- `POST /novels/import` and `GET /novels/import/search` endpoints
- Novel CRUD endpoints: `POST /novels`, `PUT /novels/:id`, `DELETE /novels/:id`
- Request review endpoint: `PUT /requests/:id` (approve/reject)
- Frontend Admin Import page — search Consumet API, import with optional chapters
- Frontend Admin Manage Novels page — list, inline edit, delete with confirmation
- Frontend Admin Review Requests page — approve/reject with mock data fallback
- Admin dropdown in Navbar (Import / Requests / Manage links)
- API client functions: `importer.search()`, `importer.import()`, `adminNovels.*`, `adminRequests.review()`
- `generateSlug()` helper in Go
- README with features, tech stack, quick start, project structure

**Changed:**
- Router registered 6 new routes (importer search/import, novel CRUD, request review)
- `novel_handler.go` — added CRUD methods alongside existing list/trending/random

---

## Uncommitted Changes (Current Session)

These changes have been made but NOT yet committed to git:

### Added
- **`frontend/src/app/en/(main)/` route group** — all existing pages moved under `(main)/` group so they share Navbar + Sidebar + Footer layout
- **`frontend/src/app/en/(main)/layout.tsx`** — new route group layout rendering Navbar, Sidebar, main content, Footer
- **`frontend/src/components/Sidebar.tsx`** — fixed desktop sidebar (w-64, `hidden lg:flex`), 4 nav sections (Browse, Discover, Community, Admin), search input, Login button
- **`frontend/src/components/ChapterReader.tsx`** — standalone reusable chapter reader component:
  - Loading skeleton with deterministic widths (hydration-safe)
  - Error state with icon and back-to-novel link
  - Reader toolbar: source tabs (Web/Web+/AI), Prev/Progress/Next nav bar
  - Display settings panel: font size (14-28), font family (Sans/Serif/Mono), theme (dark/light/sepia)
  - Fixed bottom toolbar with 5 buttons (Read, Display, Speech, Settings, More)
  - Dark/light/sepia themes with full style application (backgrounds, text, borders)
  - Progress bar showing chapter completion %

### Changed
- **`frontend/src/app/layout.tsx`** — removed Navbar and Footer from root layout (moved to `(main)/` route group); simplified `<body>` to just `{children}`
- **`frontend/src/app/globals.css`** — `.nav-gradient` changed from gradient-to-transparent to solid `#12122a` + bottom border
- **`frontend/src/components/Navbar.tsx`** — major rewrite:
  - Added `usePathname()` for active link highlighting
  - Three responsive tiers: `hidden md:flex lg:hidden` tablet nav, `md:hidden` mobile hamburger menu
  - Tablet nav: Browse shortcuts (Finder, Novels, Ranking, Top, Library) + More dropdown (Discover + Community links) + Admin dropdown
  - Mobile overlay menu: `fixed inset-0 top-16 z-40 overflow-y-auto` with Browse/Discover/Community/Admin sections
  - Body scroll lock via `useEffect` when mobile menu opens
  - `href` objects used for active state tracking
  - Added `More` dropdown with Discover + Community links
  - Added `Admin` dropdown for Import/Requests/Manage links
- **`frontend/src/components/Footer.tsx`** — added missing menu links: Trending, Recommendations, News, Request Series, Vote Series
- **`frontend/src/components/NovelCard.tsx`** — wrapped `{title}` in `<span>` to fix hydration error (`<a>` inside block element)
- **`frontend/src/app/en/novel/[id]/[slug]/chapter-[num]/page.tsx`** — simplified from ~190 lines to ~75 lines:
  - Removed inline reader UI (moved to ChapterReader component)
  - Removed `fontSize` local state (now in ChapterReader)
  - Now uses `ChapterReader` component with props only
  - Generates mock chapter content using `generateMockContent()`

### Deleted
- **`frontend/src/app/en/page.tsx`** — moved to `(main)/page.tsx`
- All 27 pages under `frontend/src/app/en/` without route group — moved to `frontend/src/app/en/(main)/`:
  - about-us, admin/import, admin/novels, admin/requests, author/[name], contact-us, cookie-policy, dmca, leaderboard, library, login, news, news/[id], novel-finder, novel-list, novel/[id]/[slug], privacy-policy, profile/[id], profile/request-serie, profile/vote-serie, public-stats, random-novels, ranking/[period], recommendation, terms-of-use, trending

---

## Architecture Summary

```
frontend/src/
├── app/
│   ├── layout.tsx              # Root layout (html, head, body with gradient-bg)
│   ├── globals.css              # Tailwind v4 + theme vars + custom classes
│   ├── page.tsx                 # Root redirect (/ → /en)
│   ├── en/
│   │   ├── (main)/              # Route group: all pages with sidebar/navbar
│   │   │   ├── layout.tsx       # Navbar + Sidebar + main + Footer
│   │   │   ├── page.tsx         # Homepage
│   │   │   ├── novel-finder/    # Enhanced finder with 14 filters
│   │   │   ├── admin/           # Import, Requests, Manage
│   │   │   └── ...27 pages
│   │   └── novel/[id]/[slug]/
│   │       └── chapter-[num]/   # Reader page (outside (main), no sidebar)
│   └── favicon.ico
├── components/
│   ├── Navbar.tsx               # Responsive nav (tablet navbar, mobile overlay)
│   ├── Sidebar.tsx              # Fixed desktop sidebar
│   ├── Footer.tsx               # Site footer
│   ├── ChapterReader.tsx        # Reusable reader component
│   ├── NovelCard.tsx            # Novel card grid item
│   ├── NovelCardSmall.tsx       # Compact novel card
│   ├── SectionHeader.tsx        # Section header with optional "See More"
│   └── UpdateItem.tsx           # Update timeline item
└── lib/
    └── api.ts                   # API client (CRUD, importer, admin)

backend/
├── cmd/server/main.go           # Server entry
├── cmd/seed/main.go             # Seed data
├── internal/
│   ├── config/config.go
│   ├── handler/                 # 16+ HTTP handlers
│   │   ├── novel_handler.go     # List, CRUD, trending, random
│   │   ├── importer_handler.go  # Search + import from Consumet
│   │   ├── request_handler.go   # Create + review requests
│   │   └── ...
│   ├── importer/importer.go     # Consumet API client
│   ├── middleware/              # Auth, CORS, Logger
│   ├── model/                   # All DB models
│   └── router/router.go        # Route registration
└── migrations/                  # 10 SQL migration files
```

## Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| Consumet API (not NovelUpdate direct) | NovelUpdate has no official REST API; Consumet is free, open-source, no API key |
| Route group `(main)/` for sidebar exclusion | Avoids hydration errors from `usePathname()` runtime branching in client components |
| ChapterReader as pure presentational component | No data fetching; props-driven; reusable from any page |
| Inline SVGs (no lucide-react) | Minimize dependencies, no external icon library |
| Solid `#12122a` navbar background | Replaced fading gradient with distinct container separation via shadow |
| Mobile overlay with `overflow-y-auto` | Body scroll lock doesn't trap the menu content |
| Hardcoded loading skeleton widths | `Math.random()` causes hydration mismatch |
| `@theme inline` in Tailwind v4 | Defines `--color-background`/`--color-foreground` for Tailwind utilities |
