# クリーンアーキテクチャ実装ガイド

## アーキテクチャ概要

このプロジェクトは、クリーンアーキテクチャの原則に従って設計されています。各レイヤーは明確な責務を持ち、依存関係は外側から内側に向かって一方向に流れます。

## ディレクトリ構造

```
backend/
├── entities/                    # ビジネスロジックを持つエンティティ
│   ├── user.go
│   ├── transaction.go
│   ├── friendship.go
│   ├── qrcode.go
│   ├── session.go
│   └── logger.go               # 全レイヤーから参照されるinterface
├── usecases/
│   ├── inputport/              # ユースケースのinterface定義
│   │   ├── auth_inputport.go
│   │   └── point_transfer_inputport.go
│   ├── interactor/             # ユースケースの実装
│   │   └── point_transfer_interactor.go
│   └── repository/             # リポジトリのinterface定義
│       ├── user_repository.go
│       ├── transaction_repository.go
│       ├── friendship_repository.go
│       ├── qrcode_repository.go
│       └── session_repository.go
├── gateways/
│   ├── repository/             # リポジトリ実装
│   │   ├── user/
│   │   │   └── user_repository_impl.go
│   │   ├── transaction/
│   │   └── datasource/         # DataSourceのinterface定義
│   │       └── dsmysql/
│   │           └── user_datasource.go
│   ├── datasource/             # DataSource実装
│   │   └── dsmysqlimpl/
│   │       └── user_datasource_impl.go
│   └── infra/                  # インフラストラクチャ実装
│       ├── inframysql/
│       │   └── postgres.go
│       ├── infralogger/
│       │   └── logger.go
│       └── inframemcached/
├── controllers/
│   ├── web/
│   │   ├── point_controller.go
│   │   └── presenter/          # Presenterの実装
│   │       └── point_presenter.go
│   └── daemon/                 # 非同期処理用
├── frameworks/
│   ├── web/
│   │   ├── router.go
│   │   ├── time_provider.go    # 時刻情報の提供
│   │   └── middleware/
│   │       ├── auth.go
│   │       ├── csrf.go
│   │       └── security.go
│   └── daemon/
├── cmd/
│   ├── server/
│   │   ├── main.go
│   │   └── main_clean.go
│   └── wire/                   # DIライブラリ定義
│       ├── wire.go
│       └── wire_gen.go         # 自動生成
└── utils/
```

## 各レイヤーの責務

### 1. Entities（エンティティ層）

**責務**: オブジェクトでビジネスロジックを表現する

- ビジネスルールのカプセル化
- 一意な識別子を持たないValueObjectも含む
- 全レイヤーから参照されるinterface（Logger等）も定義

**例**:
```go
// backend/entities/user.go
type User struct {
    ID           uuid.UUID
    Username     string
    Balance      int64
    // ...
}

func (u *User) CanTransfer(amount int64) error {
    if u.Balance < amount {
        return errors.New("insufficient balance")
    }
    return nil
}
```

### 2. UseCases（ユースケース層）

**責務**: Entity・Repositoryを使い、ユースケースを達成する

#### InputPort
- ユースケースのinterfaceを定義
- Controllerから呼び出される

```go
// backend/usecases/inputport/point_transfer_inputport.go
type PointTransferInputPort interface {
    Transfer(req *TransferRequest) (*TransferResponse, error)
}
```

#### Interactor
- InputPortの実装
- トランザクションのスコープを管理
- Repositoryを使ってビジネスロジックを実行

```go
// backend/usecases/interactor/point_transfer_interactor.go
type PointTransferInteractor struct {
    db              *gorm.DB
    userRepo        repository.UserRepository
    transactionRepo repository.TransactionRepository
}

func (i *PointTransferInteractor) Transfer(req *inputport.TransferRequest) (*inputport.TransferResponse, error) {
    // トランザクション管理
    // ビジネスロジック実行
}
```

#### Repository Interface
- Interactorが要求するRepositoryのinterfaceを定義
- CRUD操作を強制（Create/Read/Update/Delete）

```go
// backend/usecases/repository/user_repository.go
type UserRepository interface {
    Create(user *entities.User) error
    Read(id uuid.UUID) (*entities.User, error)
    Update(user *entities.User) (bool, error)
    Delete(id uuid.UUID) error
}
```

### 3. Gateways（ゲートウェイ層）

**責務**: データの集約、永続化

#### Repository実装
- UseCasesレイヤーが実際のテーブル構造を把握しなくてもEntityの永続化を行える
- DataSourceを活用してデータの整合性を保証
- データの整合性が取れる最小単位（例: MySQL側のDataSourceを更新したら、Memcached側のDataSourceも更新）

```go
// backend/gateways/repository/user/user_repository_impl.go
type RepositoryImpl struct {
    userDS dsmysql.UserDataSource
    logger entities.Logger
}

func (r *RepositoryImpl) Create(user *entities.User) error {
    return r.userDS.Insert(user)
}
```

#### DataSource Interface
- Repositoryが期待するDataSourceのinterfaceを定義
- MySQLのtableや、Memcachedのkey、ElasticSearchのtypeと1:1の関係

```go
// backend/gateways/repository/datasource/dsmysql/user_datasource.go
type UserDataSource interface {
    Insert(user *entities.User) error
    Select(id uuid.UUID) (*entities.User, error)
    // SQLの操作名に沿った命名（Select/Insert/Update/Delete）
}
```

#### DataSource実装
- Infraを活用し、Repositoryが要求するデータの取得、永続化を達成
- 該当するミドルウェア固有の操作名に沿った命名規則

```go
// backend/gateways/datasource/dsmysqlimpl/user_datasource_impl.go
type UserDataSourceImpl struct {
    db inframysql.DB
}

func (ds *UserDataSourceImpl) Insert(user *entities.User) error {
    // GORM経由でDBに挿入
}
```

#### Infra層
- ミドルウェアとの実際の接続や入出力を担当
- 内側のレイヤーが各ミドルウェアのI/Fを把握せずとも利用できる状態にする

```go
// backend/gateways/infra/inframysql/postgres.go
type DB interface {
    GetDB() *gorm.DB
}

type PostgresDB struct {
    db *gorm.DB
}
```

### 4. Controllers（コントローラー層）

**責務**: 外界からの入力を、達成するユースケースが求めるインタフェースに変換する

- HTTP Request内のパラメータを取り出してInteractorに渡す
- Presenterを呼び出して外界が求める出力フォーマットに変更

```go
// backend/controllers/web/point_controller.go
type PointController struct {
    pointTransferUC inputport.PointTransferInputPort
    presenter       *presenter.PointPresenter
}

func (c *PointController) Transfer(ctx *gin.Context, currentTime time.Time) {
    // リクエストパラメータを取得
    // Interactorを呼び出し
    // Presenterで変換して出力
}
```

#### Presenter
- Interactorから返ってきたEntityを外界が求める出力フォーマットに変更

```go
// backend/controllers/web/presenter/point_presenter.go
func (p *PointPresenter) PresentTransferResponse(resp *inputport.TransferResponse) gin.H {
    return gin.H{
        "message": "transfer successful",
        "transaction": gin.H{
            "id": resp.Transaction.ID,
            // ...
        },
    }
}
```

### 5. Frameworks（フレームワーク層）

**責務**: 外界からの入力をControllerへルーティングする

#### Web Framework
- HTTP RequestのURLなどを参照し、該当するControllerへRequestを渡す
- Sessionの解決などもこのレイヤー
- 内部に引き回す時刻情報はリクエストを受け取った時刻

```go
// backend/frameworks/web/router.go
func (r *Router) RegisterRoutes(pointController *web.PointController) {
    api := r.engine.Group("/api")
    {
        points := api.Group("/points")
        {
            points.POST("/transfer", func(c *gin.Context) {
                // Controllerに時刻情報を渡す
                pointController.Transfer(c, r.timeProvider.Now())
            })
        }
    }
}
```

#### TimeProvider
- 時刻情報も外界の一部としてみなす
- このレイヤー以外では現在時刻を取得しないように制限
- テストを決定的にしたり、動作確認する際に任意の時刻への変更を容易にする

```go
// backend/frameworks/web/time_provider.go
type TimeProvider interface {
    Now() time.Time
}

type SystemTimeProvider struct{}

func (p *SystemTimeProvider) Now() time.Time {
    return time.Now()
}
```

### 6. DI（依存性注入）

**google/wire** を使用して依存性注入を自動化

```go
// backend/cmd/wire/wire.go
//go:build wireinject

func InitializeApp(dbConfig *inframysql.Config, routerConfig *frameworksweb.RouterConfig) (*frameworksweb.Router, error) {
    wire.Build(
        // Infra層
        inframysql.NewPostgresDB,
        infralogger.NewLogger,

        // DataSource層
        dsmysqlimpl.NewUserDataSource,

        // Repository層
        userrepo.NewUserRepository,

        // Interactor層
        interactor.NewPointTransferInteractor,

        // Presenter層
        presenter.NewPointPresenter,

        // Controller層
        web.NewPointController,

        // Framework層
        frameworksweb.NewSystemTimeProvider,
        frameworksweb.NewRouter,
    )

    return nil, nil
}
```

## 依存関係のルール

1. **内側のレイヤーは外側のレイヤーに依存しない**
   - Entities → UseCases → Gateways → Controllers → Frameworks

2. **interfaceは内側で定義、実装は外側で行う**
   - Repository interfaceはUseCasesで定義、実装はGatewaysで行う

3. **時刻情報の取得はFrameworksレイヤーのみ**
   - 他のレイヤーでは`time.Now()`を使用しない

## 命名規則

### Repository層
- **CRUD操作**: Create/Read/Update/Delete を強制

### DataSource層
- **SQL操作**: Select/Insert/Update/Delete
- **Cache操作**: Get/Set

## セキュリティと整合性

- **冪等性**: IdempotencyKeyで重複転送を防止
- **トランザクション**: DBトランザクションで原子性を保証
- **悲観的ロック**: SELECT FOR UPDATEで競合を防止
- **楽観的ロック**: Versionフィールドで並行制御

## ビルドとデプロイ

```bash
# wireコード生成
cd backend/cmd/wire
wire

# ビルド
cd backend
go build -o bin/server cmd/server/main.go

# 実行
./bin/server
```

## テストの方針

- **Entities**: ビジネスロジックの単体テスト
- **Interactor**: モックRepositoryを使用したテスト
- **Repository**: 統合テスト（実際のDB）
- **Controller**: HTTPリクエスト/レスポンスのテスト

## 今後の拡張

1. 他のエンティティ（Friendship, QRCode等）の完全実装
2. Memcachedを使ったキャッシュ層の実装
3. 非同期処理用のDaemon層の実装
4. OpenAPIの定義通りかどうかを検証する機能（開発環境のみ）
