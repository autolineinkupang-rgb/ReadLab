# ReadLab

A full-stack **web novel platform** with novel discovery, reading, community features, ticket economy, XP system, and review system. Built with Go (Gin/GORM) + PostgreSQL on the backend and Next.js 16 (App Router) + Tailwind CSS 4 on the frontend.

## Features

- **Novel Finder** — Advanced search with 14 filters: text search, status, release status, genres (AND/OR), tags (AND/OR), excluded tags, minimum chapters/rating/reviews, and multiple sort options
- **Import from External APIs** — Search and import novels from NovelUpdates via the free [Consumet API](https://docs.consumet.org/)
- **Scraping** — Direct scraping from NovelUpdates, RoyalRoad, and NovelBin
- **Novel Management** — Admin panel to create, edit, and delete novels; review user requests
- **User Library** — Follow novels, reading history, bookmark progress
- **Voting & Requests** — Users can vote for novels and request new series
- **Rankings** — Daily/weekly/monthly leaderboards by views
- **Reviews** — Rate novels 1-5, write/edit reviews (max 5 edits), reply to reviews, gate bypass with tickets
- **Ticket Economy** — Virtual currency backed by `TicketUnit` model (SHA-256 serials); daily rewards, novel rewards, monthly leaderboard rewards; spend on review upgrades
- **Ticket Bank** — Spent tickets go to bank; admin can reclaim from bank via `/en/admin/bank`
- **Admin Config** — Real-time configurable ticket costs/rewards (`/en/admin/ticket-config`) and XP rewards (`/en/admin/xp-config`)
- **XP System** — Earn XP by reading chapters, voting, sharing, and writing reviews; configurable per-action amounts
- **User Profiles** — Public profiles with library, votes, requests; change password; AI translation settings
- **Reader** — Chapter-by-chapter reading with pagination, on-demand content fetching from novelfire.net with local caching
- **AI Translation** — Configure custom OpenRouter AI endpoint for per-chapter machine translation
- **News & Changelog** — Platform announcements and version history
- **Role System** — `member`, `writer`, `admin` roles with role-gated endpoints
- **Sidebar** — Collapsible navigation with state persisted in localStorage

## Tech Stack

| Layer | Technology |
|---|---|
| **Frontend** | Next.js 16 (App Router, Turbopack), React 19, TypeScript, Tailwind CSS 4 |
| **Backend** | Go 1.26+, Gin v1.12, GORM v1.31, JWT (golang-jwt v5) |
| **Database** | PostgreSQL 16 (Alpine) |
| **Auth** | JWT tokens stored in HTTP-only cookies |
| **Deployment** | Docker Compose (3 services: db, backend, frontend) + systemd fallback |

## Quick Start

### Prerequisites

- Docker & Docker Compose (recommended)
- Go 1.26+ and Node.js 22+ (for local development)

### Using Docker (Recommended)

```bash
# Clone and start all services
git clone <repo-url> readlab
cd readlab

# Build images
docker compose build

# Start stack
docker compose up -d

# Seed initial data (run once)
docker compose run --rm backend /seed

# Open in browser
open http://localhost:3000
```

### Local Development

```bash
# 1. Start PostgreSQL
docker compose up -d db

# 2. Copy environment config
cp .env.example .env
# Edit .env: set DB_HOST=localhost for local backend

# 3. Run backend
cd backend
cp .env.example .env
go run ./cmd/server/main.go

# 4. Run frontend (another terminal)
cd frontend
cp .env.example .env.local
npm install
npm run dev
```

### Demo Accounts (after seeding)

| Username | Email | Password | Role |
|---|---|---|---|
| `Mega_bells` | mega@example.com | password | Member |
| `reader1` | reader1@example.com | password | Member |
| `testuser` | test@test.com | password | Member |
| `admin2` | admin2@wtrlab.com | password | Admin |
| `superadmin` | superadmin@wtrlab.com | password | Admin |

**Default superadmin credentials for new installs** — login via `/en/admin` to access the admin dashboard.

## Project Structure

```
readlab/
├── docker-compose.yml         # Orchestrates db + backend + frontend
├── docker-compose.override.yml # Dev overrides (hot reload)
├── .env                       # Production/base environment config
├── .env.example               # Environment config template
├── backend/
│   ├── Dockerfile             # Multi-stage production build
│   ├── cmd/
│   │   ├── server/            # HTTP server + auto-migration entry point
│   │   └── seed/              # Database seeder with sample data
│   └── internal/
│       ├── config/            # Environment config loading
│       ├── handler/           # HTTP handlers (Gin controllers)
│       ├── importer/          # Consumet API import engine
│       ├── scraper/           # NovelUpdates/RoyalRoad/NovelBin scrapers
│       ├── lncrawl/           # LN Crawl integration
│       ├── middleware/        # Auth, CORS, logging, rate limiting
│       ├── model/             # GORM data models (15+ models)
│       ├── router/            # Route definitions
│       └── service/           # Business logic (tickets, XP, etc.)
└── frontend/
    ├── Dockerfile             # Multi-stage production build
    └── src/
        ├── app/               # Next.js App Router pages (40+ routes)
        │   └── en/            # English locale
        │       ├── admin/     # Users, Novels, Requests, Import, Chapters,
        │       │              # Reviews, News, Tickets, XP, Bank
        │       ├── novel/     # Novel detail + chapter reader
        │       └── ...
        ├── components/        # Reusable React components (Sidebar, Cards, etc.)
        └── lib/               # API client, AuthContext, navigation, SidebarContext
```

## Ticket Economy

ReadLab uses **TicketUnit** — each unit has a unique SHA-256 serial, an amount, and a status (`active` → `banked` → `spent`). This provides an auditable, tamper-proof ticket trail.

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

### Ticket Bank

- Spent tickets are moved to **bank** (`status = "banked"`)
- Admin can claim banked tickets via `/en/admin/bank`
- Bank balance shows total amount + unit count
- `migrateAllToBank` runs once (guarded by `bank_seeded` config flag)

### Admin Configuration

- **Ticket config** (`/en/admin/ticket-config`) — `daily_reward`, `novel_contribution`, `monthly_leaderboard`, `edit_reset_cost`, `gate_bypass_cost`, `replace_review_cost`
- **XP config** (`/en/admin/xp-config`) — `xp_read`, `xp_read_seconds`, `xp_vote`, `xp_share`, `xp_review`
- Changes take effect immediately via in-memory cache

## XP System

| Action | XP (default) | Condition |
|---|---|---|
| Read a chapter | 10 | Must read for minimum seconds (default 60s) |
| Vote for a novel | 3 | Per vote |
| Share a novel | 3 | Per share event |
| Write a review | 5 | Per approved review |

All amounts configurable at `/en/admin/xp-config`.

## Authentication & Authorization

### Roles
- `member` — Default role, can read, vote, request, review
- `writer` — Can create/edit/delete novels and chapters
- `admin` — Full access to user management, imports, scraping, ticket config, XP config, bank

### Auth Flow
1. Register/login via `POST /auth/register` or `POST /auth/login`
2. Server sets an HTTP-only cookie (`auth_token`) with JWT
3. Frontend reads `/auth/me` to get user profile and daily reward status
4. Logout via `POST /auth/logout` clears the cookie
5. Password change via `PUT /auth/password`

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
| `GET` | `/novels/:id/chapters/:num` | Chapter by number |
| `GET` | `/chapters/:id` | Single chapter |
| `GET` | `/search` | Search novels by title/author |
| `GET` | `/genres` | List all genres |
| `GET` | `/ranking/:period` | Ranking (daily/weekly/monthly) |
| `GET` | `/updates` | Recent chapter updates |
| `GET` | `/news` | News list |
| `GET` | `/stats` | Platform statistics |
| `GET` | `/leaderboard` | Top users by tickets |
| `GET` | `/profile/:id` | User public profile |
| `GET` | `/config/upgrade-costs` | Current ticket upgrade costs |
| `POST` | `/auth/register` | Register new user |
| `POST` | `/auth/login` | Login |
| `POST` | `/auth/logout` | Logout |

### Protected Endpoints (auth required)

| Method | Path | Description |
|---|---|---|
| `GET` | `/auth/me` | Current user profile + daily reward status |
| `PUT` | `/auth/password` | Change password |
| `POST` | `/votes` | Vote for a novel |
| `POST` | `/requests` | Submit novel request |
| `GET` | `/requests` | List user requests |
| `GET` | `/library` | User's follows + history |
| `POST` | `/novels/:id/reviews` | Create a review |
| `PUT` | `/novels/:id/reviews/:reviewId` | Update/replace a review |
| `POST` | `/novels/:id/chapters/:num/read` | Track chapter read |
| `POST` | `/novels/:id/chapters/:num/xp` | Claim XP for reading |
| `POST` | `/rewards/daily` | Claim daily reward |
| `GET` | `/rewards/status` | Daily reward claim status |

### Admin Endpoints (admin only)

| Method | Path | Description |
|---|---|---|
| `GET` | `/admin/users` | List users |
| `GET` | `/admin/users/:id` | Get user |
| `PUT` | `/admin/users/:id` | Update user |
| `DELETE` | `/admin/users/:id` | Delete user |
| `POST` | `/admin/users/:id/tickets` | Send tickets to user |
| `GET` | `/admin/bank` | Bank balance |
| `POST` | `/admin/bank/claim` | Claim tickets from bank |
| `GET` | `/admin/stats` | Platform stats |
| `GET` | `/admin/config/tickets` | List ticket configs |
| `PUT` | `/admin/config/tickets` | Update ticket config |
| `GET` | `/admin/config/xp` | List XP configs |
| `PUT` | `/admin/config/xp` | Update XP config |
| `POST` | `/admin/rewards/monthly` | Distribute monthly XP rewards |

## Novel Finder

The **Novel Finder** (`/en/novel-finder`) provides 14 filter controls with data fetched from the API.

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

## Database Backups

```bash
# Manual backup
docker compose --profile backup up -d

# Scheduled backup (cron)
0 3 * * * cd /path/to/ReadLab && ./scripts/db-backup.sh
```

Backups stored in Docker volume `pgbackups` with 7-day retention.

## License

MIT
