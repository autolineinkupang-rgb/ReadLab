# API Documentation

Base URL: `http://localhost:8080/api/v1`

**Authentication:** JWT tokens are sent via HTTP-only cookie (`auth_token`) or `Authorization: Bearer <token>` header.

---

## Health

### `GET /health`

Check server and database health.

```bash
curl http://localhost:8080/api/v1/health
```

Response:
```json
{
  "status": "ok",
  "database": "connected"
}
```

---

## Authentication

### `POST /auth/register`

Create a new user account.

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username": "newuser", "email": "user@example.com", "password": "secret123"}'
```

Response `201`:
```json
{
  "user": {
    "ID": 6,
    "Username": "newuser",
    "Email": "user@example.com",
    "DisplayName": "newuser",
    "Tickets": 0,
    "IsAdmin": false,
    "CreatedAt": "2026-07-04T..."
  },
  "token": "eyJhbGciOi..."
}
```

### `POST /auth/login`

Authenticate and receive a JWT token (set as cookie + returned in body).

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@example.com", "password": "admin123"}'
```

Response:
```json
{
  "user": { "...": "..." },
  "token": "eyJhbGciOi..."
}
```

### `POST /auth/logout`

Clear the auth cookie.

### `GET /auth/me`

Get the currently authenticated user's profile. **Auth required.**

```bash
curl http://localhost:8080/api/v1/auth/me \
  -H "Cookie: auth_token=eyJhbGciOi..."
```

Response:
```json
{
  "ID": 1,
  "Username": "admin",
  "Email": "admin@example.com",
  "DisplayName": "admin",
  "Tickets": 99999,
  "IsAdmin": true,
  "CreatedAt": "2026-07-04T..."
}
```

---

## Novels

### `GET /novels`

List novels with filtering, sorting, and pagination.

```bash
curl "http://localhost:8080/api/v1/novels?page=1&limit=20&sort=rating&order=desc&status=ongoing&genres=action,fantasy&genre_mode=or&min_rating=4.0&min_chapters=100"
```

**Query Parameters:**

| Param | Type | Default | Description |
|---|---|---|---|
| `page` | int | `1` | Page number |
| `limit` | int | `20` | Items per page (max 100) |
| `q` | string | — | Text search (title, alt_title, description) |
| `status` | string | — | `ongoing`, `completed`, `hiatus`, `dropped` |
| `genre` | string | — | Single genre slug (legacy) |
| `genres` | string | — | Comma-separated genre slugs |
| `genre_mode` | string | `or` | `and` or `or` |
| `min_chapters` | int | `0` | Minimum chapter count |
| `min_rating` | float | `0` | Minimum rating |
| `min_reviews` | int | `0` | Minimum rating_count |
| `sort` | string | `created_at` | `created_at`, `title`, `views`, `chapters`, `rating`, `readers`, `reviews` |
| `order` | string | `desc` | `desc` or `asc` |

Response:
```json
{
  "data": [
    {
      "ID": 1,
      "Title": "Having Dinner with His Brother...",
      "AltTitle": "...",
      "Slug": "having-dinner-with-his-brother-...",
      "Author": "半条活鱼",
      "AuthorSlug": "ban-tiao-huo-yu",
      "Status": "completed",
      "Views": 345678,
      "Rating": 3.5,
      "RatingCount": 234,
      "Chapters": 135,
      "Readers": 1234,
      "Chars": "250K",
      "AIPercent": "37%",
      "Description": "...",
      "CoverURL": "",
      "CreatedAt": "...",
      "Genres": [
        { "ID": 22, "Slug": "romance", "Name": "Romance" },
        { "ID": 30, "Slug": "slice-of-life", "Name": "Slice of Life" },
        { "ID": 35, "Slug": "urban-life", "Name": "Urban Life" }
      ]
    }
  ],
  "page": 1,
  "limit": 20,
  "total": 12,
  "total_pages": 1
}
```

### `GET /novels/:id`

Get a single novel by ID with genres.

```bash
curl http://localhost:8080/api/v1/novels/1
```

Returns a single novel object (same structure as array items above).

### `GET /novels/:id/chapters`

Get paginated chapters for a novel.

```bash
curl "http://localhost:8080/api/v1/novels/1/chapters?page=1&limit=50"
```

Response:
```json
{
  "data": [
    { "ID": 1, "NovelID": 1, "Number": 1, "Title": "Chapter 1", "IsLocked": false, "TicketCost": 0 }
  ],
  "page": 1,
  "limit": 50,
  "total": 135
}
```

### `POST /novels`

Create a novel manually. **Auth required.**

```bash
curl -X POST http://localhost:8080/api/v1/novels \
  -H "Content-Type: application/json" \
  -H "Cookie: auth_token=..." \
  -d '{
    "title": "My New Novel",
    "alt_title": "我的新小说",
    "author": "Author Name",
    "status": "ongoing",
    "description": "A thrilling story...",
    "cover_url": "https://example.com/cover.jpg",
    "genre_ids": [1, 12, 22]
  }'
```

### `PUT /novels/:id`

Update a novel. **Auth required.** All fields optional; only provided fields are updated.

```bash
curl -X PUT http://localhost:8080/api/v1/novels/1 \
  -H "Content-Type: application/json" \
  -H "Cookie: auth_token=..." \
  -d '{
    "title": "Updated Title",
    "status": "completed",
    "genre_ids": [1, 5, 22]
  }'
```

### `DELETE /novels/:id`

Delete a novel and its genre associations. **Auth required.**

```bash
curl -X DELETE http://localhost:8080/api/v1/novels/1 \
  -H "Cookie: auth_token=..."
```

Response:
```json
{ "message": "novel deleted" }
```

### `GET /novels/trending`

Top 20 novels by views.

### `GET /novels/recommendations`

Top 12 novels by rating (desc) then views (desc).

### `GET /novels/random`

Random novels. Query param: `?limit=10` (default 10, max 50).

---

## Import (External API)

### `GET /novels/import/search`

Search for novels on NovelUpdates via the free Consumet API.

```bash
curl "http://localhost:8080/api/v1/novels/import/search?q=solo+leveling"
```

Response:
```json
{
  "data": [
    {
      "id": "solo-leveling",
      "title": "Solo Leveling",
      "url": "https://www.novelupdates.com/series/solo-leveling/",
      "image": "https://cdn.novelupdates.com/images/..."
    }
  ]
}
```

### `POST /novels/import`

Import a novel from NovelUpdates by source ID. **Auth required.**

```bash
curl -X POST http://localhost:8080/api/v1/novels/import \
  -H "Content-Type: application/json" \
  -H "Cookie: auth_token=..." \
  -d '{
    "source_id": "solo-leveling",
    "with_chapters": true
  }'
```

| Field | Type | Required | Description |
|---|---|---|---|
| `source_id` | string | Yes | NovelUpdates series ID (slug) |
| `with_chapters` | bool | No | Import chapter list (default: `false`) |

The import process:
1. Fetches metadata from Consumet API (title, author, description, cover, genres, status, chapters)
2. Auto-generates a URL slug from the title
3. Matches or creates genre records
4. Creates the novel record with genre associations
5. Optionally creates chapter records
6. Returns the created novel with genres

---

## Requests

### `POST /requests`

Submit a novel request/suggestion. **Auth required.**

```bash
curl -X POST http://localhost:8080/api/v1/requests \
  -H "Content-Type: application/json" \
  -H "Cookie: auth_token=..." \
  -d '{
    "novel_title": "Solo Leveling",
    "novel_url": "https://novelupdates.com/series/solo-leveling/",
    "source": "novelupdates"
  }'
```

### `PUT /requests/:id`

Review (approve/reject) a request. **Auth required.**

```bash
curl -X PUT http://localhost:8080/api/v1/requests/1 \
  -H "Content-Type: application/json" \
  -H "Cookie: auth_token=..." \
  -d '{"status": "approved"}'
```

Valid statuses: `approved`, `rejected`, `completed`

---

## Chapters

### `GET /chapters/:id`

Get a single chapter by ID.

---

## Search

### `GET /search`

Search novels by title or author (case-insensitive ILIKE).

```bash
curl "http://localhost:8080/api/v1/search?q=solo&page=1&limit=20"
```

---

## Genres

### `GET /genres`

List all genres.

```bash
curl http://localhost:8080/api/v1/genres
```

Response:
```json
{
  "data": [
    { "ID": 1, "Slug": "action", "Name": "Action" },
    { "ID": 2, "Slug": "adult", "Name": "Adult" }
  ]
}
```

---

## Ranking

### `GET /ranking/:period`

Get ranking by period. Period: `daily`, `weekly`, `monthly`, `all_time`.

---

## Updates

### `GET /updates`

Recent chapter updates. Query param: `?limit=20`.

---

## Library

### `GET /library`

Get authenticated user's followed novels and reading history. **Auth required.**

---

## Votes

### `POST /votes`

Vote for a novel. **Auth required.** One vote per user per novel.

```bash
curl -X POST http://localhost:8080/api/v1/votes \
  -H "Content-Type: application/json" \
  -H "Cookie: auth_token=..." \
  -d '{"novel_id": 1}'
```

---

## Leaderboard

### `GET /leaderboard`

Top users by tickets. Query param: `?sort=tickets`.

---

## News

### `GET /news`

List news articles. Query params: `?type=news&page=1&limit=10`. Type: `news` or `changelog`.

### `GET /news/:id`

Get a single news article.

---

## Stats

### `GET /stats`

Platform-wide statistics.

Response:
```json
{
  "total_novels": 12,
  "total_chapters": 7456,
  "total_users": 5,
  "total_views": 1234567,
  "total_votes": 89,
  "total_requests": 3
}
```

---

## Author & Profile

### `GET /author/:name/novels`

Get novels by author name.

### `GET /profile/:id`

Get a user's public profile.

---

## Error Responses

All errors follow this format:

```json
{ "error": "description of the error" }
```

Common HTTP status codes:

| Code | Description |
|---|---|
| `200` | Success |
| `201` | Created |
| `400` | Bad request (missing/invalid params) |
| `401` | Unauthorized (missing/invalid token) |
| `404` | Not found |
| `500` | Internal server error |
