FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata wget

ENV TZ=Asia/Jakarta

WORKDIR /app

COPY --from=builder /out/api /app/api

RUN addgroup -g 1000 app && adduser -D -u 1000 -G app app
USER app

EXPOSE 8080

CMD ["/app/api"]
