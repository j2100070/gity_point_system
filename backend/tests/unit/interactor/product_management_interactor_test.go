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
// ProductManagementInteractor テスト
// ========================================

// --- Mock ProductRepository ---

type mockProductRepo struct {
	products map[uuid.UUID]*entities.Product
}

func newMockProductRepo() *mockProductRepo {
	return &mockProductRepo{products: make(map[uuid.UUID]*entities.Product)}
}

func (m *mockProductRepo) setProduct(p *entities.Product) { m.products[p.ID] = p }

func (m *mockProductRepo) Create(ctx context.Context, product *entities.Product) error {
	m.products[product.ID] = product
	return nil
}
func (m *mockProductRepo) Read(ctx context.Context, id uuid.UUID) (*entities.Product, error) {
	p, ok := m.products[id]
	if !ok {
		return nil, errors.New("product not found")
	}
	copy := *p
	return &copy, nil
}
func (m *mockProductRepo) Update(ctx context.Context, product *entities.Product) error {
	m.products[product.ID] = product
	return nil
}
func (m *mockProductRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.products, id)
	return nil
}
func (m *mockProductRepo) ReadList(ctx context.Context, offset, limit int) ([]*entities.Product, error) {
	result := make([]*entities.Product, 0)
	for _, p := range m.products {
		result = append(result, p)
	}
	return result, nil
}
func (m *mockProductRepo) ReadListByCategory(ctx context.Context, categoryCode string, offset, limit int) ([]*entities.Product, error) {
	result := make([]*entities.Product, 0)
	for _, p := range m.products {
		if p.CategoryCode == categoryCode {
			result = append(result, p)
		}
	}
	return result, nil
}
func (m *mockProductRepo) ReadAvailableList(ctx context.Context, offset, limit int) ([]*entities.Product, error) {
	result := make([]*entities.Product, 0)
	for _, p := range m.products {
		if p.IsAvailable {
			result = append(result, p)
		}
	}
	return result, nil
}
func (m *mockProductRepo) Count(ctx context.Context) (int64, error) {
	return int64(len(m.products)), nil
}
func (m *mockProductRepo) UpdateStock(ctx context.Context, productID uuid.UUID, quantity int) error {
	return nil
}

// --- CreateProduct ---

func TestProductManagementInteractor_CreateProduct(t *testing.T) {
	setup := func() (*mockProductRepo, inputport.ProductManagementInputPort) {
		prodRepo := newMockProductRepo()
		sut := interactor.NewProductManagementInteractor(prodRepo, &mockLogger{})
		return prodRepo, sut
	}

	t.Run("正常に商品を作成できる", func(t *testing.T) {
		_, sut := setup()

		resp, err := sut.CreateProduct(context.Background(), &inputport.CreateProductRequest{
			Name: "コーラ", Description: "炭酸飲料", Category: "drink", Price: 100, Stock: 50,
		})
		require.NoError(t, err)
		assert.NotNil(t, resp.Product)
		assert.Equal(t, "コーラ", resp.Product.Name)
		assert.Equal(t, int64(100), resp.Product.Price)
	})

	t.Run("名前が空の場合エラー", func(t *testing.T) {
		_, sut := setup()

		_, err := sut.CreateProduct(context.Background(), &inputport.CreateProductRequest{
			Name: "", Description: "", Category: "drink", Price: 100, Stock: 10,
		})
		assert.Error(t, err)
	})

	t.Run("価格が0以下の場合エラー", func(t *testing.T) {
		_, sut := setup()

		_, err := sut.CreateProduct(context.Background(), &inputport.CreateProductRequest{
			Name: "テスト", Description: "", Category: "drink", Price: 0, Stock: 10,
		})
		assert.Error(t, err)
	})
}

// --- UpdateProduct ---

func TestProductManagementInteractor_UpdateProduct(t *testing.T) {
	setup := func() (*mockProductRepo, inputport.ProductManagementInputPort) {
		prodRepo := newMockProductRepo()
		sut := interactor.NewProductManagementInteractor(prodRepo, &mockLogger{})
		return prodRepo, sut
	}

	t.Run("正常に商品を更新できる", func(t *testing.T) {
		prodRepo, sut := setup()
		product, _ := entities.NewProduct("旧商品", "旧説明", "drink", 100, 10)
		prodRepo.setProduct(product)

		resp, err := sut.UpdateProduct(context.Background(), &inputport.UpdateProductRequest{
			ProductID: product.ID, Name: "新商品", Description: "新説明",
			Category: "snack", Price: 200, Stock: 20, IsAvailable: true,
		})
		require.NoError(t, err)
		assert.Equal(t, "新商品", resp.Product.Name)
		assert.Equal(t, int64(200), resp.Product.Price)
	})

	t.Run("存在しない商品の場合エラー", func(t *testing.T) {
		_, sut := setup()

		_, err := sut.UpdateProduct(context.Background(), &inputport.UpdateProductRequest{
			ProductID: uuid.New(), Name: "test", Price: 100,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "product not found")
	})
}

// --- DeleteProduct ---

func TestProductManagementInteractor_DeleteProduct(t *testing.T) {
	t.Run("正常に商品を削除できる", func(t *testing.T) {
		prodRepo := newMockProductRepo()
		sut := interactor.NewProductManagementInteractor(prodRepo, &mockLogger{})
		product, _ := entities.NewProduct("削除対象", "説明", "drink", 100, 10)
		prodRepo.setProduct(product)

		err := sut.DeleteProduct(context.Background(), &inputport.DeleteProductRequest{
			ProductID: product.ID,
		})
		assert.NoError(t, err)
	})
}

// --- GetProductList ---

func TestProductManagementInteractor_GetProductList(t *testing.T) {
	setup := func() (*mockProductRepo, inputport.ProductManagementInputPort) {
		prodRepo := newMockProductRepo()
		sut := interactor.NewProductManagementInteractor(prodRepo, &mockLogger{})
		return prodRepo, sut
	}

	t.Run("全商品一覧を取得できる", func(t *testing.T) {
		prodRepo, sut := setup()
		p1, _ := entities.NewProduct("コーラ", "", "drink", 100, 10)
		p2, _ := entities.NewProduct("チョコ", "", "snack", 50, 20)
		prodRepo.setProduct(p1)
		prodRepo.setProduct(p2)

		resp, err := sut.GetProductList(context.Background(), &inputport.GetProductListRequest{
			Offset: 0, Limit: 20,
		})
		require.NoError(t, err)
		assert.Equal(t, 2, len(resp.Products))
		assert.Equal(t, int64(2), resp.Total)
	})

	t.Run("カテゴリでフィルタリングできる", func(t *testing.T) {
		prodRepo, sut := setup()
		p1, _ := entities.NewProduct("コーラ", "", "drink", 100, 10)
		p2, _ := entities.NewProduct("チョコ", "", "snack", 50, 20)
		prodRepo.setProduct(p1)
		prodRepo.setProduct(p2)

		resp, err := sut.GetProductList(context.Background(), &inputport.GetProductListRequest{
			Category: "drink", Offset: 0, Limit: 20,
		})
		require.NoError(t, err)
		assert.Equal(t, 1, len(resp.Products))
		assert.Equal(t, "コーラ", resp.Products[0].Name)
	})

	t.Run("交換可能な商品のみ取得できる", func(t *testing.T) {
		prodRepo, sut := setup()
		p1, _ := entities.NewProduct("利用可能", "", "drink", 100, 10)
		p2, _ := entities.NewProduct("非利用", "", "snack", 50, 20)
		p2.IsAvailable = false
		prodRepo.setProduct(p1)
		prodRepo.setProduct(p2)

		resp, err := sut.GetProductList(context.Background(), &inputport.GetProductListRequest{
			AvailableOnly: true, Offset: 0, Limit: 20,
		})
		require.NoError(t, err)
		assert.Equal(t, 1, len(resp.Products))
	})
}
