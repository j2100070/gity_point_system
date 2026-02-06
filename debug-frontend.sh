#!/bin/bash

# フロントエンドデバッグスクリプト
# Raspberry Piで実行して問題を診断

echo "=========================================="
echo "フロントエンド診断スクリプト"
echo "=========================================="
echo ""

# 1. アーキテクチャ確認
echo "【1. システム情報】"
echo "アーキテクチャ: $(uname -m)"
echo "OS: $(uname -s)"
echo "カーネル: $(uname -r)"
echo ""

# 2. Dockerコンテナの状態
echo "【2. Dockerコンテナの状態】"
docker ps -a | grep frontend
echo ""

# 3. フロントエンドコンテナのログ
echo "【3. フロントエンドコンテナのログ (最新50行)】"
docker logs point_system_frontend --tail 50
echo ""

# 4. コンテナ内部の確認
echo "【4. コンテナ内部のファイル確認】"
docker exec point_system_frontend ls -la /app/ 2>/dev/null || echo "コンテナが起動していません"
echo ""

# 5. node_modulesの確認
echo "【5. node_modules確認】"
docker exec point_system_frontend ls -la /app/node_modules/ 2>/dev/null | head -20 || echo "コンテナが起動していません"
echo ""

# 6. Viteプロセスの確認
echo "【6. Viteプロセス確認】"
docker exec point_system_frontend ps aux 2>/dev/null || echo "コンテナが起動していません"
echo ""

# 7. ポート確認
echo "【7. ポート確認】"
docker exec point_system_frontend netstat -tlnp 2>/dev/null || echo "netstatが利用できません"
echo ""

# 8. ネットワーク確認
echo "【8. ネットワーク接続テスト】"
curl -I http://localhost:5173 2>&1 | head -10
echo ""

# 9. コンテナ内からのテスト
echo "【9. コンテナ内部からのHTTPテスト】"
docker exec point_system_frontend wget -O- http://localhost:5173 2>&1 | head -20 || echo "wgetが利用できません"
echo ""

# 10. 環境変数確認
echo "【10. 環境変数確認】"
docker exec point_system_frontend env | grep VITE 2>/dev/null || echo "コンテナが起動していません"
echo ""

echo "=========================================="
echo "診断完了"
echo "=========================================="
echo ""
echo "次のステップ:"
echo "1. ログにエラーがないか確認"
echo "2. Viteプロセスが起動しているか確認"
echo "3. ポート5173がリッスンしているか確認"
echo ""
