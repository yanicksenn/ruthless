# Ruthless API

This directory contains the Protocol Buffer definitions for the Ruthless project. These definitions serve as the single source of truth for the gRPC services and entities used by both the backend and frontend.

## Structure

- **v1/**: Contains the version 1 API definitions.
    - **ruthless.proto**: Main service definitions (CardService, DeckService, SessionService, UserService, GameService) and core entities.
    - **config.proto**: Configuration-related message definitions.

## Code Generation

The Ruthless project uses **Bazel** to manage code generation, ensuring that all components are always in sync with the latest proto definitions.

### Go (Backend)

The Go code is generated using `rules_go`. You can find the generated code in the Bazel output directory, but it is transparently available to the backend Go packages.

### TypeScript (Frontend)

The TypeScript code is generated using `protobuf-ts`. Similar to the backend, Bazel handles the generation and provides the generated files to the frontend build process.

## Adding/Modifying APIs

1.  Modify the `.proto` files in `api/v1/`.
2.  If you add new files, update the `BUILD.bazel` file in the same directory.
3.  Run `bazel build //...` to ensure everything compiles and code is regenerated.
