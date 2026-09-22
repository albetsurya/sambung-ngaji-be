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

# Final stage
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata wget curl

# Install Supabase CLI
RUN wget -qO- https://github.com/supabase/cli/releases/latest/download/supabase_linux_amd64.tar.gz | tar -xz -C /usr/local/bin

ENV TZ=Asia/Jakarta

WORKDIR /app

COPY --from=builder /out/api /app/api
COPY start.sh /app/start.sh
COPY supabase /app/supabase
RUN chmod +x /app/start.sh

RUN addgroup -g 1000 app && adduser -D -u 1000 -G app app
USER app

ENTRYPOINT ["/app/start.sh"]
