# Ruthless - Cards Against Humanity Clone

Ruthless is a greenfield clone of Cards Against Humanity built in Go (Backend) with a planned front-end structure. It uses **Bazel** for its build system and **gRPC** for efficient, strongly-typed communication.

## Features

- **gRPC Backend API**: High-performance gRPC endpoints to manage cards, decks, games, and sessions.
- **Custom Card Decks**: Create personalized decks of white and black cards. The system automatically classifies cards with blanks (e.g. `___`) as Black Cards.
- **Session-Based Games**: Play full game loops tied to specific sessions using your custom decks.
- **Interactive CLI Client**: Manage the server and play full multiplayer games directly from your terminal! Includes features like joining sessions, playing cards, viewing hands, and judging rounds.
- **Pluggable Storage**: Toggle between in-memory storage (for local testing) and PostgreSQL.
- **Pluggable Auth**: Toggle between a `fake` auth mode (no-op for local development) and secure Google `OAuth2` (OIDC) for production.

## Project Structure

The repository is built as a monorepo partitioned into cleanly separated frontend and backend codebases:
- [backend/](backend/README.md): Core Go implementation, gRPC API, and server logic.
- `api/v1`: Protobuf definitions for the gRPC services and entities.
- `frontend/`: (Planned) Directory for the upcoming web client.
- `terraform/`: GCP deployment configurations.
- `secrets/`: Local directory for sensitive credentials.

For detailed technical documentation on authentication, security models, and backend testing, see the [Backend README](backend/README.md).

## Usage

### Docker (Quick Start)

You can run the entire stack via Docker Compose:

```bash
docker-compose up -d
docker-compose down -v
```

To view logs:
```bash
docker-compose logs -f backend
```

### CLI Client

The CLI tool acts as a full game client! Most commands require a token for identification, which can be provided via a flag or a file.
- **Using flags**: Use `--token Alice`. In **fake auth** (default local), use any name.
- **Using files**: Use `--token-file /path/to/token.txt`. The CLI will read the token from the file.
- In **google auth**, use a real ID Token (see Backend README).

#### **Interactive TUI Mode (Recommended)**
```bash
bazel run //backend/cmd/cah -- play interactive --token Alice
```

#### **Manual CLI Commands**
Example of creating a session and playing:
```bash
bazel run //backend/cmd/cah -- play start
bazel run //backend/cmd/cah -- play join <session_id> --name Alice
bazel run //backend/cmd/cah -- game create <session_id> --token Alice
```
Refer to the help command for more options: `bazel run //backend/cmd/cah -- --help`.

## Development

To build the entire project and run all tests, use Bazel:

```bash
bazel build //...
bazel test //...
```
