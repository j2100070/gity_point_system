package interactor_test

import (
	"context"
	"errors"
	"testing"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/gity/point-system/usecases/interactor"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========================================
// ProductExchangeInteractor テスト
// ========================================

// --- Mock ProductExchangeRepository ---

type mockExchangeRepo struct {
	exchanges map[uuid.UUID]*entities.ProductExchange
}

func newMockExchangeRepo() *mockExchangeRepo {
	return &mockExchangeRepo{exchanges: make(map[uuid.UUID]*entities.ProductExchange)}
}

func (m *mockExchangeRepo) Create(ctx context.Context, exchange *entities.ProductExchange) error {
	m.exchanges[exchange.ID] = exchange
	return nil
}
func (m *mockExchangeRepo) Read(ctx context.Context, id uuid.UUID) (*entities.ProductExchange, error) {
	e, ok := m.exchanges[id]
	if !ok {
		return nil, errors.New("exchange not found")
	}
	return e, nil
}
func (m *mockExchangeRepo) Update(ctx context.Context, exchange *entities.ProductExchange) error {
	m.exchanges[exchange.ID] = exchange
	return nil
}
func (m *mockExchangeRepo) ReadListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entities.ProductExchange, error) {
	result := make([]*entities.ProductExchange, 0)
	for _, e := range m.exchanges {
		if e.UserID == userID {
			result = append(result, e)
		}
	}
	return result, nil
}
func (m *mockExchangeRepo) ReadListAll(ctx context.Context, offset, limit int) ([]*entities.ProductExchange, error) {
	result := make([]*entities.ProductExchange, 0)
	for _, e := range m.exchanges {
		result = append(result, e)
	}
	return result, nil
}
func (m *mockExchangeRepo) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	count := int64(0)
	for _, e := range m.exchanges {
		if e.UserID == userID {
			count++
		}
	}
	return count, nil
}
func (m *mockExchangeRepo) CountAll(ctx context.Context) (int64, error) {
	return int64(len(m.exchanges)), nil
}

// --- ExchangeProduct ---

func TestProductExchangeInteractor_ExchangeProduct(t *testing.T) {
	setup := func() (*ctxTrackingTxManager, *ctxTrackingUserRepo, *mockProductRepo, *mockExchangeRepo, *ctxTrackingTransactionRepo, *ctxTrackingPointBatchRepo, *interactor.ProductExchangeInteractor) {
		txMgr := &ctxTrackingTxManager{}
		userRepo := newCtxTrackingUserRepo()
		prodRepo := newMockProductRepo()
		exchangeRepo := newMockExchangeRepo()
		txRepo := newCtxTrackingTransactionRepo()
		pbRepo := newCtxTrackingPointBatchRepo()
		logger := &mockLogger{}

		sut := interactor.NewProductExchangeInteractor(txMgr, prodRepo, exchangeRepo, userRepo, txRepo, pbRepo, logger)
		return txMgr, userRepo, prodRepo, exchangeRepo, txRepo, pbRepo, sut
	}

	t.Run("正常に商品を交換できる", func(t *testing.T) {
		_, userRepo, prodRepo, _, _, _, sut := setup()
		user := createTestUserWithBalance(t, "buyer", 10000, "user")
		userRepo.setUser(user)
		product, _ := entities.NewProduct("コーラ", "炭酸飲料", "drink", 100, 50)
		prodRepo.setProduct(product)

		resp, err := sut.ExchangeProduct(context.Background(), &inputport.ExchangeProductRequest{
			UserID: user.ID, ProductID: product.ID, Quantity: 2, Notes: "受取場所: 1F",
		})
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotNil(t, resp.Exchange)
		assert.NotNil(t, resp.Transaction)
		assert.Equal(t, int64(200), resp.Exchange.PointsUsed)
	})

	t.Run("数量が0以下の場合エラー", func(t *testing.T) {
		_, _, _, _, _, _, sut := setup()

		_, err := sut.ExchangeProduct(context.Background(), &inputport.ExchangeProductRequest{
			UserID: uuid.New(), ProductID: uuid.New(), Quantity: 0,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "quantity must be positive")
	})

	t.Run("商品が存在しない場合エラー", func(t *testing.T) {
		_, userRepo, _, _, _, _, sut := setup()
		user := createTestUserWithBalance(t, "buyer", 10000, "user")
		userRepo.setUser(user)

		_, err := sut.ExchangeProduct(context.Background(), &inputport.ExchangeProductRequest{
			UserID: user.ID, ProductID: uuid.New(), Quantity: 1,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "product not found")
	})

	t.Run("残高不足の場合エラー", func(t *testing.T) {
		_, userRepo, prodRepo, _, _, _, sut := setup()
		user := createTestUserWithBalance(t, "poor", 50, "user")
		userRepo.setUser(user)
		product, _ := entities.NewProduct("高級品", "高い", "luxury", 1000, 10)
		prodRepo.setProduct(product)

		_, err := sut.ExchangeProduct(context.Background(), &inputport.ExchangeProductRequest{
			UserID: user.ID, ProductID: product.ID, Quantity: 1,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient balance")
	})

	t.Run("在庫不足の場合エラー", func(t *testing.T) {
		_, userRepo, prodRepo, _, _, _, sut := setup()
		user := createTestUserWithBalance(t, "buyer", 100000, "user")
		userRepo.setUser(user)
		product, _ := entities.NewProduct("限定品", "少ない在庫", "limited", 100, 1)
		prodRepo.setProduct(product)

		_, err := sut.ExchangeProduct(context.Background(), &inputport.ExchangeProductRequest{
			UserID: user.ID, ProductID: product.ID, Quantity: 5,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient stock")
	})

	t.Run("非アクティブユーザーの場合エラー", func(t *testing.T) {
		_, userRepo, prodRepo, _, _, _, sut := setup()
		user := createTestUserWithBalance(t, "inactive", 10000, "user")
		user.IsActive = false
		userRepo.setUser(user)
		product, _ := entities.NewProduct("コーラ", "", "drink", 100, 50)
		prodRepo.setProduct(product)

		_, err := sut.ExchangeProduct(context.Background(), &inputport.ExchangeProductRequest{
			UserID: user.ID, ProductID: product.ID, Quantity: 1,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not active")
	})

	t.Run("txManager.Do内の呼び出しがトランザクションコンテキストを使用する", func(t *testing.T) {
		txMgr, userRepo, prodRepo, _, txRepo, pbRepo, sut := setup()
		user := createTestUserWithBalance(t, "buyer", 10000, "user")
		userRepo.setUser(user)
		product, _ := entities.NewProduct("コーラ", "", "drink", 100, 50)
		prodRepo.setProduct(product)

		_, err := sut.ExchangeProduct(context.Background(), &inputport.ExchangeProductRequest{
			UserID: user.ID, ProductID: product.ID, Quantity: 1,
		})
		require.NoError(t, err)
		require.NotNil(t, txMgr.TxCtx)

		assert.True(t, isTxContext(userRepo.ctxRecords["UpdateBalancesWithLock"]),
			"userRepo.UpdateBalancesWithLock はトランザクションコンテキストを使用すべき")
		assert.True(t, isTxContext(txRepo.ctxRecords["Create"]),
			"transactionRepo.Create はトランザクションコンテキストを使用すべき")
		assert.True(t, isTxContext(pbRepo.ctxRecords["ConsumePointsFIFO"]),
			"pointBatchRepo.ConsumePointsFIFO はトランザクションコンテキストを使用すべき")
	})
}

// --- GetExchangeHistory ---

func TestProductExchangeInteractor_GetExchangeHistory(t *testing.T) {
	t.Run("正常に交換履歴を取得できる", func(t *testing.T) {
		exchangeRepo := newMockExchangeRepo()
		sut := interactor.NewProductExchangeInteractor(
			&ctxTrackingTxManager{}, newMockProductRepo(), exchangeRepo,
			newCtxTrackingUserRepo(), newCtxTrackingTransactionRepo(),
			newCtxTrackingPointBatchRepo(), &mockLogger{},
		)

		userID := uuid.New()
		exchange, _ := entities.NewProductExchange(userID, uuid.New(), 1, 100, "test")
		exchangeRepo.exchanges[exchange.ID] = exchange

		resp, err := sut.GetExchangeHistory(context.Background(), &inputport.GetExchangeHistoryRequest{
			UserID: userID, Offset: 0, Limit: 20,
		})
		require.NoError(t, err)
		assert.Equal(t, 1, len(resp.Exchanges))
		assert.Equal(t, int64(1), resp.Total)
	})
}

// --- CancelExchange ---

func TestProductExchangeInteractor_CancelExchange(t *testing.T) {
	setup := func() (*mockExchangeRepo, *mockProductRepo, *ctxTrackingUserRepo, *interactor.ProductExchangeInteractor) {
		exchangeRepo := newMockExchangeRepo()
		prodRepo := newMockProductRepo()
		userRepo := newCtxTrackingUserRepo()
		sut := interactor.NewProductExchangeInteractor(
			&ctxTrackingTxManager{}, prodRepo, exchangeRepo,
			userRepo, newCtxTrackingTransactionRepo(),
			newCtxTrackingPointBatchRepo(), &mockLogger{},
		)
		return exchangeRepo, prodRepo, userRepo, sut
	}

	t.Run("正常に交換をキャンセルできる", func(t *testing.T) {
		exchangeRepo, prodRepo, _, sut := setup()
		userID := uuid.New()
		productID := uuid.New()
		product, _ := entities.NewProduct("コーラ", "", "drink", 100, 48)
		product.ID = productID
		prodRepo.setProduct(product)

		exchange, _ := entities.NewProductExchange(userID, productID, 2, 200, "")
		exchangeRepo.exchanges[exchange.ID] = exchange

		err := sut.CancelExchange(context.Background(), &inputport.CancelExchangeRequest{
			UserID: userID, ExchangeID: exchange.ID,
		})
		assert.NoError(t, err)
	})

	t.Run("他人の交換はキャンセルできない", func(t *testing.T) {
		exchangeRepo, prodRepo, _, sut := setup()
		ownerID := uuid.New()
		productID := uuid.New()
		product, _ := entities.NewProduct("コーラ", "", "drink", 100, 48)
		product.ID = productID
		prodRepo.setProduct(product)

		exchange, _ := entities.NewProductExchange(ownerID, productID, 1, 100, "")
		exchangeRepo.exchanges[exchange.ID] = exchange

		err := sut.CancelExchange(context.Background(), &inputport.CancelExchangeRequest{
			UserID: uuid.New(), ExchangeID: exchange.ID,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
	})

	t.Run("存在しない交換の場合エラー", func(t *testing.T) {
		_, _, _, sut := setup()

		err := sut.CancelExchange(context.Background(), &inputport.CancelExchangeRequest{
			UserID: uuid.New(), ExchangeID: uuid.New(),
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exchange not found")
	})
}

// --- MarkExchangeDelivered ---

func TestProductExchangeInteractor_MarkExchangeDelivered(t *testing.T) {
	t.Run("正常に配達完了にできる", func(t *testing.T) {
		exchangeRepo := newMockExchangeRepo()
		sut := interactor.NewProductExchangeInteractor(
			&ctxTrackingTxManager{}, newMockProductRepo(), exchangeRepo,
			newCtxTrackingUserRepo(), newCtxTrackingTransactionRepo(),
			newCtxTrackingPointBatchRepo(), &mockLogger{},
		)

		exchange, _ := entities.NewProductExchange(uuid.New(), uuid.New(), 1, 100, "")
		txID := uuid.New()
		_ = exchange.Complete(txID) // Completed 状態にする
		exchangeRepo.exchanges[exchange.ID] = exchange

		err := sut.MarkExchangeDelivered(context.Background(), &inputport.MarkExchangeDeliveredRequest{
			ExchangeID: exchange.ID,
		})
		assert.NoError(t, err)
	})

	t.Run("Pending状態の交換は配達完了にできない", func(t *testing.T) {
		exchangeRepo := newMockExchangeRepo()
		sut := interactor.NewProductExchangeInteractor(
			&ctxTrackingTxManager{}, newMockProductRepo(), exchangeRepo,
			newCtxTrackingUserRepo(), newCtxTrackingTransactionRepo(),
			newCtxTrackingPointBatchRepo(), &mockLogger{},
		)

		exchange, _ := entities.NewProductExchange(uuid.New(), uuid.New(), 1, 100, "")
		exchangeRepo.exchanges[exchange.ID] = exchange

		err := sut.MarkExchangeDelivered(context.Background(), &inputport.MarkExchangeDeliveredRequest{
			ExchangeID: exchange.ID,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be completed before delivery")
	})
}

// --- GetAllExchanges ---

func TestProductExchangeInteractor_GetAllExchanges(t *testing.T) {
	t.Run("正常にすべての交換履歴を取得できる", func(t *testing.T) {
		exchangeRepo := newMockExchangeRepo()
		sut := interactor.NewProductExchangeInteractor(
			&ctxTrackingTxManager{}, newMockProductRepo(), exchangeRepo,
			newCtxTrackingUserRepo(), newCtxTrackingTransactionRepo(),
			newCtxTrackingPointBatchRepo(), &mockLogger{},
		)

		e1, _ := entities.NewProductExchange(uuid.New(), uuid.New(), 1, 100, "")
		e2, _ := entities.NewProductExchange(uuid.New(), uuid.New(), 2, 200, "")
		exchangeRepo.exchanges[e1.ID] = e1
		exchangeRepo.exchanges[e2.ID] = e2

		resp, err := sut.GetAllExchanges(context.Background(), 0, 20)
		require.NoError(t, err)
		assert.Equal(t, 2, len(resp.Exchanges))
		assert.Equal(t, int64(2), resp.Total)
	})
}
