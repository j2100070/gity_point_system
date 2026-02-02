package presenter

import (
	"github.com/gin-gonic/gin"
	"github.com/gity/point-system/usecases/inputport"
)

// PointPresenter はポイント関連のPresenter
// Interactorから返ってきたEntityを外界が求める出力フォーマットに変更
type PointPresenter struct{}

// NewPointPresenter は新しいPointPresenterを作成
func NewPointPresenter() *PointPresenter {
	return &PointPresenter{}
}

// PresentTransferResponse はTransferResponseをJSON形式に変換
func (p *PointPresenter) PresentTransferResponse(resp *inputport.TransferResponse) gin.H {
	return gin.H{
		"message": "transfer successful",
		"transaction": gin.H{
			"id":          resp.Transaction.ID,
			"amount":      resp.Transaction.Amount,
			"status":      resp.Transaction.Status,
			"created_at":  resp.Transaction.CreatedAt,
		},
		"new_balance": resp.FromUser.Balance,
	}
}

// PresentBalanceResponse はBalanceResponseをJSON形式に変換
func (p *PointPresenter) PresentBalanceResponse(resp *inputport.GetBalanceResponse) gin.H {
	return gin.H{
		"balance": resp.Balance,
		"user": gin.H{
			"id":           resp.User.ID,
			"username":     resp.User.Username,
			"display_name": resp.User.DisplayName,
		},
	}
}

// PresentTransactionHistoryResponse はTransactionHistoryResponseをJSON形式に変換
func (p *PointPresenter) PresentTransactionHistoryResponse(resp *inputport.GetTransactionHistoryResponse) gin.H {
	transactions := make([]gin.H, len(resp.Transactions))
	for i, tx := range resp.Transactions {
		transactions[i] = gin.H{
			"id":               tx.ID,
			"from_user_id":     tx.FromUserID,
			"to_user_id":       tx.ToUserID,
			"amount":           tx.Amount,
			"transaction_type": tx.TransactionType,
			"status":           tx.Status,
			"description":      tx.Description,
			"created_at":       tx.CreatedAt,
		}
	}

	return gin.H{
		"transactions": transactions,
		"total":        resp.Total,
	}
}
