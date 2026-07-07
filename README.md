# Ticket System (Golang Backend Intern Assignment)

A small REST API where a user can register, log in, create tickets, view only
their own tickets, and move a ticket through `open -> in_progress -> closed`.

## Tech / Design Choices

- **Language:** Go 1.22, **standard library only** — no third-party modules.
  This was a deliberate choice so the project builds and runs with zero
  network access (nothing to `go mod download`), which also makes the
  Docker build faster and more reliable on free hosting tiers.
- **Storage:** in-memory (thread-safe via `sync.RWMutex`), as explicitly
  allowed by the assignment scope. All data access goes through a `Store`
  type, so swapping in Postgres/SQLite later would only mean rewriting
  `store.go`, not the handlers.
- **Auth:** JWT (HS256), implemented by hand with `crypto/hmac` +
  `crypto/sha256` — again to avoid a third-party JWT library.
- **Passwords:** hashed with a hand-rolled PBKDF2-HMAC-SHA256 (100,000
  iterations, random 16-byte salt per user), since `bcrypt` isn't in the
  standard library and this avoids pulling in `golang.org/x/crypto`.
  Never stored in plaintext.
- **Routing:** Go 1.22's built-in `net/http.ServeMux`, which now supports
  method matching and path parameters (`GET /tickets/{id}`) natively — no
  router library needed.

## Project Structure

```
main.go             entrypoint, routes, server startup
models.go           User / Ticket structs, status transition rules
store.go            in-memory data store
auth_handlers.go     /auth/register, /auth/login
ticket_handlers.go   /tickets endpoints
middleware.go        JWT auth middleware, JSON helpers
jwt.go               hand-rolled HS256 JWT encode/decode
password.go          PBKDF2 password hashing
util.go              ID generation, small helpers
Dockerfile
.env.example
```

## Endpoints

| Method | Endpoint               | Auth | Purpose                    |
|--------|-------------------------|------|----------------------------|
| GET    | /health                | No   | Health check               |
| POST   | /auth/register         | No   | Register (email, password) |
| POST   | /auth/login            | No   | Login, returns JWT         |
| POST   | /tickets               | Yes  | Create ticket               |
| GET    | /tickets               | Yes  | List own tickets            |
| GET    | /tickets/{id}          | Yes  | Get own ticket by ID        |
| PATCH  | /tickets/{id}/status   | Yes  | Update own ticket's status  |

Protected routes require `Authorization: Bearer <token>`.

## Local Run (Go directly)

```bash
go run .
curl http://localhost:8080/health
```

## Local Run (Docker)

```bash
docker build -t ticket-system .
docker run -p 8080:8080 ticket-system
curl http://localhost:8080/health
```

Expected health response:

```json
{"status": "ok"}
```

To set a real JWT secret:

```bash
docker run -p 8080:8080 -e JWT_SECRET=some-long-random-string ticket-system
```

## Example Requests

```bash
# Register
curl -X POST http://localhost:8080/auth/register \
  -d '{"email":"a@test.com","password":"password123"}'

# Login
curl -X POST http://localhost:8080/auth/login \
  -d '{"email":"a@test.com","password":"password123"}'
# -> {"token": "..."}

# Create ticket
curl -X POST http://localhost:8080/tickets \
  -H "Authorization: Bearer <token>" \
  -d '{"title":"Printer broken","description":"No toner"}'

# List own tickets
curl http://localhost:8080/tickets -H "Authorization: Bearer <token>"

# Update status
curl -X PATCH http://localhost:8080/tickets/<id>/status \
  -H "Authorization: Bearer <token>" \
  -d '{"status":"in_progress"}'
```

## Assumptions

- **Ownership check response codes:** if a ticket ID doesn't exist at all,
  the API returns `404`. If it exists but belongs to another user, it
  returns `403` (rather than a `404` that hides existence). This wasn't
  explicitly specified, so it's called out here.
- **Password length:** a minimum of 6 characters is enforced on register;
  not specified in the contract but a reasonable minimal safeguard.
- **Status transitions:** only the forward path
  `open -> in_progress -> closed` is allowed. Any other transition
  (including re-opening a closed ticket, or skipping straight from
  `open` to `closed`) returns `400`.
- **Email uniqueness:** registering an already-used email returns `409`.
- **Token lifetime:** JWTs are valid for 24 hours.

## Deployment

Deployed on [Render](https://render.com) free tier using the included
Dockerfile (Render auto-detects and builds it — no extra config needed
beyond setting the `JWT_SECRET` environment variable in the dashboard).

- GitHub repo: `<add your repo URL here>`
- Deployed URL: `<add your Render URL here>`
- Public health check: `<deployed URL>/health`

### Deployment steps (Render)

1. Push this repo to GitHub.
2. On Render: New -> Web Service -> connect the repo.
3. Environment: Docker (it will pick up the `Dockerfile` automatically).
4. Add environment variable `JWT_SECRET` with a random value.
5. Deploy. Render provides the `PORT` env var automatically; the app
   already reads it.
6. Verify with `curl https://<your-service>.onrender.com/health`.
