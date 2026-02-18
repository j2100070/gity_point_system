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
			ID:              resp.Transaction.ID,
			FromUserID:      resp.Transaction.FromUserID,
			ToUserID:        resp.Transaction.ToUserID,
			Amount:          resp.Transaction.Amount,
			TransactionType: string(resp.Transaction.TransactionType),
			Status:          string(resp.Transaction.Status),
			Description:     resp.Transaction.Description,
			CreatedAt:       resp.Transaction.CreatedAt,
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
			ID:              resp.Transaction.ID,
			FromUserID:      resp.Transaction.FromUserID,
			ToUserID:        resp.Transaction.ToUserID,
			Amount:          resp.Transaction.Amount,
			TransactionType: string(resp.Transaction.TransactionType),
			Status:          string(resp.Transaction.Status),
			Description:     resp.Transaction.Description,
			CreatedAt:       resp.Transaction.CreatedAt,
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
	for _, txWithUsers := range resp.Transactions {
		tx := txWithUsers.Transaction
		txResp := TransactionResponse{
			ID:              tx.ID,
			FromUserID:      tx.FromUserID,
			ToUserID:        tx.ToUserID,
			Amount:          tx.Amount,
			TransactionType: string(tx.TransactionType),
			Status:          string(tx.Status),
			Description:     tx.Description,
			CreatedAt:       tx.CreatedAt,
		}

		// 送信者情報を追加
		if txWithUsers.FromUser != nil {
			txResp.FromUser = &UserResponse{
				ID:          txWithUsers.FromUser.ID,
				Username:    txWithUsers.FromUser.Username,
				DisplayName: txWithUsers.FromUser.DisplayName,
				AvatarURL:   txWithUsers.FromUser.AvatarURL,
				Balance:     txWithUsers.FromUser.Balance,
				Role:        string(txWithUsers.FromUser.Role),
				IsActive:    txWithUsers.FromUser.IsActive,
				CreatedAt:   txWithUsers.FromUser.CreatedAt,
				UpdatedAt:   txWithUsers.FromUser.UpdatedAt,
			}
		}

		// 受信者情報を追加
		if txWithUsers.ToUser != nil {
			txResp.ToUser = &UserResponse{
				ID:          txWithUsers.ToUser.ID,
				Username:    txWithUsers.ToUser.Username,
				DisplayName: txWithUsers.ToUser.DisplayName,
				AvatarURL:   txWithUsers.ToUser.AvatarURL,
				Balance:     txWithUsers.ToUser.Balance,
				Role:        string(txWithUsers.ToUser.Role),
				IsActive:    txWithUsers.ToUser.IsActive,
				CreatedAt:   txWithUsers.ToUser.CreatedAt,
				UpdatedAt:   txWithUsers.ToUser.UpdatedAt,
			}
		}

		transactions = append(transactions, txResp)
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

// PresentAnalytics は分析データレスポンスを生成
func (p *AdminPresenter) PresentAnalytics(resp *inputport.GetAnalyticsResponse) map[string]interface{} {
	topHolders := make([]map[string]interface{}, 0, len(resp.TopHolders))
	for _, h := range resp.TopHolders {
		topHolders = append(topHolders, map[string]interface{}{
			"user_id":      h.UserID,
			"username":     h.Username,
			"display_name": h.DisplayName,
			"balance":      h.Balance,
			"percentage":   h.Percentage,
		})
	}

	dailyStats := make([]map[string]interface{}, 0, len(resp.DailyStats))
	for _, d := range resp.DailyStats {
		dailyStats = append(dailyStats, map[string]interface{}{
			"date":        d.Date.Format("2006-01-02"),
			"issued":      d.Issued,
			"consumed":    d.Consumed,
			"transferred": d.Transferred,
		})
	}

	typeBreakdown := make([]map[string]interface{}, 0, len(resp.TransactionTypeBreakdown))
	for _, t := range resp.TransactionTypeBreakdown {
		typeBreakdown = append(typeBreakdown, map[string]interface{}{
			"type":         t.Type,
			"count":        t.Count,
			"total_amount": t.TotalAmount,
		})
	}

	return map[string]interface{}{
		"summary": map[string]interface{}{
			"total_points_in_circulation": resp.Summary.TotalPointsInCirculation,
			"average_balance":             resp.Summary.AverageBalance,
			"points_issued_this_month":    resp.Summary.PointsIssuedThisMonth,
			"transactions_this_month":     resp.Summary.TransactionsThisMonth,
			"active_users":                resp.Summary.ActiveUsers,
		},
		"top_holders":                topHolders,
		"daily_stats":                dailyStats,
		"transaction_type_breakdown": typeBreakdown,
	}
}
