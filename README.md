# Tracker

Collaborative challenges and habit tracking service.

## Docker (recommended)

```bash
cp .env.example .env
# Edit .env with a real JWT_SECRET (DATABASE_URL is overridden by compose)

docker compose up --build
```

- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- PostgreSQL: localhost:5432 (user/pass/db: `tracker`)

Migrations run automatically on first start via Postgres init scripts.

## Manual Setup

### Prerequisites

- Go 1.24+
- Node.js 18+
- PostgreSQL 14+

### Backend

```bash
# 1. Install dependencies
go mod tidy

# 2. Configure environment
cp .env.example .env
# Edit .env with your DATABASE_URL and JWT_SECRET

# 3. Create database and apply migrations
createdb tracker
for f in internal/db/migrations/*.sql; do psql $DATABASE_URL -f "$f"; done

# 4. Run the server
go run ./cmd/server
```

Backend runs on `http://localhost:8080` by default.

## Frontend Setup

```bash
cd frontend
npm install
npm run dev
```

Frontend runs on `http://localhost:5173` with API requests proxied to the backend.

## API

All endpoints prefixed with `/api/v1`. Errors return `{"error": "message"}`.

### Auth

| Method | Path | Body | Description |
|--------|------|------|-------------|
| POST | `/auth/register` | `{email, password, name}` | Register, returns token pair |
| POST | `/auth/login` | `{email, password}` | Login, returns token pair |
| POST | `/auth/refresh` | `{refresh_token}` | Rotate tokens |
| POST | `/auth/logout` | `{refresh_token}` | Invalidate refresh token |

### Users (protected)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/users/me` | Get current user |
| PATCH | `/users/me` | Update name, bio, timezone |
| GET | `/users/me/stats` | Personal stats |

### Challenges (protected)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/categories` | List categories |
| POST | `/challenges` | Create challenge |
| GET | `/challenges?public=true&category=&search=&page=` | Browse public |
| GET | `/challenges/my` | My challenges |
| GET | `/challenges/{id}` | Get challenge |
| PATCH | `/challenges/{id}` | Update (upcoming only) |
| POST | `/challenges/{id}/finish` | End challenge |
| GET | `/challenges/{id}/invite-link` | Get invite token |
| POST | `/challenges/{id}/join` | Join public |
| POST | `/challenges/join/{token}` | Join by invite |
| DELETE | `/challenges/{id}/participants/{userID}` | Remove participant |
| GET | `/challenges/{id}/leaderboard` | Leaderboard |
| GET | `/challenges/{id}/stats` | Aggregate stats |

### Check-ins (protected)

| Method | Path | Description |
|--------|------|-------------|
| POST | `/challenges/{id}/checkins` | Upsert check-in |
| GET | `/challenges/{id}/checkins/my` | My check-ins |
| GET | `/challenges/{id}/feed?page=` | Challenge feed |
| POST | `/checkins/{id}/comments` | Add comment |
| GET | `/checkins/{id}/comments` | List comments |
| POST | `/checkins/{id}/like` | Toggle like |
| POST | `/uploads` | Upload image (multipart, max 5MB) |

### Token pair response

```json
{
  "access_token": "...",
  "refresh_token": "..."
}
```

Access tokens expire after **15 minutes**. Refresh tokens expire after **7 days**.