# トラブルシューティングガイド

## Raspberry Piでフロントエンドが404になる問題

### 問題の可能性リスト

1. **アーキテクチャの問題** - ARM/ARM64でのビルド失敗
2. **メモリ不足** - Raspberry Piのメモリが少ない
3. **ファイルシステムの大文字小文字** - Linuxは区別する
4. **ボリュームマウントの競合** - node_modulesのマウント問題
5. **ネットワーク設定** - ホストバインディングの問題
6. **ポート競合** - 5173が既に使用されている
7. **権限問題** - ファイルパーミッションエラー
8. **依存関係の問題** - ネイティブモジュールのビルドエラー

---

## 診断手順

### 1. デバッグスクリプトの実行

Raspberry Pi上で以下を実行：

```bash
cd /path/to/gity_point_system
./debug-frontend.sh > debug-output.txt
cat debug-output.txt
```

重要なチェックポイント：
- ✅ コンテナが起動しているか
- ✅ Viteプロセスが動いているか
- ✅ ポート5173がリッスンしているか
- ✅ ログにエラーがないか

### 2. ログの確認

```bash
# フロントエンドのログをリアルタイムで確認
docker logs -f point_system_frontend

# 特定のエラーを検索
docker logs point_system_frontend 2>&1 | grep -i error
docker logs point_system_frontend 2>&1 | grep -i failed
```

### 3. コンテナ内部の確認

```bash
# コンテナに入る
docker exec -it point_system_frontend sh

# 内部で確認
ls -la /app
cat /app/package.json
npm list
ps aux
netstat -tlnp
wget -O- http://localhost:5173
exit
```

---

## 解決方法

### 方法1: シンプルな開発用設定を使用（推奨）

```bash
# 既存のコンテナを停止
docker compose down

# 開発用設定で起動
docker compose -f docker-compose.dev.yml up -d --build

# ログを確認
docker logs -f point_system_frontend
```

### 方法2: 元のDockerfileを修正

`docker-compose.yml`を以下のように修正済み：

```yaml
frontend:
  build:
    context: ./frontend
    dockerfile: Dockerfile
    target: development  # 開発ステージを指定
```

### 方法3: メモリ制限を緩和

Raspberry Piのメモリが少ない場合：

```yaml
frontend:
  build:
    context: ./frontend
    dockerfile: Dockerfile.dev
  deploy:
    resources:
      limits:
        memory: 1G  # メモリ制限を設定
```

### 方法4: ボリュームマウントを最小化

`docker-compose.dev.yml`ではソースコードのみをマウント：

```yaml
volumes:
  - ./frontend/src:/app/src:ro  # ソースのみ
  - ./frontend/public:/app/public:ro
  # node_modulesはマウントしない
```

---

## よくあるエラーと解決策

### エラー1: "EADDRINUSE: address already in use"

**原因**: ポート5173が既に使用されている

**解決策**:
```bash
# ポートを使用しているプロセスを確認
sudo lsof -i :5173
# または
sudo netstat -tlnp | grep 5173

# プロセスを停止
sudo kill -9 <PID>
```

### エラー2: "Cannot find module"

**原因**: node_modulesが正しくインストールされていない

**解決策**:
```bash
# コンテナを再ビルド
docker compose down
docker compose build --no-cache frontend
docker compose up -d
```

### エラー3: "permission denied"

**原因**: ファイル権限の問題

**解決策**:
```bash
# ホスト側で権限を修正
cd frontend
sudo chmod -R 755 .
sudo chown -R $USER:$USER .
```

### エラー4: メモリ不足でビルド失敗

**原因**: Raspberry Piのメモリ不足

**解決策**:
```bash
# スワップを増やす
sudo dphys-swapfile swapoff
sudo nano /etc/dphys-swapfile
# CONF_SWAPSIZE=2048 に変更
sudo dphys-swapfile setup
sudo dphys-swapfile swapon
```

### エラー5: ARM アーキテクチャ非対応

**原因**: 一部のnpmパッケージがARMに対応していない

**解決策**:
```dockerfile
# Dockerfile.devで代替アーキテクチャを指定
FROM --platform=linux/arm64 node:20-alpine
```

---

## 動作確認

すべて修正後、以下で動作確認：

```bash
# 1. すべてのコンテナを停止・削除
docker compose down -v

# 2. 開発用設定で再起動
docker compose -f docker-compose.dev.yml up -d --build

# 3. ログを確認
docker logs -f point_system_frontend

# 4. ブラウザでアクセス
# http://<Raspberry-Pi-IP>:5173

# 5. コンテナ内からテスト
docker exec point_system_frontend wget -O- http://localhost:5173
```

---

## 最終手段: ホストマシンで直接実行

Dockerがどうしても動作しない場合：

```bash
cd frontend
npm install
npm run dev -- --host 0.0.0.0
```

---

## サポート情報の収集

問題が解決しない場合、以下の情報を収集：

```bash
# システム情報
uname -a
docker --version
docker compose version

# コンテナ情報
docker ps -a
docker images

# ネットワーク情報
docker network ls
docker network inspect point_system_network

# ログ
docker logs point_system_frontend > frontend.log 2>&1

# デバッグ出力
./debug-frontend.sh > debug.txt 2>&1
```

これらの情報をGitHub Issueに投稿してください。
