# Perpustakaan

A small library management project written in Go. It can be used from the
terminal or from a web browser, and all data is stored in PostgreSQL.

The project started as a CLI application. The CLI is still available, while the
web page and REST API use the same business logic and the same database.

## Features

- **Books**: add, search, list, delete, and add stock
- **Members**: add, search, list, and delete
- **Loans**: borrow and return books, with automatic due dates and fines
- **Web interface**: plain HTML, CSS, and JavaScript, no frontend framework
- **REST API**: JSON endpoints for all main operations
- **PostgreSQL storage**: tables are created automatically on startup
- **JSON import**: move data from an old JSON file into PostgreSQL

## Rules

1. A member can borrow at most **3 books** at the same time.
2. A member with an overdue loan cannot borrow another book.
3. A book can only be borrowed when its available stock is more than zero.
4. The due date is **3 days** after the borrowing date.
5. A late return costs **Rp1.000 per day**.
6. A book cannot be returned twice.
7. A book or member that still has an active loan cannot be deleted.

## Tech Stack

| Part | Used |
| --- | --- |
| Language | Go 1.26+ |
| Database | PostgreSQL |
| Database driver | `github.com/lib/pq` |
| Web server | Go standard library (`net/http`) |
| Frontend | HTML, CSS, vanilla JavaScript |

## Project Structure

```
perpustakaan_cli/
├── main.go          # Entry point, command-line menus, and flags
├── library.go       # Business logic
├── database.go      # PostgreSQL connection, tables, load and save
├── env.go           # Small reader for the .env file
├── api.go           # REST API and static web server
├── book.go          # Book model
├── anggota.go       # Member model
├── borrow.go        # Loan model
├── file.go          # JSON load/save helpers (used for import and tests)
├── web/
│   ├── index.html
│   ├── style.css
│   └── app.js
├── data/contoh.json # Safe sample data for testing the import
├── .env.example     # Example database configuration
├── docker-compose.yml
└── go.mod
```

## Requirements

- Go 1.26 or newer
- PostgreSQL 14 or newer, or Docker

## Setup

### 1. Start PostgreSQL

The easiest way is Docker:

```bash
docker compose up -d
```

Before starting Docker, create your local configuration:

```bash
cp .env.example .env
chmod 600 .env
```

Edit `.env` and set `POSTGRES_PASSWORD` to your own password. The Compose file
uses that variable, so the password is not stored in `docker-compose.yml`.

The container is bound to `127.0.0.1`, which means the database can only be
reached from this computer, not from the local network.

The included Compose file creates:

- database: `db_library`
- user: `library`
- password: from `POSTGRES_PASSWORD` in `.env`
- host port: `5433` (inside the container it is still `5432`)

Port `5433` is used so the container does not conflict with a local
PostgreSQL server on `5432`.

If you already have PostgreSQL installed, create the database and user
manually:

```sql
CREATE USER library WITH PASSWORD 'your_password_here';
CREATE DATABASE db_library OWNER library;
```

The application reads the connection string from `DATABASE_URL` in `.env`:

```text
DATABASE_URL=postgres://library:your_password_here@localhost:5433/db_library?sslmode=disable
```

The app reads `.env` automatically at startup. A value that is already set in
the terminal still takes priority. You can also run it directly:

```bash
DATABASE_URL="postgres://library:your_password_here@localhost:5433/db_library?sslmode=disable" go run . -web
```

When it connects, the app prints the target database without the password:

```text
TERHUBUNG KE DATABASE: library@localhost:5433/db_library
```

### 2. Run the web app

```bash
go mod download
go run . -web
```

Open <http://localhost:8080>.

To use another port:

```bash
go run . -web -addr :3000
```

### 3. Run the CLI

```bash
go run .
```

The CLI and the web app share the same data, so changes from one are visible in
the other.

### 4. Import old JSON data (optional)

If you want to move old JSON data into PostgreSQL, run this once on an empty
database:

```bash
go run . -import data/contoh.json
```

You can also point it to your own JSON file. Real data files should not be
committed to Git.

The import is refused when the database already contains books, members, or
loans.

## REST API

All endpoints return JSON. Errors use the form `{"error": "message"}`.

| Method | Endpoint | Description |
| --- | --- | --- |
| `GET` | `/api/books` | List books. Add `?search=keyword` to search by title. |
| `GET` | `/api/books/{id}` | Get one book. |
| `POST` | `/api/books` | Add a book. |
| `DELETE` | `/api/books/{id}` | Delete a book. |
| `POST` | `/api/books/{id}/stock` | Add stock. Body: `{"tambahan": 2}`. |
| `GET` | `/api/members` | List members. |
| `GET` | `/api/members/{id}` | Get one member. |
| `POST` | `/api/members` | Add a member. Body: `{"nama": "Rina"}`. |
| `DELETE` | `/api/members/{id}` | Delete a member. |
| `GET` | `/api/loans` | List loans. Add `?active=1` for active loans only. |
| `POST` | `/api/loans` | Borrow a book. Body: `{"buku_id": 1, "anggota_id": 1}`. |
| `POST` | `/api/loans/{id}/return` | Return a book. |

### Examples

```bash
# List books
curl http://localhost:8080/api/books

# Add a book
curl -X POST http://localhost:8080/api/books \
  -H "Content-Type: application/json" \
  -d '{"judul": "Bumi Manusia", "tahun_terbit": 1980, "jumlah_eksemplar": 3}'

# Add a member
curl -X POST http://localhost:8080/api/members \
  -H "Content-Type: application/json" \
  -d '{"nama": "Rina"}'

# Borrow a book
curl -X POST http://localhost:8080/api/loans \
  -H "Content-Type: application/json" \
  -d '{"buku_id": 1, "anggota_id": 1}'

# Return a book
curl -X POST http://localhost:8080/api/loans/1/return
```

A book response looks like this:

```json
{
  "id": 1,
  "judul": "Bumi Manusia",
  "tahun_terbit": 1980,
  "jumlah_eksemplar": 3,
  "jumlah_tersedia": 3
}
```

## Database Tables

The app creates these tables automatically:

- `buku` — books and stock
- `anggota` — library members
- `peminjaman` — loan history
- `counter` — next ID for books, members, and loans

For each write, the app loads the current data, applies the rules, and saves the
result in one transaction. This keeps the code simple. It is fine for a small
library, but not designed for large numbers of users writing at the same time.

## Tests

```bash
go test ./...
```

The tests cover the in-memory business logic, so PostgreSQL does not need to be
running.

## Security Notes

- Keep `.env` private. It is ignored by Git and should only be readable by
  your user (`chmod 600 .env`).
- `docker-compose.yml` reads the database password from `.env`, so no password
  is stored in the tracked Compose file.
- The Docker database is published on `127.0.0.1` only. It cannot be reached
  from other computers on the network.
- The application prints the database host and name, but never the password.
- If a password has been exposed, change it and update `.env`:

  ```sql
  ALTER ROLE library WITH PASSWORD 'new_password';
  ```

## Notes

- This project has no login or access control. It is meant for local use or a
  trusted network.
- The old JSON format is only used by `-import`. Real JSON data files are not
  tracked by Git.
- The frontend files are embedded into the Go binary, so the web app works even
  when the binary is started from another directory.
- `.env` is ignored by Git because it contains the database password.
  `.env.example` is safe to commit.
- Files such as `library`, `library.exe`, and `perpuscli` are build results from
  `go build`, not source files. They are ignored by Git and can be deleted.
