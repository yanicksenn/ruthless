# Ruthless Frontend

This directory contains the React web client for the Ruthless project. It provides a modern, responsive interface for playing Cards Against Humanity.

## Tech Stack

- **Framework**: [React 19](https://react.dev/)
- **Build Tool**: [Vite 8](https://vitejs.dev/)
- **Styling**: [Tailwind CSS 4](https://tailwindcss.com/)
- **Animations**: [Framer Motion](https://www.framer.com/motion/)
- **Icons**: [Lucide React](https://lucide.dev/)
- **Communication**: [gRPC-web](https://github.com/grpc/grpc-web) via [protobuf-ts](https://github.com/timostamm/protobuf-ts)
- **Auth**: [Google OAuth2](https://github.com/MGrin/react-google-oauth) (for production)

## Getting Started

### Prerequisites

- [Node.js](https://nodejs.org/) (latest LTS recommended)
- [npm](https://www.npmjs.com/)

### Running Locally (Standalone)

You can run the frontend in development mode with HMR:

```bash
cd frontend
npm install
npm run dev
```

The frontend will be available at `http://localhost:3000`. By default, it expects the backend gRPC-web proxy to be running at `http://localhost:8080`.

### Running with Docker

It is recommended to run the frontend along with the rest of the stack using the scripts in the root directory:

```bash
./scripts/run_dev_local.sh
```

## Architecture

- **Protobuf/gRPC**: The frontend uses generated TypeScript clients to communicate with the backend. Code generation is handled by Bazel, but you can see the definitions in the `api/` directory.
- **Routing**: Client-side routing is used to manage navigation between the deck editor, session browser, and active games.
- **State Management**: Uses React hooks and context for managing local and global state (e.g., authentication, active session).

## Development

### Code Generation

If you modify the `.proto` files in the `api/` directory, Bazel will automatically regenerate the TypeScript clients during the next build.

### Linting & Formatting

```bash
npm run lint
```
