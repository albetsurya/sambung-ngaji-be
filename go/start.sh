#!/bin/sh
set -e

echo "=== Menjalankan migrasi database otomatis ke Supabase ==="
if [ -n "$DATABASE_URL" ]; then
    cd /app
    supabase db push --db-url "$DATABASE_URL" --yes || {
        echo "❌ ERROR: Migrasi database gagal dijalankan!"
        exit 1
    }
    echo "✅ Migrasi database berhasil!"
else
    echo "⚠️ WARNING: DATABASE_URL tidak diset, melewati migrasi."
fi

echo "=== Memulai API server ==="
exec /app/api
