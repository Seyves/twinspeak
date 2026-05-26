#!/usr/bin/env bash

# Clean up any existing containers and volumes
docker compose -f docker-compose.test.yml down -v 2>/dev/null || true

# Run tests
docker compose -f docker-compose.test.yml up --exit-code-from backend

# Cleanup
docker compose -f docker-compose.test.yml down -v
