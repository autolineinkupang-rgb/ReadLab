# ReadLab

A full-stack **web novel platform** with novel discovery, reading, community features, ticket economy, and review system. Built with Go (Gin/GORM) + PostgreSQL on the backend and Next.js 16 (App Router) + Tailwind CSS 4 on the frontend.

## Features

- **Novel Finder** — Advanced search with 14 filters: text search, status, release status, genres (AND/OR), tags (AND/OR), excluded tags, minimum chapters/rating/reviews, and multiple sort options
- **Import from External APIs** — Search and import novels from NovelUpdates via the free [Consumet API](https://docs.consumet.org/)
- **Scraping** — Direct scraping from NovelUpdates, RoyalRoad, and NovelBin
- **Novel Management** — Admin panel to create, edit, and delete novels; review user requests
- **User Library** — Follow novels, reading history, bookmark progress
- **Voting & Requests** — Users can vote for novels and request new series
- **Rankings** — Daily/weekly/monthly leaderboards by views
- **Reviews** — Rate novels 1-5, write/edit reviews (max 5 edits), reply to reviews, gate bypass with tickets
- **Ticket Economy** — Virtual currency with daily rewards, novel creation rewards, monthly XP leaderboard rewards; spend tickets on review upgrades (edit reset, gate bypass, replace review)
- **Admin Ticket Config** — Admin can configure all ticket costs/rewards in real-time via a database-backed settings panel
- **User Profiles** — Public profiles with library, votes, requests; change password; AI translation settings
- **Reader** — Chapter-by-chapter reading with pagination
- **AI Translation** — Configure custom OpenRouter AI endpoint for per-chapter machine translation
- **News & Changelog** — Platform announcements and version history
- **Role System** — `member`, `writer`, `admin` roles with role-gated endpoints

## Tech Stack

| Layer | Technology |
|---|---|
| **Frontend** | Next.js 16 (App Router, Turbopack), React 19, TypeScript, Tailwind CSS 4 |
| **Backend** | Go 1.25+, Gin v1.12, GORM v1.31, JWT (golang-jwt v5) |
| **Database** | PostgreSQL 16 (Alpine) |
| **Auth** | JWT tokens stored in HTTP-only cookies |
| **Deployment** | Docker Compose (3 services: db, backend, frontend) |

## Quick Start

### Prerequisites

- Docker & Docker Compose
- Go 1.25+ (for local development)
- Node.js 22+ (for local development)

### Using Docker (Recommended)

```bash
# Clone and start all services
git clone <repo-url> readlab
cd readlab
docker compose up -d

# Open in browser
open http://localhost:3000
```

### Local Development

```bash
# 1. Start PostgreSQL
docker compose up -d db

# 2. Copy environment config
cp .env.example .env
# Edit .env if needed (DB_HOST=localhost for local backend)

# 3. Run backend
cd backend
cp .env.example .env        # adjust if needed
go run ./cmd/server/main.go

# 4. Run frontend (in another terminal)
cd frontend
cp .env.example .env.local
npm install
npm run dev
```

### Demo Accounts (after seeding)

| Username | Email | Password | Role |
|---|---|---|---|
| `admin` | your-email@example.com | admin123 | Admin |
| `Mega_bells` | mega@example.com | password | User |
| `reader1` | reader1@example.com | password | User |

## Project Structure

```
readlab/
├── docker-compose.yml         # Orchestrates db + backend + frontend
├── .env.example               # Environment config template
├── backend/
│   ├── cmd/
│   │   ├── server/            # HTTP server + auto-migration entry point
│   │   └── seed/              # Database seeder with sample data
│   ├── internal/
│   │   ├── config/            # Environment config loading
│   │   ├── handler/           # HTTP handlers (Gin controllers)
│   │   ├── importer/          # Consumet API import engine
│   │   ├── scraper/           # NovelUpdates/RoyalRoad/NovelBin scrapers
│   │   ├── lncrawl/           # LN Crawl integration
│   │   ├── middleware/        # Auth, CORS, logging, rate limiting
│   │   ├── model/             # GORM data models (13 models)
│   │   └── router/            # Route definitions
│   └── migrations/            # Reference SQL migrations
└── frontend/
    ├── src/
    │   ├── app/               # Next.js App Router pages
    │   │   ├── en/            # English locale routes (40+ pages)
    │   │   │   ├── admin/     # Admin panel (users, novels, import, reviews, news, tickets)
    │   │   │   ├── novel/     # Novel detail, chapter reader
    │   │   │   ├── novel-finder/  # Advanced search page
    │   │   │   └── ...        # Other pages
    │   ├── components/        # Reusable React components
    │   ├── lib/               # API client, AuthContext, utilities, navigation
    │   └── types.ts           # TypeScript interfaces
    └── public/                # Static assets
```

## Ticket Economy

ReadLab has a virtual currency system (Tickets) for rewarding users and enabling premium actions.

### Reward Sources

| Reward | Amount (default) | Trigger |
|---|---|---|
| Daily Login | 2 tickets | `POST /rewards/daily` (once per day, Asia/Makassar timezone) |
| Novel Contribution | 100 tickets | Creating a new novel as a writer |
| Monthly XP Leaderboard | 50 tickets | `POST /admin/rewards/monthly` (admin-triggered, resets XP) |

### Spend Actions

| Action | Cost (default) | Description |
|---|---|---|
| Edit Reset | 20 tickets | Reset review edit count (max 5) to continue editing |
| Gate Bypass | 50 tickets | Skip the "read 5 chapters" review gate requirement |
| Replace Review | 100 tickets | Replace an existing review with a new one |

### Admin Configuration

All ticket costs and rewards can be changed in real-time at `/en/admin/ticket-config`. Changes take effect immediately via an in-memory cache that reloads on save.

| Key | Default | Purpose |
|---|---|---|
| `daily_reward` | 2 | Daily login reward amount |
| `novel_contribution` | 100 | Reward for creating a novel |
| `monthly_leaderboard` | 50 | Monthly XP leaderboard reward |
| `edit_reset_cost` | 20 | Cost to reset review edit limit |
| `gate_bypass_cost` | 50 | Cost to bypass review gate |
| `replace_review_cost` | 100 | Cost to replace existing review |

## Review System

- Users can rate novels 1-5 and write reviews
- **Edit limit**: 5 edits per review; after that, must pay tickets to reset
- **Review gate**: Must read 5 chapters before reviewing (bypass available)
- **Duplicate check**: One review per user per novel; replace option available
- **Replies**: Users can reply to reviews (nested)
- **Upgrade flow**: Backend returns `upgrade_cost` and `upgrade_type` in 403 errors; frontend shows a payment dialog dynamically

## Authentication & Authorization

### Roles
- `member` — Default role, can read, vote, request, review
- `writer` — Can create/edit/delete novels and chapters
- `admin` — Full access to user management, imports, scraping, ticket config, news

### Auth Flow
1. Register/login via `POST /auth/register` or `POST /auth/login`
2. Server sets an HTTP-only cookie (`auth_token`) with JWT
3. Frontend reads `/auth/me` to get user profile and daily reward status
4. Logout via `POST /auth/logout` clears the cookie
5. Password change via `PUT /auth/password` (validates current password + rules)

## API Overview

All API routes prefixed with `/api/v1`. See [docs/API.md](docs/API.md) for full documentation.

### Public Endpoints

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | Health check |
| `GET` | `/novels` | List novels (filtered, sorted, paginated) |
| `GET` | `/novels/trending` | Top 20 by views |
| `GET` | `/novels/recommendations` | Top 12 by rating |
| `GET` | `/novels/random` | Random novels |
| `GET` | `/novels/:id` | Novel detail with genres |
| `GET` | `/novels/:id/chapters` | Chapter list for a novel |
| `GET` | `/novels/:id/chapters/:num` | Chapter by number (optional auth for locked) |
| `GET` | `/novels/import/search` | Search external API (Consumet) |
| `GET` | `/chapters/:id` | Single chapter |
| `GET` | `/search` | Search novels by title/author |
| `GET` | `/genres` | List all genres |
| `GET` | `/ranking/:period` | Ranking (daily/weekly/monthly) |
| `GET` | `/updates` | Recent chapter updates |
| `GET` | `/news` | News list |
| `GET` | `/stats` | Platform statistics |
| `GET` | `/leaderboard` | Top users by tickets |
| `GET` | `/author/:name/novels` | Novels by author |
| `GET` | `/profile/:id` | User public profile |
| `GET` | `/config/upgrade-costs` | Current ticket upgrade costs |
| `POST` | `/auth/register` | Register new user |
| `POST` | `/auth/login` | Login |
| `POST` | `/auth/logout` | Logout |

### Protected Endpoints (auth required)

| Method | Path | Description |
|---|---|---|
| `GET` | `/auth/me` | Current user profile + daily reward status |
| `PUT` | `/auth/password` | Change password (needs current password) |
| `POST` | `/votes` | Vote for a novel |
| `POST` | `/requests` | Submit novel request |
| `GET` | `/requests` | List user requests |
| `GET` | `/library` | User's follows + history |
| `POST` | `/novels/:id/reviews` | Create a review |
| `PUT` | `/novels/:id/reviews/:reviewId` | Update/replace a review |
| `GET` | `/novels/:id/reviews` | List reviews for a novel |
| `POST` | `/novels/:id/chapters/:num/read` | Track chapter read |
| `POST` | `/novels/:id/chapters/:num/xp` | Claim XP for reading |
| `GET` | `/novels/:id/my-progress` | User's reading progress |
| `POST` | `/novels/:id/share` | Share a novel (earn XP) |
| `POST` | `/rewards/daily` | Claim daily reward |
| `GET` | `/rewards/status` | Daily reward claim status |
| `GET` | `/user/ai-settings` | Get AI translation settings |
| `PUT` | `/user/ai-settings` | Update AI translation settings |
| `POST` | `/translate/ai` | AI translate chapter |

### Writer Endpoints (writer or admin)

| Method | Path | Description |
|---|---|---|
| `POST` | `/novels` | Create novel |
| `PUT` | `/novels/:id` | Update novel |
| `DELETE` | `/novels/:id` | Delete novel |
| `POST` | `/admin/novels/:id/chapters` | Create chapter |
| `PUT` | `/admin/novels/:id/chapters/:chapterID` | Update chapter |
| `GET` | `/admin/novels/:id/chapters` | List admin chapters |
| `GET` | `/admin/chapters/:id` | Get chapter for editing |

### Admin Endpoints (admin only)

| Method | Path | Description |
|---|---|---|
| `GET` | `/admin/users` | List users |
| `GET` | `/admin/users/:id` | Get user |
| `PUT` | `/admin/users/:id` | Update user |
| `DELETE` | `/admin/users/:id` | Delete user |
| `POST` | `/admin/users/admin` | Create admin user |
| `GET` | `/admin/stats` | Platform stats |
| `GET` | `/admin/reviews` | List all reviews |
| `DELETE` | `/admin/reviews/:id` | Delete a review |
| `GET` | `/admin/requests` | List all requests |
| `PUT` | `/requests/:id` | Approve/reject request |
| `POST` | `/admin/news` | Create news article |
| `PUT` | `/admin/news/:id` | Update news |
| `DELETE` | `/admin/news/:id` | Delete news |
| `POST` | `/admin/rewards/monthly` | Distribute monthly XP rewards |
| `GET` | `/admin/config/tickets` | List ticket configs |
| `PUT` | `/admin/config/tickets` | Update ticket config value |
| `POST` | `/novels/import` | Import from Consumet |
| `POST` | `/novels/scrape` | Scrape novel metadata |
| `POST` | `/novels/scrape/import` | Scrape and import chapters |
| `POST` | `/novels/lncrawl` | Crawl chapters via LN Crawl |

## Novel Finder

The **Novel Finder** (`/en/novel-finder`) provides 14 filter controls with data fetched from the API (including dynamic genre list).

## Importing Novels

### Via Admin Panel

1. Navigate to `/en/admin/import`
2. Search by novel name
3. Select a result and click **Import**
4. Optionally check "Also import chapter list"

### Via CLI (LN Crawl)

```bash
curl -X POST http://localhost:8080/api/v1/novels/lncrawl \
  -H "Content-Type: application/json" \
  -H "Cookie: auth_token=..." \
  -d '{"novel_url": "https://novelbin.com/b/novel-title"}'
```

## License

MIT
