# Ticket System API

A REST backend for a ticket system built with Go. Users can register, log in, create tickets, view their own tickets, and update ticket status with JWT authentication.

## Features

- User registration and login with bcrypt password hashing
- JWT-based authentication (`Authorization: Bearer <token>`)
- Ticket CRUD scoped to the authenticated user
- Status workflow: `open` → `in_progress` → `closed` (closed tickets cannot be reopened)
- In-memory storage (data resets on restart)
- Dockerized for local and cloud deployment

## API Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/health` | No | Health check |
| POST | `/auth/register` | No | Register a new user |
| POST | `/auth/login` | No | Login and receive JWT |
| POST | `/tickets` | Yes | Create a ticket |
| GET | `/tickets` | Yes | List your tickets |
| GET | `/tickets/{id}` | Yes | Get your ticket by ID |
| PATCH | `/tickets/{id}/status` | Yes | Update ticket status |

## Local Run (Go)

```bash
cp .env.example .env
go mod download
go run .
```

Health check:

```bash
curl http://localhost:8080/health
```

Expected response:

```json
{"status":"ok"}
```

## Docker Run

```bash
docker build -t ticket-system .
docker run -p 8080:8080 -e JWT_SECRET=your-secret ticket-system
```

Or with Docker Compose:

```bash
docker compose up --build
```

## Example Usage

```bash
# Register
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"secret123"}'

# Login
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"secret123"}'

# Create ticket
curl -X POST http://localhost:8080/tickets \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"title":"Fix login bug","description":"Users cannot reset password"}'

# List tickets
curl http://localhost:8080/tickets \
  -H "Authorization: Bearer <token>"

# Update status
curl -X PATCH http://localhost:8080/tickets/<id>/status \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"status":"in_progress"}'
```

After deployment, verify:

\`\`\`
curl https://ticket-system-5z9p.onrender.com/health
\`\`\`
```

## Deployed URL

* App URL: `https://ticket-system-5z9p.onrender.com`
* Health check: `https://ticket-system-5z9p.onrender.com/health`

## Assumptions

- Users are identified by **email** and **password**.
- Login returns `{"token":"<jwt>"}`.
- Tickets require a **title**; **description** is optional.
- New tickets start with status `open`.
- Valid status transitions: `open` → `in_progress`, `in_progress` → `closed`.
- Users can only access tickets they created (other users' tickets return 404).
- Storage is in-memory; data is lost when the process restarts.
- `JWT_SECRET` should be set in production; a dev default is used locally if unset.

## Project Structure

```
.
├── main.go
├── internal/
│   ├── auth/        # JWT and password hashing
│   ├── handlers/    # HTTP handlers
│   ├── middleware/  # JWT auth middleware
│   ├── models/      # Request/response types
│   └── store/       # In-memory data store
├── Dockerfile
├── docker-compose.yml
└── .env.example
```
