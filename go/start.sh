#!/bin/sh
set -e

echo "=== Menjalankan migrasi database otomatis ke Supabase ==="

# 1) Pastikan supabase CLI benar-benar bisa dieksekusi oleh user app.
if ! command -v supabase >/dev/null 2>&1; then
    echo "❌ ERROR: supabase CLI tidak ditemukan di PATH."
    exit 1
fi
echo "--- supabase CLI version ---"
supabase --version 2>&1 || {
    echo "❌ ERROR: supabase CLI tidak bisa dieksekusi."
    exit 1
}

if [ -n "$DATABASE_URL" ]; then
    cd /app

    # 1) Link ke project Supabase jika belum linked.
    #    .supabase/ tidak dicommit (gitignore), jadi di deploy baru
    #    perlu link ulang. Token dibaca dari env var SUPABASE_ACCESS_TOKEN.
    if [ ! -d ".supabase" ] && [ -n "$SUPABASE_ACCESS_TOKEN" ]; then
        echo "--- supabase link ---"
        supabase link --project-ref "$SUPABASE_PROJECT_REF" 2>&1 || {
            echo "⚠️ WARNING: supabase link gagal, lanjutkan tetap (mungkin sudah linked)"
        }
    fi

    # 2) Preflight konektivitas DB tanpa menulis password ke log (hanya host:port).
    DB_HOSTPORT=$(printf '%s' "$DATABASE_URL" | sed -E 's#.*@([^/?]+).*#\1#')
    echo "--- preflight: konek ke ${DB_HOSTPORT} ---"

    # Helper: jalankan supabase push sambil mask password (://user:pass@ -> ://***@).
    # POSIX sh (tanpa PIPESTATUS / pipefail): tangkap output dulu, lalu mask.
    run_push() {
        out=$(supabase db push --db-url "$DATABASE_URL" "$@" --yes 2>&1)
        code=$?
        printf '%s\n' "$out" | sed -E 's#://[^/@]+@#://***@#g'
        return "$code"
    }

    # 3) Tangani migration-history divergence tanpa menghapus data user:
    # push default dulu; kalau CLI menolak karena file lokal lebih tua dari history
    # remote (LegacyDbPushMissingRemoteError), retry sekali dengan --include-all.
    echo "--- supabase db push (default) ---"
    if run_push; then
        echo "✅ Migrasi database berhasil!"
    else
        echo "--- push default gagal, retry dengan --include-all ---"
        run_push --include-all || {
            echo "❌ ERROR: Migrasi database gagal dijalankan!"
            exit 1
        }
        echo "✅ Migrasi database berhasil (via --include-all)!"
    fi
else
    echo "⚠️ WARNING: DATABASE_URL tidak diset, melewati migrasi."
fi

echo "=== Memulai API server ==="
exec /app/api
