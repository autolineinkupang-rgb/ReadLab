# API Documentation

Base URL: `http://localhost:8080/api/v1`

**Authentication:** JWT tokens are sent via HTTP-only cookie (`auth_token`).

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

Create a new user account. Password must be 8+ chars with uppercase, lowercase, digit, and special character.

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username": "newuser", "email": "user@example.com", "password": "Secure@123"}'
```

Response `201`:
```json
{
  "user": { "id": 6, "username": "newuser", "email": "user@example.com", "role": "member" }
}
```

### `POST /auth/login`

Authenticate and receive a JWT token (set as cookie).

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@example.com", "password": "admin123"}'
```

Response:
```json
{
  "user": { "id": 1, "username": "admin", "email": "admin@example.com", "role": "admin" }
}
```

### `POST /auth/logout`

Clear the auth cookie.

### `GET /auth/me`

Get the currently authenticated user's profile. **Auth required.**

Response:
```json
{
  "id": 1,
  "username": "admin",
  "email": "admin@example.com",
  "display_name": "admin",
  "avatar_url": "",
  "tickets": 99999,
  "xp": 1500,
  "role": "admin",
  "daily_reward": {
    "can_claim": true,
    "reward": 2
  }
}
```

The `daily_reward` field includes the current config value from the server cache.

### `PUT /auth/password`

Change password. **Auth required.**

```bash
curl -X PUT http://localhost:8080/api/v1/auth/password \
  -H "Content-Type: application/json" \
  -H "Cookie: auth_token=..." \
  -d '{"current_password": "oldPass1!", "new_password": "NewPass@123"}'
```

Validates current password, then applies the same strength rules (8+ chars, upper, lower, digit, special).

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
| `genres` | string | — | Comma-separated genre slugs |
| `genre_mode` | string | `or` | `and` or `or` |
| `tag` | string | — | Single tag slug |
| `tags` | string | — | Comma-separated tag slugs |
| `tag_mode` | string | `or` | `and` or `or` |
| `exclude_tags` | string | — | Tags to exclude |
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
        { "ID": 22, "Slug": "romance", "Name": "Romance" }
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

### `GET /novels/:id/chapters`

Get paginated chapters for a novel.

```bash
curl "http://localhost:8080/api/v1/novels/1/chapters?page=1&limit=50"
```

### `GET /novels/:id/chapters/:num`

Get a chapter by number. Supports optional auth for locked chapters.

### `POST /novels`

Create a novel manually. **Writer or admin.**

### `PUT /novels/:id`

Update a novel. **Writer or admin.** All fields optional.

### `DELETE /novels/:id`

Delete a novel and its genre associations. **Writer or admin.**

### `GET /novels/trending`

Top 20 novels by views.

### `GET /novels/recommendations`

Top 12 novels by rating (desc) then views (desc).

### `GET /novels/random`

Random novels. Query param: `?limit=10` (default 10, max 50).

---

## Reviews

### `GET /novels/:id/reviews`

List reviews for a novel. Includes rating summary and nested replies.

Response:
```json
{
  "data": [
    {
      "id": 1,
      "user_id": 1,
      "novel_id": 1,
      "rating": 4,
      "content": "Great novel!",
      "edit_count": 0,
      "created_at": "...",
      "updated_at": "...",
      "user": { "id": 1, "username": "admin", "display_name": "admin" },
      "replies": [
        { "id": 2, "user_id": 2, "content": "I agree!", "user": { ... } }
      ]
    }
  ],
  "rating_summary": {
    "average": 4.2,
    "count": 15,
    "distribution": { "1": 0, "2": 1, "3": 2, "4": 5, "5": 7 }
  }
}
```

### `POST /novels/:id/reviews`

Create a review. **Auth required.**

```bash
curl -X POST http://localhost:8080/api/v1/novels/1/reviews \
  -H "Content-Type: application/json" \
  -H "Cookie: auth_token=..." \
  -d '{"rating": 4, "content": "Great story!"}'
```

- Must have read 5+ chapters (or pay gate bypass)
- One review per user per novel (or pay replace cost)
- Rating: 0-5 (0 for replies)

**Upgrade errors** return 403 with upgrade metadata:
```json
{
  "error": "you need to read at least 5 chapters before reviewing",
  "chapter_count": 2,
  "upgrade_available": true,
  "upgrade_cost": 50,
  "upgrade_type": "gate"
}
```

Pass `{"upgrade": true}` in the request to confirm and pay.

### `PUT /novels/:id/reviews/:reviewId`

Update a review. **Auth required.**

- Max 5 edits per review
- After 5 edits, pass `{"upgrade": true}` to pay `edit_reset_cost` and reset count
- To replace a different user's review, pass `{"upgrade": true}` to pay `replace_review_cost`

---

## Rewards

### `POST /rewards/daily`

Claim daily login reward. **Auth required.** Once per day (Asia/Makassar timezone).

Response:
```json
{
  "message": "daily reward claimed",
  "tickets": 102,
  "rewarded": 2
}
```

If already claimed:
```json
{
  "error": "daily reward already claimed",
  "next_claim_in": "12h30m0s",
  "next_claim_at": "2026-07-10T00:00:00+08:00"
}
```

### `GET /rewards/status`

Check daily reward claim status. **Auth required.**

Response:
```json
{
  "daily_reward": {
    "can_claim": true,
    "reward": 2,
    "next_claim_at": ""
  }
}
```

### `POST /admin/rewards/monthly`

Distribute monthly XP leaderboard rewards. **Admin only.**

Resets all user XP to 0. Awards tickets to top N users (default 10). Query params: `?period=2026-07&limit=10`.

---

## Ticket Config

### `GET /config/upgrade-costs`

Public endpoint returning current upgrade costs from server cache.

```bash
curl http://localhost:8080/api/v1/config/upgrade-costs
```

Response:
```json
{
  "edit_reset": 20,
  "gate_bypass": 50,
  "replace_review": 100
}
```

### `GET /admin/config/tickets`

List all ticket configuration keys and values. **Admin only.**

### `PUT /admin/config/tickets`

Update a ticket config value. **Admin only.**

```bash
curl -X PUT http://localhost:8080/api/v1/admin/config/tickets \
  -H "Content-Type: application/json" \
  -H "Cookie: auth_token=..." \
  -d '{"key": "daily_reward", "value": 5}'
```

Reloads the server cache on success. Changes take effect immediately.

Config keys: `daily_reward`, `novel_contribution`, `monthly_leaderboard`, `edit_reset_cost`, `gate_bypass_cost`, `replace_review_cost`

---

## Import & Scraping

### `GET /novels/import/search`

Search for novels on NovelUpdates via the free Consumet API.

### `POST /novels/import`

Import a novel from NovelUpdates by source ID. **Admin only.**

### `POST /novels/scrape`

Scrape novel metadata from a URL (NovelUpdates, RoyalRoad, NovelBin). **Admin only.**

### `POST /novels/scrape/import`

Scrape and import chapters from a URL. **Admin only.**

### `POST /novels/lncrawl`

Crawl chapters via LN Crawl engine. **Admin only.**

---

## Chapters

### `GET /chapters/:id`

Get a single chapter by ID.

### `POST /admin/novels/:id/chapters`

Create a chapter. **Writer or admin.**

### `PUT /admin/novels/:id/chapters/:chapterID`

Update a chapter. **Writer or admin.**

### `DELETE /admin/chapters/:id`

Delete a chapter. **Admin only.**

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

## Reading

### `POST /novels/:id/chapters/:num/read`

Track a chapter read. **Auth required.**

### `POST /novels/:id/chapters/:num/xp`

Claim XP for reading a chapter. **Auth required.** One claim per chapter.

### `GET /novels/:id/my-progress`

Get user's reading progress for a novel. **Auth required.** Includes `chapter_count`, `can_review`, `my_review`.

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

## Requests

### `POST /requests`

Submit a novel request. **Auth required.**

### `GET /requests`

List user's requests. **Auth required.**

### `PUT /requests/:id`

Review (approve/reject) a request. **Admin only.** Status: `approved`, `rejected`, `completed`.

---

## Leaderboard

### `GET /leaderboard`

Top users by XP or tickets. Query param: `?sort=xp` (default) or `?sort=tickets`.

---

## News

### `GET /news`

List news articles. Query params: `?type=news&page=1&limit=10`. Type: `news` or `changelog`.

### `GET /news/:id`

Get a single news article.

### `POST /admin/news`

Create news article. **Admin only.**

### `PUT /admin/news/:id`

Update news. **Admin only.**

### `DELETE /admin/news/:id`

Delete news. **Admin only.**

---

## Stats

### `GET /stats`

Platform-wide statistics.

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

## AI Translation

### `GET /user/ai-settings`

Get authenticated user's AI translation settings. **Auth required.**

### `PUT /user/ai-settings`

Update AI translation settings. **Auth required.**

```bash
curl -X PUT http://localhost:8080/api/v1/user/ai-settings \
  -H "Content-Type: application/json" \
  -H "Cookie: auth_token=..." \
  -d '{
    "provider": "openrouter",
    "model": "google/gemini-2.0-flash-exp:free",
    "endpoint": "https://openrouter.ai/api/v1/chat/completions",
    "key": "sk-or-v1-...",
    "target_language": "id-ID",
    "instruction": "Use casual Indonesian"
  }'
```

### `POST /translate/ai`

Translate a chapter using the user's configured AI provider. **Auth required.**

```bash
curl -X POST http://localhost:8080/api/v1/translate/ai \
  -H "Content-Type: application/json" \
  -H "Cookie: auth_token=..." \
  -d '{"chapter_id": 1}'
```

---

## Admin

### `GET /admin/users`

List all users. **Admin only.**

### `GET /admin/users/:id`

Get user details. **Admin only.**

### `PUT /admin/users/:id`

Update user (role, tickets, xp, etc.). **Admin only.**

### `DELETE /admin/users/:id`

Delete a user. **Admin only.**

### `POST /admin/users/admin`

Create a new admin user. **Admin only.**

### `GET /admin/stats`

Platform statistics. **Admin only.**

### `GET /admin/reviews`

List all reviews. **Admin only.**

### `DELETE /admin/reviews/:id`

Delete a review. **Admin only.**

### `GET /admin/requests`

List all requests. **Admin only.**

---

## Author & Profile

### `GET /author/:name/novels`

Get novels by author name.

### `GET /profile/:id`

Get a user's public profile.

---

## Share

### `POST /novels/:id/share`

Share a novel (earns XP). **Auth required.**

---

## Translate (Legacy)

### `POST /translate`

Legacy translate endpoint using Google Translate (no auth needed).

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
| `403` | Forbidden (insufficient role) |
| `404` | Not found |
| `409` | Conflict (duplicate, already claimed) |
| `500` | Internal server error |
