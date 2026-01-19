#!/bin/bash

# ================================================
# Gity Point System API テストスクリプト
# ================================================
# 使用方法:
#   chmod +x test_api.sh
#   ./test_api.sh
# ================================================

set -e

API_URL="http://localhost:8080"
COOKIE_FILE="cookies.txt"
CSRF_TOKEN=""

# 色付き出力
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# ヘッダー表示
print_header() {
    echo ""
    echo -e "${BLUE}======================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}======================================${NC}"
}

# テスト結果表示
print_result() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓ PASS${NC}: $2"
    else
        echo -e "${RED}✗ FAIL${NC}: $2"
        echo -e "${RED}Response: $3${NC}"
    fi
}

# クッキーファイルの初期化
rm -f $COOKIE_FILE

# ================================================
# 1. ヘルスチェック
# ================================================
print_header "1. ヘルスチェック"

response=$(curl -s -w "\n%{http_code}" "$API_URL/health")
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | head -n-1)

if [ "$http_code" -eq 200 ]; then
    print_result 0 "ヘルスチェック成功"
    echo "Response: $body"
else
    print_result 1 "ヘルスチェック失敗" "$body"
    exit 1
fi

# ================================================
# 2. ユーザー登録
# ================================================
print_header "2. ユーザー登録"

response=$(curl -s -w "\n%{http_code}" -X POST "$API_URL/api/auth/register" \
    -H "Content-Type: application/json" \
    -d '{
        "username": "testuser1",
        "email": "test1@example.com",
        "password": "Test@123456",
        "display_name": "Test User 1"
    }')

http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | head -n-1)

if [ "$http_code" -eq 201 ]; then
    print_result 0 "ユーザー登録成功"
    echo "Response: $body"
else
    print_result 1 "ユーザー登録失敗" "$body"
fi

# ================================================
# 3. ログイン（既存ユーザー: user1）
# ================================================
print_header "3. ログイン"

response=$(curl -s -w "\n%{http_code}" -X POST "$API_URL/api/auth/login" \
    -H "Content-Type: application/json" \
    -c $COOKIE_FILE \
    -d '{
        "username": "user1",
        "password": "User@123456"
    }')

http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | head -n-1)

if [ "$http_code" -eq 200 ]; then
    print_result 0 "ログイン成功"
    echo "Response: $body"
    # CSRFトークンを抽出
    CSRF_TOKEN=$(echo "$body" | grep -o '"csrf_token":"[^"]*"' | cut -d'"' -f4)
    echo -e "${YELLOW}CSRF Token: $CSRF_TOKEN${NC}"
else
    print_result 1 "ログイン失敗" "$body"
    exit 1
fi

# ================================================
# 4. 現在のユーザー情報取得
# ================================================
print_header "4. 現在のユーザー情報取得"

response=$(curl -s -w "\n%{http_code}" "$API_URL/api/auth/me" \
    -b $COOKIE_FILE)

http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | head -n-1)

if [ "$http_code" -eq 200 ]; then
    print_result 0 "ユーザー情報取得成功"
    echo "Response: $body"
else
    print_result 1 "ユーザー情報取得失敗" "$body"
fi

# ================================================
# 5. 残高確認
# ================================================
print_header "5. 残高確認"

response=$(curl -s -w "\n%{http_code}" "$API_URL/api/points/balance" \
    -b $COOKIE_FILE)

http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | head -n-1)

if [ "$http_code" -eq 200 ]; then
    print_result 0 "残高確認成功"
    echo "Response: $body"
else
    print_result 1 "残高確認失敗" "$body"
fi

# ================================================
# 6. ポイント転送（CSRFトークン必要）
# ================================================
print_header "6. ポイント転送"

# user2のIDを取得する必要があるため、ここではスキップ
# 実際のテストではuser2のUUIDを指定する必要があります

echo -e "${YELLOW}Note: ポイント転送はユーザーUUIDが必要なためスキップ${NC}"
echo "実行例:"
echo 'curl -X POST "$API_URL/api/points/transfer" \'
echo '    -H "Content-Type: application/json" \'
echo '    -H "X-CSRF-Token: $CSRF_TOKEN" \'
echo '    -b $COOKIE_FILE \'
echo '    -d '"'"'{'
echo '        "to_user_id": "USER_UUID_HERE",'
echo '        "amount": 100,'
echo '        "idempotency_key": "unique-key-'$(date +%s)'",'
echo '        "description": "Test transfer"'
echo '    }'"'"

# ================================================
# 7. QRコード生成（受取用）
# ================================================
print_header "7. QRコード生成（受取用）"

response=$(curl -s -w "\n%{http_code}" -X POST "$API_URL/api/qr/generate/receive" \
    -H "Content-Type: application/json" \
    -H "X-CSRF-Token: $CSRF_TOKEN" \
    -b $COOKIE_FILE \
    -d '{
        "amount": 500
    }')

http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | head -n-1)

if [ "$http_code" -eq 200 ]; then
    print_result 0 "QRコード生成成功"
    echo "Response: $body"
else
    print_result 1 "QRコード生成失敗" "$body"
fi

# ================================================
# 8. トランザクション履歴
# ================================================
print_header "8. トランザクション履歴"

response=$(curl -s -w "\n%{http_code}" "$API_URL/api/points/history?limit=10" \
    -b $COOKIE_FILE)

http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | head -n-1)

if [ "$http_code" -eq 200 ]; then
    print_result 0 "トランザクション履歴取得成功"
    echo "Response: $body"
else
    print_result 1 "トランザクション履歴取得失敗" "$body"
fi

# ================================================
# 9. 友達一覧
# ================================================
print_header "9. 友達一覧"

response=$(curl -s -w "\n%{http_code}" "$API_URL/api/friends" \
    -b $COOKIE_FILE)

http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | head -n-1)

if [ "$http_code" -eq 200 ]; then
    print_result 0 "友達一覧取得成功"
    echo "Response: $body"
else
    print_result 1 "友達一覧取得失敗" "$body"
fi

# ================================================
# 10. ログアウト
# ================================================
print_header "10. ログアウト"

response=$(curl -s -w "\n%{http_code}" -X POST "$API_URL/api/auth/logout" \
    -b $COOKIE_FILE)

http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | head -n-1)

if [ "$http_code" -eq 200 ]; then
    print_result 0 "ログアウト成功"
    echo "Response: $body"
else
    print_result 1 "ログアウト失敗" "$body"
fi

# ================================================
# 管理者機能テスト（adminでログイン）
# ================================================
print_header "11. 管理者ログイン"

response=$(curl -s -w "\n%{http_code}" -X POST "$API_URL/api/auth/login" \
    -H "Content-Type: application/json" \
    -c $COOKIE_FILE \
    -d '{
        "username": "admin",
        "password": "Admin@123456"
    }')

http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | head -n-1)

if [ "$http_code" -eq 200 ]; then
    print_result 0 "管理者ログイン成功"
    echo "Response: $body"
    CSRF_TOKEN=$(echo "$body" | grep -o '"csrf_token":"[^"]*"' | cut -d'"' -f4)
else
    print_result 1 "管理者ログイン失敗" "$body"
fi

# ================================================
# 12. 全ユーザー一覧（管理者のみ）
# ================================================
print_header "12. 全ユーザー一覧（管理者）"

response=$(curl -s -w "\n%{http_code}" "$API_URL/api/admin/users?limit=10" \
    -b $COOKIE_FILE)

http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | head -n-1)

if [ "$http_code" -eq 200 ]; then
    print_result 0 "ユーザー一覧取得成功"
    echo "Response: $body"
else
    print_result 1 "ユーザー一覧取得失敗" "$body"
fi

# ================================================
# 13. 全トランザクション一覧（管理者のみ）
# ================================================
print_header "13. 全トランザクション一覧（管理者）"

response=$(curl -s -w "\n%{http_code}" "$API_URL/api/admin/transactions?limit=10" \
    -b $COOKIE_FILE)

http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | head -n-1)

if [ "$http_code" -eq 200 ]; then
    print_result 0 "トランザクション一覧取得成功"
    echo "Response: $body"
else
    print_result 1 "トランザクション一覧取得失敗" "$body"
fi

# クリーンアップ
rm -f $COOKIE_FILE

print_header "テスト完了"
echo -e "${GREEN}全てのテストが完了しました${NC}"
