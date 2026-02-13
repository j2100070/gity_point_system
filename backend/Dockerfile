# ビルドステージ
FROM golang:1.23-alpine AS builder

WORKDIR /app

# 依存関係のキャッシュ
COPY go.mod go.sum ./
RUN go mod download

# ソースコードのコピー
COPY . .

# ビルド
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/server ./cmd/clean_server

# 本番ステージ
FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/bin/server ./server

EXPOSE 8080

CMD ["./server"]
