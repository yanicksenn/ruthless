# Verification Instructions

This document outlines the steps to verify recent changes across the Ruthless project. These steps should be performed as part of the verification process for any changes made to the project.

## 🧪 Integration Tests
Verifies cross-service interactions and storage logic.
```bash
bazel test //backend/scripts/integration:integration_test
```

## 🏗️ Build the Whole Stack
Ensure the entire environment is freshly built and up and running.
```bash
docker-compose up --build -d
```