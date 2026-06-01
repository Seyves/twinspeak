#!/usr/bin/env bash

# Clean up containers and reset DB volume (keep go caches)
docker compose -f docker-compose.test.yml down 2>/dev/null || true
docker volume rm twinspeak_postgresdata 2>/dev/null || true

# Run tests
docker compose -f docker-compose.test.yml up --no-log-prefix --attach backend --exit-code-from backend

# Cleanup
docker compose -f docker-compose.test.yml down
