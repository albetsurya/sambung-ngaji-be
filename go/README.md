# 🕌 Sambung Ngaji — Backend API

Backend API untuk aplikasi manajemen pengajian **Sambung Ngaji**. Dibangun dengan **Go 1.26**, **Fiber v2**, **PostgreSQL (pgx)**, **JWT**, dan **zerolog**.

## ✨ Fitur Utama

### 🔐 Autentikasi & Otorisasi
- **JWT-based auth** dengan access token (TTL 12 jam) & refresh token
- **Role-based access control (RBAC)**:
  - `SUPER_ADMIN` — akses penuh sistem
  - `ADMIN` — kelola jamaah, kelompok, absensi, pengumuman, pengaturan
  - `TIM_PNKB` — khusus pembinaan pra nikah (monitoring)
  - `TIM_ABSENSI` — khusus absensi pengajian
- **Session management** dengan validasi token & blacklist logout
- **Audit logging** untuk aktivitas sensitif (login, CRUD data master, dll)

### 👥 Manajemen Jamaah (Members)
- CRUD lengkap biodata jamaah (identitas, kontak, alamat, pendidikan, dll)
- **Kategori otomatis** berdasarkan usia & status pernikahan:
  - `BALITA`, `CABERAWIT`, `PRA_REMAJA`, `REMAJA`, `PRA_NIKAH`, `DEWASA`, `ISTIMEWA`
- Upload foto jamaah ke **Supabase Storage** (kompresi otomatis via `browser-image-compression`)
- Import massal dari CSV/Excel
- **Import dari teks** (WhatsApp/PDF) dengan parsing otomatis
- Pencarian, filter, pagination, virtualisasi list
- Export CSV

### 📅 Jadwal & Absensi
- CRUD jadwal pengajian (tanggal, jam, tempat, kategori target, gender target)
- **Bulk create jadwal** berdasarkan hari & bulan
- Absensi dengan status: `HADIR`, `IJIN`, `SAKIT`, `ALPA`
- Bulk action "Tandai semua hadir"
- Progress kehadiran real-time
- Riwayat absensi per jamaah & per jadwal
- Picker jadwal dengan aksi cepat (edit, hapus, lihat absensi)

### 💬 Monitoring & Pembinaan
- Catatan monitoring per jamaah (teks bebas + kategori)
- **Status pembinaan otomatis**: `AKTIF`, `PERLU_PERHATIAN`, `KURANG_AKTIF`, `TIDAK_AKTIF`
- Timeline aktivitas jamaah
- Dashboard khusus **Tim PNKB** (pra nikah)
- Rekap monitoring per periode

### 📢 Pengumuman & Template
- Template pengumuman dengan placeholder dinamis
- Generate otomatis dari template + data jadwal/jamaah
- Preview sebelum kirim
- Kirim ke WhatsApp (via WA Gateway)
- Riwayat pengumuman per kelompok

### 🤖 AI Assistant
- **Multi-provider**: OmniRoute (utama), Gemini, Groq (fallback otomatis)
- **Tool calling** untuk akses data real-time (jamaah, jadwal, absensi, monitoring)
- Chat history tersimpan di database
- Usage tracking & quota per user
- Streaming response (SSE)

### 📚 Fitur Member (Jamaah)
- **Beranda personal** dengan jadwal terdekat, status absensi, statistik
- **Doa harian** (pagi, sore, sebelum/sesudah tidur, dll)
- **Dzikir** (pagi, sore, setelah sholat, counter)
- **Sholat** (jadwal, arah kiblat, jurnal sholat)
- **Puasa** (sunah, wajib, catatan)
- **Tahfidz** (target surah, progress, ayat)
- **Al-Quran** (mushaf, terjemah, bookmark, pencarian)
- **Mood tracker** & **Privasi data**
- Edit profil & ganti password

### ⚙️ Admin & Pengaturan
- Manajemen pengguna (CRUD, role, status aktif)
- Manajemen kelompok/pengajian
- Pengaturan aplikasi (nama, zona waktu, dll)
- **Audit log** dengan filter & pencarian
- **AI Usage** monitoring (token, cost, quota)

## 🚀 Tech Stack

| Teknologi | Versi | Fungsi |
|-----------|-------|--------|
| **Go** | 1.26 | Language runtime |
| **Fiber** | v2.52 | Web framework |
| **pgx** | v5.11 | PostgreSQL driver & pool |
| **JWT-Go** | v5.3 | Token authentication |
| **zerolog** | v1.35 | Structured logging |
| **Supabase Go** | - | Storage (foto jamaah) |
| **Google AI** | - | Gemini provider |
| **Groq** | - | Groq provider |
| **OmniRoute** | - | AI gateway provider |

## 📁 Struktur Project

```
backend/go/
├── cmd/
│   ├── api/           # Entry point HTTP server
│   └── import/        # CLI import data massal (CSV)
├── internal/
│   ├── ai/            # AI providers (OmniRoute, Gemini, Groq) + tools
│   ├── api/           # HTTP handlers, middleware, dispatcher
│   ├── auth/          # JWT, password hashing, permissions
│   ├── config/        # Config loader (env-based)
│   ├── database/      # DB connection pool
│   ├── errors/        # Sentinel errors & wrapping helpers
│   ├── handler/       # Health check
│   ├── model/         # Domain models & DTOs
│   ├── repository/    # Data access layer (PostgreSQL)
│   └── service/       # Business logic layer
├── db/
│   └── migration/     # SQL migrations (up/down)
├── Dockerfile         # Multi-stage production build
├── Dockerfile.dev     # Development build
├── go.mod / go.sum
└── run-dev.sh         # Dev runner dengan .env.local
```

## 🔧 Konfigurasi

Salin `.env.example` ke `.env` (atau `.env.local` untuk development):

```bash
# App
APP_ENV=development
APP_PORT=8080

# Database
DATABASE_URL=postgres://user:pass@localhost:5432/sambung_ngaji?sslmode=disable

# JWT
JWT_SECRET=your-super-secret-key
JWT_EXPIRY_HOURS=12

# AI Providers
AI_PROVIDER=omniroute
OMNIROUTE_ENDPOINT=https://your-omniroute-endpoint
OMNIROUTE_API_KEY=your-key
OMNIROUTE_MODEL=auto/best-vision
OMNIROUTE_TIMEOUT_SEC=60
GROQ_API_KEY=your-groq-key
GROQ_MODEL=openai/gpt-oss-120b
GEMINI_API_KEY=your-gemini-key
GEMINI_MODEL=gemini-2.5-flash

# Supabase Storage (foto jamaah)
SUPABASE_URL=https://xxx.supabase.co
SUPABASE_SERVICE_ROLE_KEY=your-service-role-key
SUPABASE_BUCKET=jamaah-photos

# DB Pool Tuning (optional)
DB_MAX_CONNS=10
DB_MIN_CONNS=1
DB_MAX_CONN_LIFETIME=1h
DB_MAX_CONN_IDLE_TIME=30m
DB_HEALTH_CHECK_PERIOD=30s
```

## 🏃 Menjalankan

### Development
```bash
cd backend/go
./run-dev.sh
# atau manual:
source .env.local  # jika ada
go run ./cmd/api
```

### Production (Docker)
```bash
cd backend/go
docker build -t sambung-ngaji-backend .
docker run -d -p 8080:8080 --env-file .env sambung-ngaji-backend
```

### Health Check
```bash
curl http://localhost:8080/health   # Liveness
curl http://localhost:8080/ready    # Readiness (DB connectivity + pool stats)
```

## 📦 API Endpoints (Ringkas)

| Area | Endpoint | Method | Roles |
|------|----------|--------|-------|
| Auth | `/api/login` | POST | Public |
| Auth | `/api/logout` | POST | All |
| Members | `/api/members` | GET/POST | Admin+ |
| Members | `/api/members/:id` | GET/PUT/DELETE | Admin+ |
| Members | `/api/members/import-text` | POST | Admin+ |
| Attendance | `/api/attendance` | GET/POST | Tim Absensi+ |
| Attendance | `/api/attendance/bulk` | POST | Tim Absensi+ |
| Schedule | `/api/meetings` | GET/POST | Admin+ |
| Schedule | `/api/meetings/bulk` | POST | Admin+ |
| Monitoring | `/api/monitoring` | GET/POST | Tim PNKB+ |
| Announcements | `/api/announcements` | GET/POST | Admin+ |
| AI | `/api/ai/chat` | POST | All (quota) |
| AI | `/api/ai/usage` | GET | Admin+ |
| Settings | `/api/settings` | GET/PUT | Super Admin |

> Format request/response: JSON dengan envelope `{ success: bool, data: T, message: string }`

## 🧪 Testing & Linting

```bash
# Format & build
gofmt -w .
go build ./...

# Vet & static analysis
go vet ./...
staticcheck ./...

# Security scan
govulncheck ./...

# Tests (kalau ada)
go test ./... -race
```

## 📊 Observability

- **Structured logging** (zerolog) dengan correlation ID (`X-Request-ID`)
- **Prometheus metrics**: `http_requests_total`, `http_request_duration_seconds`, `http_requests_in_flight`
- **Health/Readiness probes** untuk Kubernetes
- **Graceful shutdown** dengan drain connections (15s timeout)

## 🔒 Keamanan

- Input validation di semua handler (`ValidateBody` middleware)
- Rate limiting per IP (100 req/menit default)
- CORS dikonfigurasi eksplisit
- Password hashing dengan bcrypt (cost 12)
- JWT secret minimal 32 char di production
- Prepared statements (pgx) mencegah SQL injection
- Field-level visibility berbasis role

## 📝 Migrasi Database

```bash
# Jalankan migrasi (manual atau via tool)
psql -d sambung_ngaji -f db/migration/001_init.up.sql
psql -d sambung_ngaji -f db/migration/002_xxx.up.sql
# Rollback:
psql -d sambung_ngaji -f db/migration/002_xxx.down.sql
```

## 📄 Lisensi

MIT License — lihat [LICENSE](../LICENSE) untuk detail.