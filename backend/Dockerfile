# ビルドステージ
FROM golang:1.23-alpine AS builder

WORKDIR /app

# 必要なツールのインストール
RUN apk add --no-cache git

# 依存関係のキャッシュ
COPY go.mod go.sum ./
RUN go mod download

# ソースコードのコピー
COPY . .

# Clean Architecture版のビルド
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/clean_server ./cmd/clean_server

# 本番ステージ
FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

# タイムゾーン設定
ENV TZ=Asia/Tokyo

COPY --from=builder /app/bin/clean_server ./clean_server

EXPOSE 8080

CMD ["./clean_server"]
