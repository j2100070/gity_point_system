package dsmysqlimpl

import (
	"errors"
	"time"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/infra/inframysql"
	"github.com/gity/point-system/gateways/repository/datasource/dsmysql"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProductExchangeModel はGORM用の商品交換モデル
type ProductExchangeModel struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID        uuid.UUID  `gorm:"type:uuid;not null"`
	ProductID     uuid.UUID  `gorm:"type:uuid;not null"`
	Quantity      int        `gorm:"not null;check:quantity > 0"`
	PointsUsed    int64      `gorm:"not null;check:points_used > 0"`
	Status        string     `gorm:"type:varchar(50);not null;default:'pending'"`
	TransactionID *uuid.UUID `gorm:"type:uuid"`
	Notes         string     `gorm:"type:text"`
	CreatedAt     time.Time  `gorm:"not null;default:now()"`
	CompletedAt   *time.Time
	DeliveredAt   *time.Time
}

// TableName はテーブル名を指定
func (ProductExchangeModel) TableName() string {
	return "product_exchanges"
}

// ToDomain はドメインモデルに変換
func (e *ProductExchangeModel) ToDomain() *entities.ProductExchange {
	return &entities.ProductExchange{
		ID:            e.ID,
		UserID:        e.UserID,
		ProductID:     e.ProductID,
		Quantity:      e.Quantity,
		PointsUsed:    e.PointsUsed,
		Status:        entities.ExchangeStatus(e.Status),
		TransactionID: e.TransactionID,
		Notes:         e.Notes,
		CreatedAt:     e.CreatedAt,
		CompletedAt:   e.CompletedAt,
		DeliveredAt:   e.DeliveredAt,
	}
}

// FromDomain はドメインモデルから変換
func (e *ProductExchangeModel) FromDomain(exchange *entities.ProductExchange) {
	e.ID = exchange.ID
	e.UserID = exchange.UserID
	e.ProductID = exchange.ProductID
	e.Quantity = exchange.Quantity
	e.PointsUsed = exchange.PointsUsed
	e.Status = string(exchange.Status)
	e.TransactionID = exchange.TransactionID
	e.Notes = exchange.Notes
	e.CreatedAt = exchange.CreatedAt
	e.CompletedAt = exchange.CompletedAt
	e.DeliveredAt = exchange.DeliveredAt
}

// ProductExchangeDataSourceImpl はProductExchangeDataSourceの実装
type ProductExchangeDataSourceImpl struct {
	db inframysql.DB
}

// NewProductExchangeDataSource は新しいProductExchangeDataSourceを作成
func NewProductExchangeDataSource(db inframysql.DB) dsmysql.ProductExchangeDataSource {
	return &ProductExchangeDataSourceImpl{db: db}
}

// Insert は新しい交換を挿入
func (ds *ProductExchangeDataSourceImpl) Insert(tx interface{}, exchange *entities.ProductExchange) error {
	db := ds.db.GetDB()
	if tx != nil {
		db = tx.(*gorm.DB)
	}

	model := &ProductExchangeModel{}
	model.FromDomain(exchange)

	if err := db.Create(model).Error; err != nil {
		return err
	}

	*exchange = *model.ToDomain()
	return nil
}

// Select はIDで交換を検索
func (ds *ProductExchangeDataSourceImpl) Select(id uuid.UUID) (*entities.ProductExchange, error) {
	var model ProductExchangeModel

	err := ds.db.GetDB().Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("exchange not found")
		}
		return nil, err
	}

	return model.ToDomain(), nil
}

// Update は交換情報を更新
func (ds *ProductExchangeDataSourceImpl) Update(tx interface{}, exchange *entities.ProductExchange) error {
	db := ds.db.GetDB()
	if tx != nil {
		db = tx.(*gorm.DB)
	}

	model := &ProductExchangeModel{}
	model.FromDomain(exchange)

	return db.Where("id = ?", exchange.ID).Updates(model).Error
}

// SelectListByUserID はユーザーの交換履歴を取得
func (ds *ProductExchangeDataSourceImpl) SelectListByUserID(userID uuid.UUID, offset, limit int) ([]*entities.ProductExchange, error) {
	var models []ProductExchangeModel

	err := ds.db.GetDB().Where("user_id = ?", userID).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	exchanges := make([]*entities.ProductExchange, len(models))
	for i, model := range models {
		exchanges[i] = model.ToDomain()
	}

	return exchanges, nil
}

// SelectListAll はすべての交換履歴を取得
func (ds *ProductExchangeDataSourceImpl) SelectListAll(offset, limit int) ([]*entities.ProductExchange, error) {
	var models []ProductExchangeModel

	err := ds.db.GetDB().
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	exchanges := make([]*entities.ProductExchange, len(models))
	for i, model := range models {
		exchanges[i] = model.ToDomain()
	}

	return exchanges, nil
}

// CountByUserID はユーザーの交換総数を取得
func (ds *ProductExchangeDataSourceImpl) CountByUserID(userID uuid.UUID) (int64, error) {
	var count int64
	err := ds.db.GetDB().Model(&ProductExchangeModel{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

// CountAll は全体の交換総数を取得
func (ds *ProductExchangeDataSourceImpl) CountAll() (int64, error) {
	var count int64
	err := ds.db.GetDB().Model(&ProductExchangeModel{}).Count(&count).Error
	return count, err
}
