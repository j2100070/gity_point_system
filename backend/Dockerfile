# ビルドステージ
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 依存関係のキャッシュ
COPY go.mod go.sum ./
RUN go mod download

# ソースコードのコピー
COPY . .

# ビルド
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# 実行ステージ
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# ビルド成果物のコピー
COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations

# ポート公開
EXPOSE 8080

# 実行
CMD ["./main"]
