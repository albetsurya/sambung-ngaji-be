# Build stage
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build

# Download dependencies first (cached layer)
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w -buildid=" -o /out/api ./cmd/api

# Final stage — debian-slim (glibc) karena Supabase CLI adalah binary glibc
# dan TIDAK bisa jalan di alpine (musl): Error relocating __pthread_key_create, rc=127.
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates tzdata curl \
    && rm -rf /var/lib/apt/lists/*

# Install Supabase CLI (glibc binary berjalan native di debian-slim)
RUN curl -sL https://github.com/supabase/cli/releases/latest/download/supabase_linux_amd64.tar.gz | tar -xz -C /usr/local/bin \
    && chmod +x /usr/local/bin/supabase \
    && supabase --version

ENV TZ=Asia/Jakarta

WORKDIR /app

COPY --from=builder /out/api /app/api
COPY start.sh /app/start.sh
COPY supabase /app/supabase
RUN chmod +x /app/start.sh

RUN groupadd -g 1000 app && useradd -u 1000 -g app -m -s /bin/sh app
USER app

ENTRYPOINT ["/app/start.sh"]
