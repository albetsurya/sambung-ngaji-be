#!/bin/bash
cd "$(dirname "$0")"

fuser -k 8080/tcp 2>/dev/null

if [ -f .env.local ]; then
    echo "🔧 Pakai .env.local (development)"
    set -a
    source .env.local
    set +a
else
    echo "⚠️  .env.local tidak ada, pakai .env"
    set -a
    source .env
    set +a
fi

go run ./cmd/api
