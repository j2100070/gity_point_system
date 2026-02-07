package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Category は商品カテゴリエンティティ
type Category struct {
	ID          uuid.UUID
	Name        string // カテゴリ名（例: "飲み物", "お菓子"）
	Code        string // カテゴリコード（例: "drink", "snack"）
	Description string // 説明
	DisplayOrder int   // 表示順序
	IsActive    bool   // 有効/無効
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// NewCategory は新しいカテゴリを作成
func NewCategory(name, code, description string, displayOrder int) (*Category, error) {
	if name == "" {
		return nil, errors.New("category name is required")
	}
	if code == "" {
		return nil, errors.New("category code is required")
	}

	now := time.Now()
	return &Category{
		ID:           uuid.New(),
		Name:         name,
		Code:         code,
		Description:  description,
		DisplayOrder: displayOrder,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// Update はカテゴリ情報を更新
func (c *Category) Update(name, description string, displayOrder int, isActive bool) {
	c.Name = name
	c.Description = description
	c.DisplayOrder = displayOrder
	c.IsActive = isActive
	c.UpdatedAt = time.Now()
}

// Delete はカテゴリを論理削除
func (c *Category) Delete() {
	now := time.Now()
	c.DeletedAt = &now
}
