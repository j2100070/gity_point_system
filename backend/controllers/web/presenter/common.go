package presenter

import (
	"time"

	"github.com/google/uuid"
)

// UserResponse はユーザーの共通レスポンス型
type UserResponse struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	AvatarURL   *string   `json:"avatar_url,omitempty"`
	Balance     int64     `json:"balance"`
	Role        string    `json:"role"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TransactionResponse は取引の共通レスポンス型
type TransactionResponse struct {
	ID              uuid.UUID     `json:"id"`
	FromUserID      *uuid.UUID    `json:"from_user_id"`
	ToUserID        *uuid.UUID    `json:"to_user_id"`
	Amount          int64         `json:"amount"`
	TransactionType string        `json:"transaction_type"`
	Status          string        `json:"status"`
	Description     string        `json:"description"`
	FromUser        *UserResponse `json:"from_user,omitempty"`
	ToUser          *UserResponse `json:"to_user,omitempty"`
	CreatedAt       time.Time     `json:"created_at"`
}
