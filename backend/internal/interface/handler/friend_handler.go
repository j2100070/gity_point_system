package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gity/point-system/internal/interface/middleware"
	"github.com/gity/point-system/internal/usecase"
	"github.com/google/uuid"
)

// FriendHandler は友達関連のHTTPハンドラー
type FriendHandler struct {
	friendshipUC *usecase.FriendshipUseCase
}

// NewFriendHandler は新しいFriendHandlerを作成
func NewFriendHandler(friendshipUC *usecase.FriendshipUseCase) *FriendHandler {
	return &FriendHandler{
		friendshipUC: friendshipUC,
	}
}

// SendFriendRequestRequest は友達申請リクエスト
type SendFriendRequestRequest struct {
	AddresseeID string `json:"addressee_id" binding:"required,uuid"`
}

// SendFriendRequest は友達申請を送信
// POST /api/friends/request
func (h *FriendHandler) SendFriendRequest(c *gin.Context) {
	var req SendFriendRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	requesterID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	addresseeID, err := uuid.Parse(req.AddresseeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid addressee_id"})
		return
	}

	resp, err := h.friendshipUC.SendFriendRequest(&usecase.SendFriendRequestRequest{
		RequesterID: requesterID,
		AddresseeID: addresseeID,
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "friend request sent",
		"friendship": gin.H{
			"id":           resp.Friendship.ID,
			"requester_id": resp.Friendship.RequesterID,
			"addressee_id": resp.Friendship.AddresseeID,
			"status":       resp.Friendship.Status,
		},
	})
}

// AcceptFriendRequestRequest は友達申請承認リクエスト
type AcceptFriendRequestRequest struct {
	FriendshipID string `json:"friendship_id" binding:"required,uuid"`
}

// AcceptFriendRequest は友達申請を承認
// POST /api/friends/accept
func (h *FriendHandler) AcceptFriendRequest(c *gin.Context) {
	var req AcceptFriendRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	friendshipID, err := uuid.Parse(req.FriendshipID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid friendship_id"})
		return
	}

	resp, err := h.friendshipUC.AcceptFriendRequest(&usecase.AcceptFriendRequestRequest{
		FriendshipID: friendshipID,
		UserID:       userID,
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "friend request accepted",
		"friendship": gin.H{
			"id":     resp.Friendship.ID,
			"status": resp.Friendship.Status,
		},
	})
}

// RejectFriendRequest は友達申請を拒否
// POST /api/friends/reject
func (h *FriendHandler) RejectFriendRequest(c *gin.Context) {
	var req AcceptFriendRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	friendshipID, err := uuid.Parse(req.FriendshipID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid friendship_id"})
		return
	}

	resp, err := h.friendshipUC.RejectFriendRequest(&usecase.RejectFriendRequestRequest{
		FriendshipID: friendshipID,
		UserID:       userID,
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "friend request rejected",
		"friendship": gin.H{
			"id":     resp.Friendship.ID,
			"status": resp.Friendship.Status,
		},
	})
}

// GetFriends は友達一覧を取得
// GET /api/friends
func (h *FriendHandler) GetFriends(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	offset := 0
	limit := 20
	if c.Query("offset") != "" {
		fmt.Sscanf(c.Query("offset"), "%d", &offset)
	}
	if c.Query("limit") != "" {
		fmt.Sscanf(c.Query("limit"), "%d", &limit)
	}

	resp, err := h.friendshipUC.GetFriends(&usecase.GetFriendsRequest{
		UserID: userID,
		Offset: offset,
		Limit:  limit,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	friends := make([]gin.H, len(resp.Friends))
	for i, f := range resp.Friends {
		friends[i] = gin.H{
			"friendship_id": f.Friendship.ID,
			"friend": gin.H{
				"id":           f.Friend.ID,
				"username":     f.Friend.Username,
				"display_name": f.Friend.DisplayName,
			},
			"status":     f.Friendship.Status,
			"created_at": f.Friendship.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"friends": friends,
	})
}

// GetPendingRequests は保留中の友達申請を取得
// GET /api/friends/pending
func (h *FriendHandler) GetPendingRequests(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	offset := 0
	limit := 20
	if c.Query("offset") != "" {
		fmt.Sscanf(c.Query("offset"), "%d", &offset)
	}
	if c.Query("limit") != "" {
		fmt.Sscanf(c.Query("limit"), "%d", &limit)
	}

	resp, err := h.friendshipUC.GetPendingRequests(&usecase.GetPendingRequestsRequest{
		UserID: userID,
		Offset: offset,
		Limit:  limit,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	requests := make([]gin.H, len(resp.Requests))
	for i, r := range resp.Requests {
		requests[i] = gin.H{
			"friendship_id": r.Friendship.ID,
			"requester": gin.H{
				"id":           r.Requester.ID,
				"username":     r.Requester.Username,
				"display_name": r.Requester.DisplayName,
			},
			"created_at": r.Friendship.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"requests": requests,
	})
}

// RemoveFriendRequest は友達削除リクエスト
type RemoveFriendRequest struct {
	FriendshipID string `json:"friendship_id" binding:"required,uuid"`
}

// RemoveFriend は友達を削除
// DELETE /api/friends/remove
func (h *FriendHandler) RemoveFriend(c *gin.Context) {
	var req RemoveFriendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	friendshipID, err := uuid.Parse(req.FriendshipID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid friendship_id"})
		return
	}

	_, err = h.friendshipUC.RemoveFriend(&usecase.RemoveFriendRequest{
		UserID:       userID,
		FriendshipID: friendshipID,
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "friend removed successfully",
	})
}
