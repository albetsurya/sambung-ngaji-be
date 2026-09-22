#!/bin/sh
set -e

# Auto-migrate sebelum mulai
echo "Menjalankan migrasi database otomatis ke Supabase..."
if [ -n "$DATABASE_URL" ]; then
    cd /app
    supabase db push --db-url "$DATABASE_URL" || echo "⚠️ Migrasi gagal atau sudah up-to-date, melanjutkan..."
else
    echo "⚠️ DATABASE_URL tidak diset, melewati migrasi."
fi

# Jalankan aplikasi
echo "Memulai API..."
exec /app/api
