package dsmysqlimpl

import (
	"context"
	"errors"
	"time"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/gateways/infra/inframysql"
	"github.com/gity/point-system/gateways/repository/datasource/dsmysql"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TransferRequestModel はGORM用の送金リクエストモデル
type TransferRequestModel struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	FromUserID     uuid.UUID  `gorm:"type:uuid;not null;index"`
	ToUserID       uuid.UUID  `gorm:"type:uuid;not null;index"`
	Amount         int64      `gorm:"not null"`
	Message        string     `gorm:"type:text"`
	Status         string     `gorm:"type:varchar(50);not null;index;default:'pending'"`
	IdempotencyKey string     `gorm:"type:varchar(255);not null;unique;index"`
	ExpiresAt      time.Time  `gorm:"not null;index"`
	ApprovedAt     *time.Time `gorm:"type:timestamp with time zone"`
	RejectedAt     *time.Time `gorm:"type:timestamp with time zone"`
	CancelledAt    *time.Time `gorm:"type:timestamp with time zone"`
	TransactionID  *uuid.UUID `gorm:"type:uuid"`
	CreatedAt      time.Time  `gorm:"not null;default:now()"`
	UpdatedAt      time.Time  `gorm:"not null;default:now()"`
}

// TableName はテーブル名を指定
func (TransferRequestModel) TableName() string {
	return "transfer_requests"
}

// ToDomain はドメインモデルに変換
func (tr *TransferRequestModel) ToDomain() *entities.TransferRequest {
	return &entities.TransferRequest{
		ID:             tr.ID,
		FromUserID:     tr.FromUserID,
		ToUserID:       tr.ToUserID,
		Amount:         tr.Amount,
		Message:        tr.Message,
		Status:         entities.TransferRequestStatus(tr.Status),
		IdempotencyKey: tr.IdempotencyKey,
		ExpiresAt:      tr.ExpiresAt,
		ApprovedAt:     tr.ApprovedAt,
		RejectedAt:     tr.RejectedAt,
		CancelledAt:    tr.CancelledAt,
		TransactionID:  tr.TransactionID,
		CreatedAt:      tr.CreatedAt,
		UpdatedAt:      tr.UpdatedAt,
	}
}

// FromDomain はドメインモデルから変換
func (tr *TransferRequestModel) FromDomain(transferRequest *entities.TransferRequest) {
	tr.ID = transferRequest.ID
	tr.FromUserID = transferRequest.FromUserID
	tr.ToUserID = transferRequest.ToUserID
	tr.Amount = transferRequest.Amount
	tr.Message = transferRequest.Message
	tr.Status = string(transferRequest.Status)
	tr.IdempotencyKey = transferRequest.IdempotencyKey
	tr.ExpiresAt = transferRequest.ExpiresAt
	tr.ApprovedAt = transferRequest.ApprovedAt
	tr.RejectedAt = transferRequest.RejectedAt
	tr.CancelledAt = transferRequest.CancelledAt
	tr.TransactionID = transferRequest.TransactionID
	tr.CreatedAt = transferRequest.CreatedAt
	tr.UpdatedAt = transferRequest.UpdatedAt
}

// TransferRequestDataSourceImpl はTransferRequestDataSourceの実装
type TransferRequestDataSourceImpl struct {
	db inframysql.DB
}

// NewTransferRequestDataSource は新しいTransferRequestDataSourceを作成
func NewTransferRequestDataSource(db inframysql.DB) dsmysql.TransferRequestDataSource {
	return &TransferRequestDataSourceImpl{db: db}
}

// Insert は新しい送金リクエストを挿入
func (ds *TransferRequestDataSourceImpl) Insert(ctx context.Context, transferRequest *entities.TransferRequest) error {
	model := &TransferRequestModel{}
	model.FromDomain(transferRequest)

	if err := inframysql.GetDB(ctx, ds.db.GetDB()).Create(model).Error; err != nil {
		return err
	}

	*transferRequest = *model.ToDomain()
	return nil
}

// Select はIDで送金リクエストを検索
func (ds *TransferRequestDataSourceImpl) Select(ctx context.Context, id uuid.UUID) (*entities.TransferRequest, error) {
	var model TransferRequestModel

	err := inframysql.GetDB(ctx, ds.db.GetDB()).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return model.ToDomain(), nil
}

// SelectByIdempotencyKey は冪等性キーで送金リクエストを検索
func (ds *TransferRequestDataSourceImpl) SelectByIdempotencyKey(ctx context.Context, key string) (*entities.TransferRequest, error) {
	var model TransferRequestModel

	err := inframysql.GetDB(ctx, ds.db.GetDB()).Where("idempotency_key = ?", key).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return model.ToDomain(), nil
}

// Update は送金リクエストを更新
func (ds *TransferRequestDataSourceImpl) Update(ctx context.Context, transferRequest *entities.TransferRequest) error {
	model := &TransferRequestModel{}
	model.FromDomain(transferRequest)

	result := inframysql.GetDB(ctx, ds.db.GetDB()).Save(model)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("no rows affected")
	}

	*transferRequest = *model.ToDomain()
	return nil
}

// SelectPendingByToUser は受取人宛の承認待ちリクエストを取得
func (ds *TransferRequestDataSourceImpl) SelectPendingByToUser(ctx context.Context, toUserID uuid.UUID, offset, limit int) ([]*entities.TransferRequest, error) {
	var models []TransferRequestModel

	err := inframysql.GetDB(ctx, ds.db.GetDB()).
		Where("to_user_id = ? AND status = ?", toUserID, string(entities.TransferRequestStatusPending)).
		Where("expires_at > ?", time.Now()). // 有効期限内のみ
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	requests := make([]*entities.TransferRequest, 0, len(models))
	for _, model := range models {
		requests = append(requests, model.ToDomain())
	}

	return requests, nil
}

// SelectSentByFromUser は送信者が送ったリクエストを取得
func (ds *TransferRequestDataSourceImpl) SelectSentByFromUser(ctx context.Context, fromUserID uuid.UUID, offset, limit int) ([]*entities.TransferRequest, error) {
	var models []TransferRequestModel

	err := inframysql.GetDB(ctx, ds.db.GetDB()).
		Where("from_user_id = ?", fromUserID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	requests := make([]*entities.TransferRequest, 0, len(models))
	for _, model := range models {
		requests = append(requests, model.ToDomain())
	}

	return requests, nil
}

// CountPendingByToUser は受取人宛の承認待ちリクエスト数を取得
func (ds *TransferRequestDataSourceImpl) CountPendingByToUser(ctx context.Context, toUserID uuid.UUID) (int64, error) {
	var count int64

	err := inframysql.GetDB(ctx, ds.db.GetDB()).
		Model(&TransferRequestModel{}).
		Where("to_user_id = ? AND status = ?", toUserID, string(entities.TransferRequestStatusPending)).
		Where("expires_at > ?", time.Now()). // 有効期限内のみ
		Count(&count).Error

	if err != nil {
		return 0, err
	}

	return count, nil
}

// UpdateExpiredRequests は期限切れのリクエストを一括更新
func (ds *TransferRequestDataSourceImpl) UpdateExpiredRequests(ctx context.Context) (int64, error) {
	result := inframysql.GetDB(ctx, ds.db.GetDB()).
		Model(&TransferRequestModel{}).
		Where("status = ? AND expires_at <= ?", string(entities.TransferRequestStatusPending), time.Now()).
		Update("status", string(entities.TransferRequestStatusExpired))

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

// transferRequestWithUsersRow はJOINクエリの結果を受け取る構造体
type transferRequestWithUsersRow struct {
	// TransferRequest fields
	ID             uuid.UUID  `gorm:"column:id"`
	FromUserID     uuid.UUID  `gorm:"column:from_user_id"`
	ToUserID       uuid.UUID  `gorm:"column:to_user_id"`
	Amount         int64      `gorm:"column:amount"`
	Message        string     `gorm:"column:message"`
	Status         string     `gorm:"column:status"`
	IdempotencyKey string     `gorm:"column:idempotency_key"`
	ExpiresAt      time.Time  `gorm:"column:expires_at"`
	ApprovedAt     *time.Time `gorm:"column:approved_at"`
	RejectedAt     *time.Time `gorm:"column:rejected_at"`
	CancelledAt    *time.Time `gorm:"column:cancelled_at"`
	TransactionID  *uuid.UUID `gorm:"column:transaction_id"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at"`
	// FromUser fields
	FromID          string  `gorm:"column:from_id"`
	FromUsername    string  `gorm:"column:from_username"`
	FromDisplayName string  `gorm:"column:from_display_name"`
	FromFirstName   string  `gorm:"column:from_first_name"`
	FromLastName    string  `gorm:"column:from_last_name"`
	FromAvatarURL   *string `gorm:"column:from_avatar_url"`
	FromAvatarType  string  `gorm:"column:from_avatar_type"`
	// ToUser fields
	ToID          string  `gorm:"column:to_id"`
	ToUsername    string  `gorm:"column:to_username"`
	ToDisplayName string  `gorm:"column:to_display_name"`
	ToFirstName   string  `gorm:"column:to_first_name"`
	ToLastName    string  `gorm:"column:to_last_name"`
	ToAvatarURL   *string `gorm:"column:to_avatar_url"`
	ToAvatarType  string  `gorm:"column:to_avatar_type"`
}

func (r *transferRequestWithUsersRow) toDomain() *entities.TransferRequestWithUsers {
	fromID, _ := uuid.Parse(r.FromID)
	toID, _ := uuid.Parse(r.ToID)
	return &entities.TransferRequestWithUsers{
		TransferRequest: &entities.TransferRequest{
			ID:             r.ID,
			FromUserID:     r.FromUserID,
			ToUserID:       r.ToUserID,
			Amount:         r.Amount,
			Message:        r.Message,
			Status:         entities.TransferRequestStatus(r.Status),
			IdempotencyKey: r.IdempotencyKey,
			ExpiresAt:      r.ExpiresAt,
			ApprovedAt:     r.ApprovedAt,
			RejectedAt:     r.RejectedAt,
			CancelledAt:    r.CancelledAt,
			TransactionID:  r.TransactionID,
			CreatedAt:      r.CreatedAt,
			UpdatedAt:      r.UpdatedAt,
		},
		FromUser: &entities.User{
			ID:          fromID,
			Username:    r.FromUsername,
			DisplayName: r.FromDisplayName,
			FirstName:   r.FromFirstName,
			LastName:    r.FromLastName,
			AvatarURL:   r.FromAvatarURL,
			AvatarType:  entities.AvatarType(r.FromAvatarType),
		},
		ToUser: &entities.User{
			ID:          toID,
			Username:    r.ToUsername,
			DisplayName: r.ToDisplayName,
			FirstName:   r.ToFirstName,
			LastName:    r.ToLastName,
			AvatarURL:   r.ToAvatarURL,
			AvatarType:  entities.AvatarType(r.ToAvatarType),
		},
	}
}

const transferRequestWithUsersSQL = `SELECT tr.id, tr.from_user_id, tr.to_user_id, tr.amount, tr.message,
	tr.status, tr.idempotency_key, tr.expires_at, tr.approved_at, tr.rejected_at,
	tr.cancelled_at, tr.transaction_id, tr.created_at, tr.updated_at,
	from_u.id AS from_id, from_u.username AS from_username,
	from_u.display_name AS from_display_name, from_u.first_name AS from_first_name,
	from_u.last_name AS from_last_name, from_u.avatar_url AS from_avatar_url,
	from_u.avatar_type AS from_avatar_type,
	to_u.id AS to_id, to_u.username AS to_username,
	to_u.display_name AS to_display_name, to_u.first_name AS to_first_name,
	to_u.last_name AS to_last_name, to_u.avatar_url AS to_avatar_url,
	to_u.avatar_type AS to_avatar_type
FROM transfer_requests tr
LEFT JOIN users from_u ON from_u.id = tr.from_user_id
LEFT JOIN users to_u ON to_u.id = tr.to_user_id`

// SelectPendingByToUserWithUsers は受取人宛の承認待ちリクエストをユーザー情報付きで取得（JOIN）
func (ds *TransferRequestDataSourceImpl) SelectPendingByToUserWithUsers(ctx context.Context, toUserID uuid.UUID, offset, limit int) ([]*entities.TransferRequestWithUsers, error) {
	var rows []transferRequestWithUsersRow

	err := inframysql.GetDB(ctx, ds.db.GetDB()).
		Raw(transferRequestWithUsersSQL+`
		WHERE tr.to_user_id = ? AND tr.status = ? AND tr.expires_at > ?
		ORDER BY tr.created_at DESC
		LIMIT ? OFFSET ?`,
			toUserID, string(entities.TransferRequestStatusPending), time.Now(), limit, offset).
		Scan(&rows).Error

	if err != nil {
		return nil, err
	}

	results := make([]*entities.TransferRequestWithUsers, len(rows))
	for i, row := range rows {
		results[i] = row.toDomain()
	}
	return results, nil
}

// SelectSentByFromUserWithUsers は送信者が送ったリクエストをユーザー情報付きで取得（JOIN）
func (ds *TransferRequestDataSourceImpl) SelectSentByFromUserWithUsers(ctx context.Context, fromUserID uuid.UUID, offset, limit int) ([]*entities.TransferRequestWithUsers, error) {
	var rows []transferRequestWithUsersRow

	err := inframysql.GetDB(ctx, ds.db.GetDB()).
		Raw(transferRequestWithUsersSQL+`
		WHERE tr.from_user_id = ?
		ORDER BY tr.created_at DESC
		LIMIT ? OFFSET ?`,
			fromUserID, limit, offset).
		Scan(&rows).Error

	if err != nil {
		return nil, err
	}

	results := make([]*entities.TransferRequestWithUsers, len(rows))
	for i, row := range rows {
		results[i] = row.toDomain()
	}
	return results, nil
}
