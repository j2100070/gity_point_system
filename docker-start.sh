#!/bin/bash

# Point System Docker Compose 起動スクリプト

set -e

echo "🚀 Point System を起動しています..."
echo ""

# カラー定義
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Docker と Docker Compose の確認
if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker がインストールされていません${NC}"
    exit 1
fi

if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo -e "${RED}❌ Docker Compose がインストールされていません${NC}"
    exit 1
fi

# Docker Compose コマンドの決定
if docker compose version &> /dev/null; then
    DOCKER_COMPOSE="docker compose"
else
    DOCKER_COMPOSE="docker-compose"
fi

# 古いコンテナを停止
echo -e "${YELLOW}📦 既存のコンテナを停止しています...${NC}"
$DOCKER_COMPOSE down

# イメージをビルド
echo -e "${BLUE}🔨 Docker イメージをビルドしています...${NC}"
$DOCKER_COMPOSE build --no-cache

# コンテナを起動
echo -e "${GREEN}🚀 コンテナを起動しています...${NC}"
$DOCKER_COMPOSE up -d

# 起動を待機
echo -e "${YELLOW}⏳ サービスの起動を待っています...${NC}"
sleep 5

# ステータス確認
echo ""
echo -e "${GREEN}✅ Point System が起動しました！${NC}"
echo ""
echo "📊 サービス一覧:"
echo "  - データベース: http://localhost:5432"
echo "  - バックエンドAPI: http://localhost:8080"
echo "  - フロントエンド: http://localhost:5173"
echo ""
echo "👤 初期ユーザー:"
echo "  - 管理者: admin / admin123"
echo "  - テストユーザー: testuser / test123"
echo ""
echo "📝 ログを表示:"
echo "  $ docker compose logs -f"
echo ""
echo "🛑 停止:"
echo "  $ docker compose down"
echo ""

# ログを表示
read -p "ログを表示しますか？ (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    $DOCKER_COMPOSE logs -f
fi
