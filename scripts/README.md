# Ruthless Scripts

This directory contains utility scripts for managing the Ruthless development and production environments locally.

## Environment Management

These scripts use Docker Compose to orchestrate the backend, frontend, and database containers.

### Development Environment (Fake Auth)

Best for local development and UI testing. Auto-creates user profiles on login.

- **run_dev_local.sh**: Builds the backend and frontend images using Bazel and starts the environment using `docker-compose.dev.yml`.
- **stop_dev_local.sh**: Stops the development environment and removes volumes.

### Production-like Environment (Google Auth)

Used for testing the full authentication flow with real Google OAuth2 credentials.

- **run_prod_local.sh**: Builds the images and starts the environment using `docker-compose.prod.yml`.
- **stop_prod_local.sh**: Stops the production-like environment and removes volumes.

## Usage

From the root of the repository:

```bash
./scripts/run_dev_local.sh
```

Ensure the scripts have execution permissions:

```bash
chmod +x scripts/*.sh
```
