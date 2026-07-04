# WTR-LAB Clone

A full-stack **web novel platform** with novel discovery, reading, and community features. Built with Go (Gin/GORM) + PostgreSQL on the backend and Next.js 16 (App Router) + Tailwind CSS 4 on the frontend.

## Features

- **Novel Finder** — Advanced search with 14 filters: text search, status, release status, genres (AND/OR), tags (AND/OR), excluded tags, minimum chapters/rating/reviews, and multiple sort options
- **Import from External APIs** — Search and import novels from NovelUpdates via the free [Consumet API](https://docs.consumet.org/)
- **Novel Management** — Admin panel to create, edit, and delete novels; review user requests
- **User Library** — Follow novels, reading history, bookmark progress
- **Voting & Requests** — Users can vote for novels and request new series
- **Rankings** — Daily/weekly/monthly leaderboards by views
- **User Profiles** — Public profiles with library, votes, and requests
- **Reader** — Chapter-by-chapter reading with pagination
- **News & Changelog** — Platform announcements and version history
- **Tickets System** — Virtual currency for unlocking premium chapters

## Tech Stack

| Layer | Technology |
|---|---|
| **Frontend** | Next.js 16 (App Router), React 19, TypeScript, Tailwind CSS 4 |
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
git clone <repo-url> wtr-lab-clone
cd wtr-lab-clone
docker compose up -d

# Seed the database with sample data
docker compose exec backend ./seed

# Open in browser
open http://localhost:3000
```

### Local Development

```bash
# 1. Start PostgreSQL
docker compose up -d db

# 2. Copy environment config
cp .env.example .env

# 3. Run backend
cd backend
cp .env.example .env        # adjust if needed
go run ./cmd/server/main.go

# 4. Seed data (optional)
go run ./cmd/seed/main.go

# 5. Run frontend (in another terminal)
cd frontend
cp .env.example .env.local
npm install
npm run dev
```

### Demo Accounts (after seeding)

| Username | Email | Password | Role |
|---|---|---|---|
| `admin` | admin@wtrlab.com | admin123 | Admin |
| `Mega_bells` | mega@example.com | password | User |
| `reader1` | reader1@example.com | password | User |

## Project Structure

```
wtr-lab-clone/
├── docker-compose.yml         # Orchestrates db + backend + frontend
├── .env.example               # Environment config template
├── backend/
│   ├── cmd/
│   │   ├── server/            # HTTP server entry point
│   │   └── seed/              # Database seeder
│   ├── internal/
│   │   ├── config/            # Environment config loading
│   │   ├── handler/           # HTTP handlers (Gin controllers)
│   │   ├── importer/          # External API import engine
│   │   ├── middleware/        # Auth, CORS, logging middleware
│   │   ├── model/             # GORM data models
│   │   └── router/            # Route definitions
│   └── migrations/            # Reference SQL migrations
└── frontend/
    ├── src/
    │   ├── app/               # Next.js App Router pages
    │   │   ├── en/            # English locale routes
    │   │   │   ├── admin/     # Admin panel (import, requests, novels)
    │   │   │   ├── novel/     # Novel detail & chapter reader
    │   │   │   ├── novel-finder/  # Advanced search page
    │   │   │   ├── novel-list/    # Simple listing with filters
    │   │   │   └── ...        # Other pages
    │   ├── components/        # Reusable React components
    │   └── lib/               # API client, utilities
    └── public/                # Static assets
```

## API Overview

All API routes are prefixed with `/api/v1`. See [docs/API.md](docs/API.md) for full documentation.

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
| `POST` | `/auth/register` | Register new user |
| `POST` | `/auth/login` | Login |
| `POST` | `/auth/logout` | Logout |

### Protected Endpoints (auth required)

| Method | Path | Description |
|---|---|---|
| `GET` | `/auth/me` | Current user profile |
| `POST` | `/votes` | Vote for a novel |
| `POST` | `/requests` | Submit novel request |
| `PUT` | `/requests/:id` | Approve/reject request |
| `GET` | `/library` | User's follows + history |
| `POST` | `/novels` | Create novel manually |
| `PUT` | `/novels/:id` | Update novel |
| `DELETE` | `/novels/:id` | Delete novel |
| `POST` | `/novels/import` | Import novel from external API |

## Novel Finder

The **Novel Finder** (`/en/novel-finder`) provides 14 filter controls:

- **Search** — Title, raw title, and description (with "description only" toggle)
- **Sort** — Addition Date, Rating, Chapters, Views, Title, Readers, Reviews
- **Order** — Descending / Ascending
- **Status** — All, Ongoing, Completed, Hiatus, Dropped
- **Release Status** — All, Released, Voting
- **Minimums** — Chapters (100+/500+/1000+/2000+), Rating (3.0+/3.5+/4.0+/4.5+), Reviews (50+/100+/500+/1000+)
- **Genres** — 2-column checkboxes with AND/OR matching mode
- **Tags** — Typeahead multi-select (5 categories) with AND/OR mode
- **Excluded Tags** — Exclude novels containing specific tags

Data is fetched from the backend API with client-side fallback filtering.

## Importing Novels

Novels can be imported from [NovelUpdates](https://www.novelupdates.com/) via the free [Consumet API](https://docs.consumet.org/).

### Via Admin Panel

1. Navigate to `/en/admin/import`
2. Search by novel name
3. Select a result and click **Import**
4. Optionally check "Also import chapter list"

### Via API

```bash
# Search
curl "http://localhost:8080/api/v1/novels/import/search?q=solo+leveling"

# Import
curl -X POST http://localhost:8080/api/v1/novels/import \
  -H "Content-Type: application/json" \
  -H "Cookie: auth_token=<token>" \
  -d '{"source_id": "solo-leveling", "with_chapters": true}'
```

## Novel List Filter Parameters

`GET /api/v1/novels` supports the following query parameters:

| Param | Type | Example | Description |
|---|---|---|---|
| `page` | int | `1` | Page number |
| `limit` | int | `20` | Items per page (max 100) |
| `q` | string | `solo` | Text search (title, alt title, description) |
| `status` | string | `ongoing` | Filter by status |
| `genres` | string | `action,fantasy` | Comma-separated genre slugs |
| `genre_mode` | string | `and` / `or` | Genre matching mode |
| `min_chapters` | int | `500` | Minimum chapters |
| `min_rating` | float | `4.0` | Minimum rating |
| `min_reviews` | int | `100` | Minimum review count |
| `sort` | string | `rating` | Sort field (see below) |
| `order` | string | `desc` / `asc` | Sort order |

**Sort options:** `created_at`, `title`, `views`, `chapters`, `rating`, `readers`, `reviews`

## License

MIT
