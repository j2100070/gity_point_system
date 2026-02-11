package presenter

import (
	"github.com/gity/point-system/usecases/inputport"
)

// AdminPresenter は管理者機能のプレゼンター
type AdminPresenter struct{}

// NewAdminPresenter は新しいAdminPresenterを作成
func NewAdminPresenter() *AdminPresenter {
	return &AdminPresenter{}
}

// PresentGrantPoints はポイント付与レスポンスを生成
func (p *AdminPresenter) PresentGrantPoints(resp *inputport.GrantPointsResponse) map[string]interface{} {
	return map[string]interface{}{
		"transaction": TransactionResponse{
			ID:          resp.Transaction.ID,
			FromUserID:  resp.Transaction.FromUserID,
			ToUserID:    resp.Transaction.ToUserID,
			Amount:      resp.Transaction.Amount,
			Description: resp.Transaction.Description,
			CreatedAt:   resp.Transaction.CreatedAt,
		},
		"user": UserResponse{
			ID:          resp.User.ID,
			Username:    resp.User.Username,
			DisplayName: resp.User.DisplayName,
			AvatarURL:   resp.User.AvatarURL,
			Balance:     resp.User.Balance,
			Role:        string(resp.User.Role),
			IsActive:    resp.User.IsActive,
			CreatedAt:   resp.User.CreatedAt,
			UpdatedAt:   resp.User.UpdatedAt,
		},
	}
}

// PresentDeductPoints はポイント減算レスポンスを生成
func (p *AdminPresenter) PresentDeductPoints(resp *inputport.DeductPointsResponse) map[string]interface{} {
	return map[string]interface{}{
		"transaction": TransactionResponse{
			ID:          resp.Transaction.ID,
			FromUserID:  resp.Transaction.FromUserID,
			ToUserID:    resp.Transaction.ToUserID,
			Amount:      resp.Transaction.Amount,
			Description: resp.Transaction.Description,
			CreatedAt:   resp.Transaction.CreatedAt,
		},
		"user": UserResponse{
			ID:          resp.User.ID,
			Username:    resp.User.Username,
			DisplayName: resp.User.DisplayName,
			AvatarURL:   resp.User.AvatarURL,
			Balance:     resp.User.Balance,
			Role:        string(resp.User.Role),
			IsActive:    resp.User.IsActive,
			CreatedAt:   resp.User.CreatedAt,
			UpdatedAt:   resp.User.UpdatedAt,
		},
	}
}

// PresentListAllUsers はユーザー一覧レスポンスを生成
func (p *AdminPresenter) PresentListAllUsers(resp *inputport.ListAllUsersResponse) map[string]interface{} {
	users := make([]UserResponse, 0, len(resp.Users))
	for _, user := range resp.Users {
		users = append(users, UserResponse{
			ID:          user.ID,
			Username:    user.Username,
			DisplayName: user.DisplayName,
			AvatarURL:   user.AvatarURL,
			Balance:     user.Balance,
			Role:        string(user.Role),
			IsActive:    user.IsActive,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
		})
	}

	return map[string]interface{}{
		"users": users,
		"total": resp.Total,
	}
}

// PresentListAllTransactions は取引履歴一覧レスポンスを生成
func (p *AdminPresenter) PresentListAllTransactions(resp *inputport.ListAllTransactionsResponse) map[string]interface{} {
	transactions := make([]TransactionResponse, 0, len(resp.Transactions))
	for _, tx := range resp.Transactions {
		transactions = append(transactions, TransactionResponse{
			ID:          tx.ID,
			FromUserID:  tx.FromUserID,
			ToUserID:    tx.ToUserID,
			Amount:      tx.Amount,
			Description: tx.Description,
			CreatedAt:   tx.CreatedAt,
		})
	}

	return map[string]interface{}{
		"transactions": transactions,
		"total":        resp.Total,
	}
}

// PresentUpdateUserRole はユーザー役割更新レスポンスを生成
func (p *AdminPresenter) PresentUpdateUserRole(resp *inputport.UpdateUserRoleResponse) map[string]interface{} {
	return map[string]interface{}{
		"user": UserResponse{
			ID:          resp.User.ID,
			Username:    resp.User.Username,
			DisplayName: resp.User.DisplayName,
			AvatarURL:   resp.User.AvatarURL,
			Balance:     resp.User.Balance,
			Role:        string(resp.User.Role),
			IsActive:    resp.User.IsActive,
			CreatedAt:   resp.User.CreatedAt,
			UpdatedAt:   resp.User.UpdatedAt,
		},
	}
}

// PresentDeactivateUser はユーザー無効化レスポンスを生成
func (p *AdminPresenter) PresentDeactivateUser(resp *inputport.DeactivateUserResponse) map[string]interface{} {
	return map[string]interface{}{
		"user": UserResponse{
			ID:          resp.User.ID,
			Username:    resp.User.Username,
			DisplayName: resp.User.DisplayName,
			AvatarURL:   resp.User.AvatarURL,
			Balance:     resp.User.Balance,
			Role:        string(resp.User.Role),
			IsActive:    resp.User.IsActive,
			CreatedAt:   resp.User.CreatedAt,
			UpdatedAt:   resp.User.UpdatedAt,
		},
	}
}
