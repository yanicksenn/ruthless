# Verification Instructions

This document outlines the steps to verify recent changes across the Ruthless project. These steps should be performed as part of the verification process for any changes made to the project.

## 🏗️ Build the Whole Stack
Ensure the entire environment is freshly built and up and running.
```bash
docker-compose up --build -d
```

## 🧪 Interactive Tests
These tests interact with the running stack and require the services to be healthy.

### Registration Test
Verifies the user registration flow. (Requires user input).
```bash
bazel run //backend/scripts/registration:registration_test -- --addr=localhost:8080 -v --nocache
```

### Integration Test
Verifies cross-service interactions and storage logic. (Requires user input).
```bash
bazel test //backend/scripts/integration:integration_test -- --addr=localhost:8080
```

### End-to-End (E2E) Test
Full system flow verification. (Requires user input).
```bash
bazel test //backend/scripts/e2e:e2e_test -- --addr=localhost:8080
```
