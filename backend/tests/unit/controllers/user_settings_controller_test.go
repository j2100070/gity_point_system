package controllers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gity/point-system/controllers/web"
	"github.com/gity/point-system/controllers/web/presenter"
	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserSettingsInputPort はUserSettingsInputPortのモック
type MockUserSettingsInputPort struct {
	mock.Mock
}

func (m *MockUserSettingsInputPort) UpdateProfile(req *inputport.UpdateProfileRequest) (*inputport.UpdateProfileResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*inputport.UpdateProfileResponse), args.Error(1)
}

func (m *MockUserSettingsInputPort) UpdateUsername(req *inputport.UpdateUsernameRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func (m *MockUserSettingsInputPort) ChangePassword(req *inputport.ChangePasswordRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func (m *MockUserSettingsInputPort) UploadAvatar(req *inputport.UploadAvatarRequest) (*inputport.UploadAvatarResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*inputport.UploadAvatarResponse), args.Error(1)
}

func (m *MockUserSettingsInputPort) DeleteAvatar(req *inputport.DeleteAvatarRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func (m *MockUserSettingsInputPort) SendEmailVerification(req *inputport.SendEmailVerificationRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func (m *MockUserSettingsInputPort) VerifyEmail(req *inputport.VerifyEmailRequest) (*inputport.VerifyEmailResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*inputport.VerifyEmailResponse), args.Error(1)
}

func (m *MockUserSettingsInputPort) ArchiveAccount(req *inputport.ArchiveAccountRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func (m *MockUserSettingsInputPort) GetProfile(req *inputport.GetProfileRequest) (*inputport.GetProfileResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*inputport.GetProfileResponse), args.Error(1)
}

// テスト用のヘルパー関数
func setupTestController() (*web.UserSettingsController, *MockUserSettingsInputPort) {
	mockUC := new(MockUserSettingsInputPort)
	presenter := presenter.NewUserSettingsPresenter()
	controller := web.NewUserSettingsController(mockUC, presenter)
	return controller, mockUC
}

func setupTestContext(method, path string, body interface{}) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	var req *http.Request
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		req = httptest.NewRequest(method, path, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}

	c.Request = req
	return c, w
}

// TestUpdateProfile はUpdateProfileメソッドのテスト
func TestUpdateProfile(t *testing.T) {
	controller, mockUC := setupTestController()

	t.Run("成功: プロフィール更新", func(t *testing.T) {
		userID := uuid.New()
		reqBody := web.UpdateProfileRequest{
			DisplayName: "Updated Name",
			Email:       "updated@example.com",
		}

		c, w := setupTestContext("PUT", "/api/settings/profile", reqBody)
		c.Set("userID", userID)

		mockUser := &entities.User{
			ID:          userID,
			Username:    "testuser",
			Email:       "updated@example.com",
			DisplayName: "Updated Name",
		}

		mockUC.On("UpdateProfile", mock.AnythingOfType("*inputport.UpdateProfileRequest")).
			Return(&inputport.UpdateProfileResponse{
				User:                  mockUser,
				EmailVerificationSent: false,
			}, nil)

		controller.UpdateProfile(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUC.AssertExpectations(t)
	})

	t.Run("失敗: ユーザー未認証", func(t *testing.T) {
		reqBody := web.UpdateProfileRequest{
			DisplayName: "Updated Name",
			Email:       "updated@example.com",
		}

		c, w := setupTestContext("PUT", "/api/settings/profile", reqBody)
		// userIDをセットしない

		controller.UpdateProfile(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// TestChangePassword はChangePasswordメソッドのテスト
func TestChangePassword(t *testing.T) {
	controller, mockUC := setupTestController()

	t.Run("成功: パスワード変更", func(t *testing.T) {
		userID := uuid.New()
		reqBody := web.ChangePasswordRequest{
			CurrentPassword: "oldpassword",
			NewPassword:     "newpassword123",
		}

		c, w := setupTestContext("PUT", "/api/settings/password", reqBody)
		c.Set("userID", userID)

		mockUC.On("ChangePassword", mock.AnythingOfType("*inputport.ChangePasswordRequest")).
			Return(nil)

		controller.ChangePassword(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUC.AssertExpectations(t)
	})

	t.Run("失敗: ユーザー未認証", func(t *testing.T) {
		reqBody := web.ChangePasswordRequest{
			CurrentPassword: "oldpassword",
			NewPassword:     "newpassword123",
		}

		c, w := setupTestContext("PUT", "/api/settings/password", reqBody)
		// userIDをセットしない

		controller.ChangePassword(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// TestGetProfile はGetProfileメソッドのテスト
func TestGetProfile(t *testing.T) {
	controller, mockUC := setupTestController()

	t.Run("成功: プロフィール取得", func(t *testing.T) {
		userID := uuid.New()

		c, w := setupTestContext("GET", "/api/settings/profile", nil)
		c.Set("userID", userID)

		mockUser := &entities.User{
			ID:          userID,
			Username:    "testuser",
			Email:       "test@example.com",
			DisplayName: "Test User",
			Balance:     1000,
			Role:        entities.RoleUser,
		}

		mockUC.On("GetProfile", mock.AnythingOfType("*inputport.GetProfileRequest")).
			Return(&inputport.GetProfileResponse{
				User: mockUser,
			}, nil)

		controller.GetProfile(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUC.AssertExpectations(t)
	})

	t.Run("失敗: ユーザー未認証", func(t *testing.T) {
		c, w := setupTestContext("GET", "/api/settings/profile", nil)
		// userIDをセットしない

		controller.GetProfile(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
