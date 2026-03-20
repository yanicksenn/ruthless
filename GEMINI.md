# Verification Instructions

This document outlines the steps to verify recent changes across the Ruthless project. These steps should be performed as part of the verification process for any changes made to the project.

## 🧪 Integration Tests
Verifies cross-service interactions and storage logic.
```bash
bazel test //backend/scripts/integration:integration_test
```

## 🏗️ Build the Whole Stack
Ensure the entire environment is freshly built and up and running.

IMPORTANT: If you want to verify whether a feature works in the browser, YOU MUST ALWAYS user the `docker-compose.dev.yml` file. This is because the `docker-compose.prod.yml` file uses Google OAuth provider and you can't login there.

```bash
docker-compose -p ruthless-dev -f docker-compose.dev.yml up --build -d
```