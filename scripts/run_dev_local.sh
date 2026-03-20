#!/usr/bin/env bash
set -e

echo "==> Building and loading backend image..."
bazel run --platforms=@rules_go//go/toolchain:linux_amd64 //backend:tarball

echo "==> Building and loading frontend image..."
bazel run --platforms=@rules_go//go/toolchain:linux_amd64 //frontend:tarball

echo "==> Starting environment..."
docker-compose -p ruthless-dev -f docker-compose.dev.yml up -d
