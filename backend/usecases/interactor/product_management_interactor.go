package interactor

import (
	"fmt"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/gity/point-system/usecases/repository"
)

// ProductManagementInteractor は商品管理のユースケース実装（管理者用）
type ProductManagementInteractor struct {
	productRepo repository.ProductRepository
	logger      entities.Logger
}

// NewProductManagementInteractor は新しいProductManagementInteractorを作成
func NewProductManagementInteractor(
	productRepo repository.ProductRepository,
	logger entities.Logger,
) inputport.ProductManagementInputPort {
	return &ProductManagementInteractor{
		productRepo: productRepo,
		logger:      logger,
	}
}

// CreateProduct は新しい商品を作成（管理者のみ）
func (i *ProductManagementInteractor) CreateProduct(req *inputport.CreateProductRequest) (*inputport.CreateProductResponse, error) {
	i.logger.Info("Creating new product", entities.NewField("name", req.Name))

	product, err := entities.NewProduct(
		req.Name,
		req.Description,
		req.Category,
		req.Price,
		req.Stock,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	product.ImageURL = req.ImageURL

	if err := i.productRepo.Create(product); err != nil {
		i.logger.Error("Failed to create product", entities.NewField("error", err))
		return nil, fmt.Errorf("failed to save product: %w", err)
	}

	i.logger.Info("Product created successfully", entities.NewField("product_id", product.ID))

	return &inputport.CreateProductResponse{
		Product: product,
	}, nil
}

// UpdateProduct は商品情報を更新（管理者のみ）
func (i *ProductManagementInteractor) UpdateProduct(req *inputport.UpdateProductRequest) (*inputport.UpdateProductResponse, error) {
	i.logger.Info("Updating product", entities.NewField("product_id", req.ProductID))

	product, err := i.productRepo.Read(req.ProductID)
	if err != nil {
		return nil, fmt.Errorf("product not found: %w", err)
	}

	// 商品情報を更新
	product.Name = req.Name
	product.Description = req.Description
	product.CategoryCode = req.Category
	product.Price = req.Price
	product.Stock = req.Stock
	product.ImageURL = req.ImageURL
	product.IsAvailable = req.IsAvailable

	if err := i.productRepo.Update(product); err != nil {
		i.logger.Error("Failed to update product", entities.NewField("error", err))
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	i.logger.Info("Product updated successfully", entities.NewField("product_id", product.ID))

	return &inputport.UpdateProductResponse{
		Product: product,
	}, nil
}

// DeleteProduct は商品を削除（管理者のみ）
func (i *ProductManagementInteractor) DeleteProduct(req *inputport.DeleteProductRequest) error {
	i.logger.Info("Deleting product", entities.NewField("product_id", req.ProductID))

	if err := i.productRepo.Delete(req.ProductID); err != nil {
		i.logger.Error("Failed to delete product", entities.NewField("error", err))
		return fmt.Errorf("failed to delete product: %w", err)
	}

	i.logger.Info("Product deleted successfully", entities.NewField("product_id", req.ProductID))
	return nil
}

// GetProductList は商品一覧を取得
func (i *ProductManagementInteractor) GetProductList(req *inputport.GetProductListRequest) (*inputport.GetProductListResponse, error) {
	var products []*entities.Product
	var err error

	// カテゴリでフィルタリング
	if req.Category != "" {
		products, err = i.productRepo.ReadListByCategory(req.Category, req.Offset, req.Limit)
	} else if req.AvailableOnly {
		// 交換可能な商品のみ
		products, err = i.productRepo.ReadAvailableList(req.Offset, req.Limit)
	} else {
		// すべての商品
		products, err = i.productRepo.ReadList(req.Offset, req.Limit)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get product list: %w", err)
	}

	total, err := i.productRepo.Count()
	if err != nil {
		return nil, fmt.Errorf("failed to count products: %w", err)
	}

	return &inputport.GetProductListResponse{
		Products: products,
		Total:    total,
	}, nil
}
