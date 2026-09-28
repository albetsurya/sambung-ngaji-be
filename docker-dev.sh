#!/bin/bash
# Script helper untuk development stack Docker lokal

set -e
cd "$(dirname "$0")"

case "${1:-}" in
  up)
    echo "🚀 Start stack..."
    docker compose up -d
    echo "⏳ Tunggu postgres ready..."
    sleep 5
    docker compose ps
    echo ""
    echo "✅ Backend:  http://localhost:29001"
    echo "✅ Adminer:  http://localhost:8091"
    echo ""
    echo "Login Adminer:"
    echo "  System: PostgreSQL"
    echo "  Server: postgres"
    echo "  User:   pengajian"
    echo "  Pass:   dev_password"
    echo "  DB:     pengajian"
    ;;

  down)
    echo "🛑 Stop stack..."
    docker compose down
    ;;

  restart)
    echo "🔄 Restart api..."
    docker compose restart api
    ;;

  logs)
    docker compose logs -f api
    ;;

  logs-all)
    docker compose logs -f
    ;;

  migrate)
    echo "📦 Apply migration ke postgres lokal..."
    # Baca DATABASE_URL dari .env.docker, tapi ganti host dari 'postgres' jadi 'localhost:5433'
    export DB_URL="postgresql://pengajian:dev_password@localhost:5433/pengajian?sslmode=disable"
    migrate -path go/db/migrations -database "$DB_URL" up
    ;;

  migrate-down)
    export DB_URL="postgresql://pengajian:dev_password@localhost:5433/pengajian?sslmode=disable"
    migrate -path go/db/migrations -database "$DB_URL" down 1
    ;;

  psql)
    echo "🐘 Masuk psql postgres lokal..."
    docker compose exec postgres psql -U pengajian -d pengajian
    ;;

  reset)
    echo "⚠️  Reset database (hapus semua data)..."
    read -p "Yakin? (y/N): " CONFIRM
    if [[ "$CONFIRM" != "y" ]]; then echo "Batal."; exit 1; fi
    docker compose down -v
    docker compose up -d postgres
    sleep 5
    ./docker-dev.sh migrate
    ./docker-dev.sh seed
    ;;

  seed)
    echo "🌱 Seed user albetsurya..."
    cat > /tmp/seed.go <<'GOEOF'
package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	ctx := context.Background()
	url := "postgresql://pengajian:dev_password@localhost:5433/pengajian?sslmode=disable"
	pool, err := pgxpool.New(ctx, url)
	if err != nil { panic(err) }
	defer pool.Close()

	hash, _ := bcrypt.GenerateFromPassword([]byte("albetsurya123"), bcrypt.DefaultCost)
	_, err = pool.Exec(ctx, `
		INSERT INTO users (user_id, username, password_hash, nama, role, status_aktif, created_at, updated_at)
		VALUES ('USR001', 'albetsurya', $1, 'Albet Surya Kembara', 'SUPER_ADMIN', true, now(), now())
		ON CONFLICT (user_id) DO UPDATE SET password_hash = $1
	`, string(hash))
	if err != nil { panic(err) }
	fmt.Println("✅ User albetsurya / albetsurya123 siap")
}
GOEOF
    cd go && go run /tmp/seed.go && rm /tmp/seed.go
    ;;

  *)
    echo "Usage: $0 {up|down|restart|logs|logs-all|migrate|migrate-down|psql|reset|seed}"
    exit 1
    ;;
esac
