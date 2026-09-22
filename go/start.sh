#!/bin/sh
set -e

echo "=== Menjalankan migrasi database otomatis ke Supabase ==="
if [ -n "$DATABASE_URL" ]; then
    cd /app
    # Pastikan menggunakan direct connection (port 5432) untuk supabase db push
    # Jika DATABASE_URL menggunakan pooler (port 6543), supabase db push bisa gagal
    supabase db push --db-url "$DATABASE_URL" --yes 2>&1 || {
        echo "❌ ERROR: Migrasi database gagal dijalankan!"
        exit 1
    }
    echo "✅ Migrasi database berhasil!"
else
    echo "⚠️ WARNING: DATABASE_URL tidak diset, melewati migrasi."
fi

echo "=== Memulai API server ==="
exec /app/api
