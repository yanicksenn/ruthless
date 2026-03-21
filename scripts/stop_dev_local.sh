#!/usr/bin/env bash
set -e

echo "==> Stopping environment..."
docker-compose -p ruthless-dev down
