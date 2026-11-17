# My Go App

This project is a simple Go application that demonstrates the structure and organization of a Go project. It includes an HTTP server, handlers for processing requests, and utility functions.

## Project Structure

```
my-go-app
├── cmd
│   └── main.go          # Entry point of the application
├── internal
│   ├── handlers
│   │   └── handler.go   # HTTP request handlers
│   └── models
│       └── model.go     # Data models
├── pkg
│   └── utils
│       └── utils.go     # Utility functions
├── go.mod               # Module dependencies
└── go.sum               # Module checksums
```

## Setup Instructions

1. **Clone the repository:**
   ```
   git clone <repository-url>
   cd my-go-app
   ```

2. **Install dependencies:**
   ```
   go mod tidy
   ```

3. **Run the application:**
   ```
   go run cmd/main.go
   ```

## Usage

Once the application is running, you can access it at `http://localhost:8080`. You can test the endpoints defined in the handlers.

## Contributing

Feel free to submit issues or pull requests for improvements or bug fixes.
## Database schema and migrations

This service owns the Postgres schema (Prisma was removed from webbuilderv2).
GORM is used for queries only; schema changes are versioned SQL files in
`migrations/`, applied with [goose](https://github.com/pressly/goose).

| Command | What it does |
|---|---|
| `make migrate-create name=add_x` | New `migrations/<timestamp>_add_x.sql` |
| `make migrate-up` | Apply pending migrations |
| `make migrate-status` | List applied / pending |
| `make migrate-down` | Roll back the latest migration |
| `make migrate-baseline` | **Once** per database that Prisma already migrated |
| `make seed` | Demo marketplace data (deletes existing marketplace data; dev only) |

`00001_baseline.sql` is the full schema at the Prisma handover. A fresh
database runs it via `migrate-up`. An existing database must instead run
`migrate-baseline`, which checks `_prisma_migrations` contains
`20261111120000_add_element_order_key` and records the baseline as applied.

## Realtime collaboration (WebSocket)

The collaborative editing server lives in this process under `internal/realtime`
and is served at `ws(s)://<host>/ws/:project?token=<clerk-jwt>`. There is no
separate realtime service: point the frontend's `VITE_COLLAB_WS_URL` at this
server's URL.

- It needs Redis: set `REDIS_URL` (or `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`, `REDIS_DB`).
  It reuses `DATABASE_URL` and `CLERK_SECRET_KEY`, and listens on `PORT`.
- If Redis or its database connection fails at startup, the REST API still starts,
  `/ws` is not registered, and an `ERROR: realtime WebSocket disabled` line is logged.
- Access: connecting requires being the project owner or a collaborator. Viewers can read and
  see presence but every mutating message is refused (`ACCESS_DENIED`).
- `DEV_MODE=true` skips JWT checks and access checks on `/ws` (never set it in production);
  every client is `DEV_USER_ID`; add `REALTIME_ENFORCE_ACCESS=true` to keep the access checks.
- Logic messages: `logic:update` (`{scope: "project"|"page", patch}`; the server broadcasts the
  full resulting document as `logic:update` with `{scope, pageId?, logic, revision}`) and
  `logic:sync` (`{project, page?, pageId, projectRevision, pageRevision}`), also sent on join.
  A revision increases with every commit of that document; clients ignore older ones.
- Protocol notes: see `docs/realtime/websocket_api.md` and `docs/realtime/architecture.md`.
