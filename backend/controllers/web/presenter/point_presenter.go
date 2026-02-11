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
	for i, txWithUsers := range resp.Transactions {
		tx := txWithUsers.Transaction
		txData := gin.H{
			"id":               tx.ID,
			"from_user_id":     tx.FromUserID,
			"to_user_id":       tx.ToUserID,
			"amount":           tx.Amount,
			"transaction_type": tx.TransactionType,
			"status":           tx.Status,
			"description":      tx.Description,
			"created_at":       tx.CreatedAt,
		}

		// 送信者情報を追加
		if txWithUsers.FromUser != nil {
			txData["from_user"] = gin.H{
				"id":           txWithUsers.FromUser.ID,
				"username":     txWithUsers.FromUser.Username,
				"display_name": txWithUsers.FromUser.DisplayName,
				"avatar_url":   txWithUsers.FromUser.AvatarURL,
			}
		}

		// 受信者情報を追加
		if txWithUsers.ToUser != nil {
			txData["to_user"] = gin.H{
				"id":           txWithUsers.ToUser.ID,
				"username":     txWithUsers.ToUser.Username,
				"display_name": txWithUsers.ToUser.DisplayName,
				"avatar_url":   txWithUsers.ToUser.AvatarURL,
			}
		}

		transactions[i] = txData
	}

	return gin.H{
		"transactions": transactions,
		"total":        resp.Total,
	}
}
