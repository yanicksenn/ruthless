# Ruthless - Cards Against Humanity Clone

Ruthless is a greenfield clone of Cards Against Humanity built in Go (Backend) and React (Frontend). It uses **Bazel** for its build system and **gRPC** for efficient, strongly-typed communication.

## Features

- **Web UI**: A modern, responsive React interface for browsing decks, creating sessions, and playing games.
- **gRPC Backend API**: High-performance gRPC endpoints to manage cards, decks, games, and sessions.
- **Custom Card Decks**: Create personalized decks of white and black cards. The system automatically classifies cards with blanks (e.g. `___`) as Black Cards.
- **Session-Based Games**: Play full game loops tied to specific sessions using your custom decks.
- **Interactive CLI Client**: Manage the server and play full multiplayer games directly from your terminal! Includes features like joining sessions, playing cards, viewing hands, and judging rounds.
- **Pluggable Storage**: Toggle between in-memory storage (for local testing) and PostgreSQL.
- **Pluggable Auth**: Toggle between a `fake` auth mode (no-op for local development) and secure Google `OAuth2` (OIDC) for production.

## Project Structure

The repository is organized as a monorepo:
- [backend/](backend/README.md): Core Go implementation, gRPC API, and server logic.
- [frontend/](frontend/README.md): Modern React web client built with Vite and Tailwind CSS.
- [api/](api/README.md): Protobuf definitions for gRPC services and entities.
- [scripts/](scripts/README.md): Utility scripts for environment management and local deployment.
- [terraform/](terraform/README.md): GCP deployment configurations.
- `secrets/`: Local directory for sensitive credentials.

For detailed instructions on verification and testing, see [GEMINI.md](GEMINI.md).

## Usage

### Docker (Quick Start)

The easiest way to run the entire stack is via the provided scripts:

```bash
# Development (Fake Auth - recommended for local browser testing)
./scripts/run_dev_local.sh

# Production (Google Auth - requires real credentials)
./scripts/run_prod_local.sh

# Stop
./scripts/stop_dev_local.sh
./scripts/stop_prod_local.sh
```

> [!IMPORTANT]
> To verify whether a feature works in the browser through automated tests, you MUST use the `dev` environment (`docker-compose.dev.yml`). The `prod` environment uses Google OAuth which cannot be easily bypassed for testing.

### CLI Client

The CLI tool acts as a full game client! Most commands require a token for identification.
- **In fake auth**: Use any string as a token (e.g., `--token Alice`).
- **In google auth**: Use a real ID Token (see [Backend README](backend/README.md)).

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
