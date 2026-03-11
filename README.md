# Ruthless - Cards Against Humanity Clone

Ruthless is a greenfield clone of Cards Against Humanity built in Go (Backend) with a planned front-end structure. It uses **Bazel** for its build system and **gRPC** for efficient, strongly-typed communication.

## Features

- **gRPC Backend API**: High-performance gRPC endpoints to manage cards, decks, games, and sessions.
- **Custom Card Decks**: Create personalized decks of white and black cards. The system automatically classifies cards with blanks (e.g. `___`) as Black Cards.
- **Session-Based Games**: Play full game loops tied to specific sessions using your custom decks.
- **Interactive CLI Client**: Manage the server and play full multiplayer games directly from your terminal! Includes features like joining sessions, playing cards, viewing hands, and judging rounds.
- **Pluggable Storage**: Toggle between in-memory storage (for local testing) and PostgreSQL.
- **Pluggable Auth**: Toggle between a fake auth mode for local development and secure OAuth.

## Project Structure

The repository is built as a monorepo partitioned into cleanly separated frontend and backend codebases:
- `backend/cmd`: Entrypoints for the backend application and CLI.
- `backend/internal`: Core domain logic, CLI commands, gRPC server definitions, and storage adapters.
- `frontend/`: (Planned) Directory for the upcoming web client.
- `api/v1`: Protobuf definitions for the gRPC services and entities.
- `terraform/`: GCP deployment configurations.

## Usage

### Server

Start the gRPC server using `memory` storage and `fake` auth (great for local testing):

```bash
bazel run //backend/cmd/cah -- server --storage=memory --auth=fake
```

#### **Seeding the Server**
You can pre-populate the server (memory storage only) with a JSON seed file containing users, cards, decks, and sessions:

```bash
bazel run //backend/cmd/cah -- server --storage=memory --auth=fake --seed=$(pwd)/seed.json
```

> [!NOTE]
> The `--seed` flag is a **server-side** flag. You must apply it when starting the server, not when running the CLI client commands.

### CLI Gameplay

The CLI tool acts as a full game client! It connects to the server using the `--host` flag (defaults to `localhost:8080`).

#### **Interactive TUI Mode (Recommended)**
For a rich, auto-refreshing experience, use the interactive mode. It allows you to select a session and play through the game loop without manual command typing:

```bash
bazel run //backend/cmd/cah -- play interactive --token Alice
```

#### **Manual CLI Commands**

**1. Create Cards and Decks:**
```bash
bazel run //backend/cmd/cah -- cards create --text "A big black ___"
bazel run //backend/cmd/cah -- decks create --name "My Awesome Deck" --token Alice
bazel run //backend/cmd/cah -- decks add-card <deck_id> <card_id> --token Alice
```

**2. Create a Session and Join:**
```bash
bazel run //backend/cmd/cah -- play start
bazel run //backend/cmd/cah -- play add-deck <session_id> <deck_id>
bazel run //backend/cmd/cah -- play join <session_id> --name Alice
```

**3. Play the Game:**
```bash
bazel run //backend/cmd/cah -- game create <session_id> --token Alice
bazel run //backend/cmd/cah -- game begin <game_id> --token Alice
bazel run //backend/cmd/cah -- game status <game_id> --token Alice
bazel run //backend/cmd/cah -- game hand <game_id> --token Alice
bazel run //backend/cmd/cah -- game play-cards <game_id> <white_card_id> --token Alice
bazel run //backend/cmd/cah -- game judge <game_id> <play_id> --token Alice (If you are the Czar)
```

## Development

To build the entire project and run all tests, use Bazel:

```bash
bazel build //...
bazel test //...
```

If you add new dependencies or modify Go imports, remember to update the Bazel configurations via Gazelle:

```bash
bazel run //:gazelle-update-repos
bazel run //:gazelle
```
