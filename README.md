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
- `backend/cmd`: Entrypoints for the backend application and CLI.
- `backend/internal`: Core domain logic, CLI commands, gRPC server definitions, and storage adapters.
- `frontend/`: (Planned) Directory for the upcoming web client.
- `api/v1`: Protobuf definitions for the gRPC services and entities.
- `terraform/`: GCP deployment configurations.
- `secrets/`: Local directory for sensitive credentials (e.g., Google OAuth Client Secrets).

## Authentication

Ruthless supports two primary authentication modes, toggled via the `--auth` flag on the server:

### 1. Fake Auth (`--auth=fake`)
Used for local development and automated testing. Any string passed in the `Authorization` header is treated as the player's name (e.g., `--token Alice`). No password or external validation is performed.

### 2. Google OAuth (`--auth=google`)
Production-ready OIDC validation. The server validates tokens against Google's public keys.

**Requirements**:
- A `--google-audience` (The Google Client ID for your app).
- An internet connection for the server to fetch Google's JWKS.

#### **Obtaining an ID Token**
To get a real ID token for manual testing or gameplay:
```bash
bazel run //backend/cmd/cah -- token login --callback-port 9999
```
This will open your browser, allow you to log in with your Google account, and print a JWT `ID Token` to the terminal.

## Security Model

Ruthless enforces a strict security model to ensure fair play and data integrity:

- **Enforced Registration**: Every player must be registered in the database before they can access authenticated services. Even with a valid OAuth token, the server will reject requests with `PermissionDenied` if the sub (user ID) is not found in the `users` table. Use `UserService.Register` to create your profile.
- **Resource Ownership**:
  - **Cards**: All cards have an owner. Only the creator of a card can add it to a deck.
  - **Decks**: Decks have an owner and optional contributors. Only authorized users can modify a deck's metadata or card list.
  - **Sessions**: The session owner (creator) is the only one who can add decks or start the game.
- **Role Enforcement**: During gameplay, the system strictly enforces roles. Only the **Czar** can select a winner, and the Czar is prohibited from playing white cards in their own round.

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

### CLI Client & Tokens
Most CLI commands require a `--token` to identify the player.
- In **fake auth**, use any name: `--token Alice`.
- In **google auth**, use a real ID Token obtained via `token login`: `--token <long_jwt_string>`.

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

## Testing

Ruthless maintains a comprehensive test suite to ensure backend reliability and security. We use a dual-mode testing strategy:

### 1. Automated Tests (Integration & E2E)
These tests are fully automated and run against an **in-process gRPC server** with memory storage. They use a permanent refresh token for Alice to authenticate via Google OAuth, even if no backend is running externally.

**Requirements**:
- A Google Client Secret JSON in `secrets/client_secret_dev.json`.
- Alice's refresh token in `secrets/ruthless.alice.sec`.

**Running Integration Tests**:
```bash
bazel test //backend/scripts/integration:integration_test
```

**Running E2E Validation**:
```bash
bazel test //backend/scripts/e2e:e2e_test
```

### 2. Interactive Registration Tests
The registration flow is tested separately through an interactive suite that prompts for manual OAuth login against an **active backend**.

**Running Registration Tests**:
```bash
# 1. Start your backend (e.g. via Docker or locally)
bazel run //backend/cmd/cah -- server --storage=postgres --auth=google

# 2. Run the interactive test
bazel run //backend/scripts/registration:registration_test -- --addr=localhost:8080 -v --nocache
```
Follow the URL in the terminal to log in. The test will automatically capture the token and verify the registration on the server.

### 3. Docker Testing
You can run the entire stack via Docker Compose:

```bash
docker-compose up -d
```
Then run tests against the active container by providing the `--addr` flag:
```bash
bazel test //backend/scripts/integration:integration_test -- --addr=localhost:8080
```

### Docker
 
 You can also run the backend directly as a Docker container.
 
#### **Build the Backend Image**
```bash
docker build -t ruthless-backend -f backend/Dockerfile .
```
 
#### **Environment Variables**
You can customize the deployment via variables:
- `AUTH`: `google` (default) or `fake`.
- `GOOGLE_AUDIENCE`: Your Google Client ID.

```bash
# Example: Run with fake auth for testing
AUTH=fake docker-compose up -d
```

To view logs:
```bash
docker-compose logs -f backend
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
