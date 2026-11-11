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
