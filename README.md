# Ruthless - Cards Against Humanity Clone

Ruthless is a greenfield clone of Cards Against Humanity built in Go.

## Features

- **Backend API**: RESTful endpoints to manage cards and game sessions.
- **CLI Client**: Manage the server and play games directly from your terminal.
- **Pluggable Storage**: Toggle between in-memory storage and PostgreSQL.
- **Pluggable Auth**: Toggle between a fake auth mode for local development and real Google OAuth.
- **Card Validation**: Ensures cards contain the required `___` blank placeholder.

## Usage

### Server

Start the server using `memory` storage and `fake` auth (great for local testing):

```bash
go run cmd/cah/main.go server --storage=memory --auth=fake
```

Start the server using `postgres` and `oauth`:

```bash
go run cmd/cah/main.go server --storage=postgres --auth=oauth
```

### CLI

The CLI tool can connect to either a local or remote server using the `--url` flag.

**Create a card:**
```bash
go run cmd/cah/main.go --url http://localhost:8080 cards create --text "A big black ___"
```

## Architecture

- `internal/domain`: Core domain entities (`Card`, `Player`, `Session`)
- `internal/storage`: Storage abstractions and implementations (`memory`, `postgres`)
- `internal/auth`: Authentication abstractions (`fake`, `oauth`)
- `internal/server`: HTTP server routing and handlers
- `internal/cli`: Cobra CLI commands
- `terraform/`: GCP deployment configurations
