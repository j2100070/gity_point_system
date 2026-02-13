package web

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gity/point-system/controllers/web/presenter"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/google/uuid"
)

// FriendController は友達機能のコントローラー
type FriendController struct {
	friendshipUC inputport.FriendshipInputPort
	userQueryUC  inputport.UserQueryInputPort
	presenter    *presenter.FriendPresenter
}

// NewFriendController は新しいFriendControllerを作成
func NewFriendController(
	friendshipUC inputport.FriendshipInputPort,
	userQueryUC inputport.UserQueryInputPort,
	presenter *presenter.FriendPresenter,
) *FriendController {
	return &FriendController{
		friendshipUC: friendshipUC,
		userQueryUC:  userQueryUC,
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

	resp, err := c.userQueryUC.SearchUserByUsername(ctx.Request.Context(), &inputport.SearchUserByUsernameRequest{
		Username: username,
	})
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "ユーザーが見つかりません"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":           resp.User.ID.String(),
			"username":     resp.User.Username,
			"display_name": resp.User.DisplayName,
			"avatar_url":   resp.User.AvatarURL,
			"avatar_type":  resp.User.AvatarType,
		},
	})
}

// GetUserByID はユーザーIDでユーザー情報を取得
// GET /api/users/:id
func (c *FriendController) GetUserByID(ctx *gin.Context) {
	userIDStr := ctx.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	resp, err := c.userQueryUC.GetUserByID(ctx.Request.Context(), &inputport.GetUserByIDRequest{
		UserID: userID,
	})
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "ユーザーが見つかりません"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":           resp.User.ID.String(),
			"username":     resp.User.Username,
			"display_name": resp.User.DisplayName,
			"avatar_url":   resp.User.AvatarURL,
			"avatar_type":  resp.User.AvatarType,
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
	resp, err := c.friendshipUC.SendFriendRequest(ctx, &inputport.SendFriendRequestRequest{
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
	resp, err := c.friendshipUC.AcceptFriendRequest(ctx, &inputport.AcceptFriendRequestRequest{
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
	resp, err := c.friendshipUC.RejectFriendRequest(ctx, &inputport.RejectFriendRequestRequest{
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
	resp, err := c.friendshipUC.GetFriends(ctx, &inputport.GetFriendsRequest{
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
	resp, err := c.friendshipUC.GetPendingRequests(ctx, &inputport.GetPendingRequestsRequest{
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
	resp, err := c.friendshipUC.RemoveFriend(ctx, &inputport.RemoveFriendRequest{
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
