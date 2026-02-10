# ファイルストレージサービス

## 概要

アバター画像などのファイルをローカルファイルシステムに保存・管理するサービスです。

## 機能

- **アバター画像の保存**: ユーザーごとにディレクトリを作成し、画像を保存
- **ファイルサイズ制限**: 最大ファイルサイズを設定可能（デフォルト5MB）
- **拡張子制限**: 許可する拡張子を設定可能（jpg, jpeg, png, gif, webp）
- **セキュリティ**: パストラバーサル攻撃を防ぐ
- **ファイル削除**: 古いアバターを削除（冪等性を保証）
- **URL生成**: 保存したファイルのアクセスURLを生成

## ディレクトリ構造

```
uploads/
└── avatars/
    ├── user-id-1/
    │   ├── 1234567890_abcdef123456.jpg
    │   └── 1234567891_fedcba654321.png
    ├── user-id-2/
    │   └── 1234567892_123456abcdef.jpg
    └── ...
```

## ファイル命名規則

保存されるファイル名は以下の形式になります：

```
{timestamp}_{hash}.{ext}
```

- `timestamp`: Unixタイムスタンプ（秒）
- `hash`: SHA256ハッシュの最初の12文字（衝突回避）
- `ext`: 元のファイルの拡張子

例: `1707552000_a1b2c3d4e5f6.jpg`

## 使い方

### 1. ストレージの初期化

```go
import (
    "github.com/gity/point-system/gateways/infra/infrastorage"
)

cfg := &infrastorage.Config{
    BaseDir:    "./uploads/avatars",      // 保存先ディレクトリ
    BaseURL:    "/uploads/avatars",       // アクセス用URL
    MaxSizeMB:  5,                         // 最大5MB
    AllowedExt: []string{".jpg", ".jpeg", ".png", ".gif", ".webp"},
}

storage, err := infrastorage.NewLocalStorage(cfg)
if err != nil {
    log.Fatal(err)
}
```

### 2. アバター画像の保存

```go
import (
    "io"
    "mime/multipart"
)

// HTTPリクエストからファイルを受け取る
file, header, err := r.FormFile("avatar")
if err != nil {
    // エラー処理
}
defer file.Close()

userID := "user-123"
fileName := header.Filename
fileSize := header.Size

// ファイルを保存
filePath, err := storage.SaveAvatar(userID, fileName, file, fileSize)
if err != nil {
    // エラー処理
    // - ファイルサイズ超過: "exceeds maximum allowed size"
    // - 拡張子エラー: "is not allowed"
    // - その他のエラー
}

// filePathの例: "user-123/1707552000_a1b2c3d4e5f6.jpg"
```

### 3. URL取得

```go
url := storage.GetAvatarURL(filePath)
// url の例: "/uploads/avatars/user-123/1707552000_a1b2c3d4e5f6.jpg"

// データベースに保存する完全なURL
fullURL := fmt.Sprintf("https://example.com%s", url)
```

### 4. アバター画像の削除

```go
// 古いアバターを削除
err := storage.DeleteAvatar(oldFilePath)
if err != nil {
    // エラー処理
}

// 新しいアバターを保存
newFilePath, err := storage.SaveAvatar(userID, newFileName, newFile, newFileSize)
```

## エラーハンドリング

### ファイルサイズ超過

```go
filePath, err := storage.SaveAvatar(userID, fileName, file, fileSize)
if err != nil {
    if strings.Contains(err.Error(), "exceeds maximum allowed size") {
        // ユーザーに「ファイルサイズが大きすぎます」とメッセージを表示
        return errors.New("ファイルサイズは5MB以下にしてください")
    }
}
```

### 許可されていない拡張子

```go
filePath, err := storage.SaveAvatar(userID, fileName, file, fileSize)
if err != nil {
    if strings.Contains(err.Error(), "is not allowed") {
        // ユーザーに「許可されていないファイル形式」とメッセージを表示
        return errors.New("jpg, png, gif, webp形式の画像をアップロードしてください")
    }
}
```

## セキュリティ

### パストラバーサル攻撃の防止

ファイル削除時にパストラバーサル攻撃を防ぐため、パスに `..` が含まれているとエラーになります。

```go
err := storage.DeleteAvatar("../../etc/passwd")
// Error: "invalid file path"
```

### ファイルサイズ制限

リクエストボディ全体を読み込む前にファイルサイズをチェックし、メモリ枯渇攻撃を防ぎます。

### 拡張子ホワイトリスト

許可された拡張子のみを受け付けることで、悪意のあるファイルのアップロードを防ぎます。

## 本番環境での設定

### Nginxの静的ファイル配信設定

```nginx
server {
    listen 80;
    server_name example.com;

    # アバター画像の配信
    location /uploads/avatars/ {
        alias /var/www/uploads/avatars/;
        expires 30d;
        add_header Cache-Control "public, immutable";
    }

    # APIサーバー
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### ディレクトリのパーミッション

```bash
# アップロードディレクトリの作成
mkdir -p /var/www/uploads/avatars

# パーミッション設定（Webサーバーユーザーが書き込み可能）
chown -R www-data:www-data /var/www/uploads
chmod -R 755 /var/www/uploads
```

### Dockerでの設定例

```yaml
# docker-compose.yml
services:
  backend:
    image: point-system-backend
    volumes:
      - ./uploads:/app/uploads
    environment:
      - AVATAR_BASE_DIR=/app/uploads/avatars
      - AVATAR_BASE_URL=/uploads/avatars
      - AVATAR_MAX_SIZE_MB=5
```

## テスト

```bash
# ユニットテストを実行
go test ./tests/unit/infrastorage/local_storage_test.go -v
```

テストカバレッジ:
- ファイル保存（正常系・異常系）
- ファイルサイズ制限
- 拡張子チェック
- ファイル削除（冪等性確認）
- パストラバーサル攻撃の防止
- URL生成
- 統合シナリオ（保存→更新→削除）

## 今後の拡張

将来的に以下の機能を追加する可能性があります：

1. **S3互換ストレージ対応**
   - AWS S3
   - MinIO
   - Google Cloud Storage

2. **画像処理機能**
   - リサイズ（サムネイル生成）
   - 画像形式の変換
   - 画像の最適化

3. **CDN統合**
   - CloudFront
   - Cloudflare

これらの機能を追加する際は、`FileStorageService`インターフェースを拡張し、新しい実装を追加します。
