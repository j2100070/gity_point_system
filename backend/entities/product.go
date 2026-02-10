package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Product は商品エンティティ
type Product struct {
	ID          uuid.UUID
	Name        string
	Description string
	CategoryCode string // カテゴリコード（categoriesテーブルのcodeを参照）
	Price       int64   // 交換に必要なポイント数
	Stock       int     // 在庫数（-1 = 無制限）
	ImageURL    string
	IsAvailable bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// NewProduct は新しい商品を作成
func NewProduct(name, description string, categoryCode string, price int64, stock int) (*Product, error) {
	if name == "" {
		return nil, errors.New("product name is required")
	}
	if price <= 0 {
		return nil, errors.New("price must be positive")
	}
	if stock < -1 {
		return nil, errors.New("stock must be -1 (unlimited) or positive")
	}

	return &Product{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		CategoryCode: categoryCode,
		Price:       price,
		Stock:       stock,
		IsAvailable: true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

// IsUnlimitedStock は在庫無制限かどうか
func (p *Product) IsUnlimitedStock() bool {
	return p.Stock == -1
}

// CanExchange は交換可能かどうか
func (p *Product) CanExchange(quantity int) error {
	if !p.IsAvailable {
		return errors.New("product is not available")
	}
	if p.DeletedAt != nil {
		return errors.New("product is deleted")
	}
	if quantity <= 0 {
		return errors.New("quantity must be positive")
	}
	if !p.IsUnlimitedStock() && p.Stock < quantity {
		return errors.New("insufficient stock")
	}
	return nil
}

// DeductStock は在庫を減らす
func (p *Product) DeductStock(quantity int) error {
	if err := p.CanExchange(quantity); err != nil {
		return err
	}
	if !p.IsUnlimitedStock() {
		p.Stock -= quantity
	}
	p.UpdatedAt = time.Now()
	return nil
}

// RestoreStock は在庫を戻す（キャンセル時）
func (p *Product) RestoreStock(quantity int) error {
	if quantity <= 0 {
		return errors.New("quantity must be positive")
	}
	if !p.IsUnlimitedStock() {
		p.Stock += quantity
	}
	p.UpdatedAt = time.Now()
	return nil
}

// ExchangeStatus は交換ステータス
type ExchangeStatus string

const (
	ExchangeStatusPending   ExchangeStatus = "pending"   // 申請中
	ExchangeStatusCompleted ExchangeStatus = "completed" // 完了（ポイント減算済み）
	ExchangeStatusCancelled ExchangeStatus = "cancelled" // キャンセル
	ExchangeStatusDelivered ExchangeStatus = "delivered" // 配達済み
)

// ProductExchange は商品交換エンティティ
type ProductExchange struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	ProductID     uuid.UUID
	Quantity      int
	PointsUsed    int64
	Status        ExchangeStatus
	TransactionID *uuid.UUID
	Notes         string
	CreatedAt     time.Time
	CompletedAt   *time.Time
	DeliveredAt   *time.Time
}

// NewProductExchange は新しい商品交換を作成
func NewProductExchange(userID, productID uuid.UUID, quantity int, pointsUsed int64, notes string) (*ProductExchange, error) {
	if quantity <= 0 {
		return nil, errors.New("quantity must be positive")
	}
	if pointsUsed <= 0 {
		return nil, errors.New("points used must be positive")
	}

	return &ProductExchange{
		ID:         uuid.New(),
		UserID:     userID,
		ProductID:  productID,
		Quantity:   quantity,
		PointsUsed: pointsUsed,
		Status:     ExchangeStatusPending,
		Notes:      notes,
		CreatedAt:  time.Now(),
	}, nil
}

// Complete は交換を完了状態にする
func (e *ProductExchange) Complete(transactionID uuid.UUID) error {
	if e.Status != ExchangeStatusPending {
		return errors.New("exchange is not pending")
	}
	now := time.Now()
	e.Status = ExchangeStatusCompleted
	e.TransactionID = &transactionID
	e.CompletedAt = &now
	return nil
}

// Cancel は交換をキャンセルする
func (e *ProductExchange) Cancel() error {
	if e.Status != ExchangeStatusPending {
		return errors.New("can only cancel pending exchange")
	}
	e.Status = ExchangeStatusCancelled
	return nil
}

// MarkAsDelivered は配達済みにする
func (e *ProductExchange) MarkAsDelivered() error {
	if e.Status != ExchangeStatusCompleted {
		return errors.New("exchange must be completed before delivery")
	}
	now := time.Now()
	e.Status = ExchangeStatusDelivered
	e.DeliveredAt = &now
	return nil
}
