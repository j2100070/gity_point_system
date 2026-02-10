package web

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gity/point-system/controllers/web/presenter"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/gity/point-system/usecases/repository"
	"github.com/google/uuid"
)

// FriendController は友達機能のコントローラー
type FriendController struct {
	friendshipUC inputport.FriendshipInputPort
	userRepo     repository.UserRepository
	presenter    *presenter.FriendPresenter
}

// NewFriendController は新しいFriendControllerを作成
func NewFriendController(
	friendshipUC inputport.FriendshipInputPort,
	userRepo repository.UserRepository,
	presenter *presenter.FriendPresenter,
) *FriendController {
	return &FriendController{
		friendshipUC: friendshipUC,
		userRepo:     userRepo,
		presenter:    presenter,
	}
}

// SearchUserByUsername はユーザー名でユーザーを検索
// GET /api/users/search?username=xxx
func (c *FriendController) SearchUserByUsername(ctx *gin.Context) {
	username := ctx.Query("username")
	if username == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}

	user, err := c.userRepo.ReadByUsername(username)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "ユーザーが見つかりません"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":           user.ID.String(),
			"username":     user.Username,
			"display_name": user.DisplayName,
		},
	})
}

// SendFriendRequest は友達申請を送信
// POST /api/friends/requests
func (c *FriendController) SendFriendRequest(ctx *gin.Context) {
	// ログインユーザー取得
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// リクエストボディ解析
	var req struct {
		AddresseeID string `json:"addressee_id" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// UUID変換
	addresseeID, err := uuid.Parse(req.AddresseeID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid addressee_id"})
		return
	}

	// ユースケース実行
	resp, err := c.friendshipUC.SendFriendRequest(&inputport.SendFriendRequestRequest{
		RequesterID: userID.(uuid.UUID),
		AddresseeID: addresseeID,
	})
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// レスポンス生成
	ctx.JSON(http.StatusCreated, c.presenter.PresentSendFriendRequest(resp))
}

// AcceptFriendRequest は友達申請を承認
// POST /api/friends/requests/:id/accept
func (c *FriendController) AcceptFriendRequest(ctx *gin.Context) {
	// ログインユーザー取得
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// パスパラメータ取得
	friendshipID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid friendship_id"})
		return
	}

	// ユースケース実行
	resp, err := c.friendshipUC.AcceptFriendRequest(&inputport.AcceptFriendRequestRequest{
		UserID:       userID.(uuid.UUID),
		FriendshipID: friendshipID,
	})
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// レスポンス生成
	ctx.JSON(http.StatusOK, c.presenter.PresentAcceptFriendRequest(resp))
}

// RejectFriendRequest は友達申請を拒否
// POST /api/friends/requests/:id/reject
func (c *FriendController) RejectFriendRequest(ctx *gin.Context) {
	// ログインユーザー取得
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// パスパラメータ取得
	friendshipID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid friendship_id"})
		return
	}

	// ユースケース実行
	resp, err := c.friendshipUC.RejectFriendRequest(&inputport.RejectFriendRequestRequest{
		UserID:       userID.(uuid.UUID),
		FriendshipID: friendshipID,
	})
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// レスポンス生成
	ctx.JSON(http.StatusOK, c.presenter.PresentRejectFriendRequest(resp))
}

// GetFriends は友達一覧を取得
// GET /api/friends
func (c *FriendController) GetFriends(ctx *gin.Context) {
	// ログインユーザー取得
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// クエリパラメータ取得
	var offset, limit int
	fmt.Sscanf(ctx.Query("offset"), "%d", &offset)
	fmt.Sscanf(ctx.Query("limit"), "%d", &limit)
	if limit == 0 {
		limit = 20
	}

	// ユースケース実行
	resp, err := c.friendshipUC.GetFriends(&inputport.GetFriendsRequest{
		UserID: userID.(uuid.UUID),
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// レスポンス生成
	ctx.JSON(http.StatusOK, c.presenter.PresentGetFriends(resp))
}

// GetPendingRequests は保留中の友達申請を取得
// GET /api/friends/requests
func (c *FriendController) GetPendingRequests(ctx *gin.Context) {
	// ログインユーザー取得
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// クエリパラメータ取得
	var offset, limit int
	fmt.Sscanf(ctx.Query("offset"), "%d", &offset)
	fmt.Sscanf(ctx.Query("limit"), "%d", &limit)
	if limit == 0 {
		limit = 20
	}

	// ユースケース実行
	resp, err := c.friendshipUC.GetPendingRequests(&inputport.GetPendingRequestsRequest{
		UserID: userID.(uuid.UUID),
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// レスポンス生成
	ctx.JSON(http.StatusOK, c.presenter.PresentGetPendingRequests(resp))
}

// RemoveFriend は友達を削除
// DELETE /api/friends/:id
func (c *FriendController) RemoveFriend(ctx *gin.Context) {
	// ログインユーザー取得
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// パスパラメータ取得
	friendshipID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid friendship_id"})
		return
	}

	// ユースケース実行
	resp, err := c.friendshipUC.RemoveFriend(&inputport.RemoveFriendRequest{
		UserID:       userID.(uuid.UUID),
		FriendshipID: friendshipID,
	})
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// レスポンス生成
	ctx.JSON(http.StatusOK, c.presenter.PresentRemoveFriend(resp))
}
