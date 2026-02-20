package interactor_test

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/gity/point-system/usecases/interactor"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========================================
// UserSettingsInteractor テスト
// ========================================

// --- Mock UserSettingsRepository ---

type mockUserSettingsRepo struct {
	updateProfileOK   bool
	updateUsernameOK  bool
	updatePasswordOK  bool
	usernameExists    bool
	updateProfileErr  error
	updateUsernameErr error
	updatePasswordErr error
}

func newMockUserSettingsRepo() *mockUserSettingsRepo {
	return &mockUserSettingsRepo{
		updateProfileOK:  true,
		updateUsernameOK: true,
		updatePasswordOK: true,
	}
}

func (m *mockUserSettingsRepo) UpdateProfile(ctx context.Context, user *entities.User) (bool, error) {
	if m.updateProfileErr != nil {
		return false, m.updateProfileErr
	}
	return m.updateProfileOK, nil
}
func (m *mockUserSettingsRepo) UpdateUsername(ctx context.Context, user *entities.User) (bool, error) {
	if m.updateUsernameErr != nil {
		return false, m.updateUsernameErr
	}
	return m.updateUsernameOK, nil
}
func (m *mockUserSettingsRepo) UpdatePassword(ctx context.Context, user *entities.User) (bool, error) {
	if m.updatePasswordErr != nil {
		return false, m.updatePasswordErr
	}
	return m.updatePasswordOK, nil
}
func (m *mockUserSettingsRepo) CheckUsernameExists(ctx context.Context, username string, excludeUserID uuid.UUID) (bool, error) {
	return m.usernameExists, nil
}
func (m *mockUserSettingsRepo) CheckEmailExists(ctx context.Context, email string, excludeUserID uuid.UUID) (bool, error) {
	return false, nil
}

// --- Mock ArchivedUserRepository ---

type mockArchivedUserRepo struct {
	createErr error
}

func (m *mockArchivedUserRepo) Create(ctx context.Context, user *entities.ArchivedUser) error {
	if m.createErr != nil {
		return m.createErr
	}
	return nil
}
func (m *mockArchivedUserRepo) Read(ctx context.Context, id uuid.UUID) (*entities.ArchivedUser, error) {
	return nil, nil
}
func (m *mockArchivedUserRepo) ReadByUsername(ctx context.Context, username string) (*entities.ArchivedUser, error) {
	return nil, nil
}
func (m *mockArchivedUserRepo) ReadList(ctx context.Context, offset, limit int) ([]*entities.ArchivedUser, error) {
	return nil, nil
}
func (m *mockArchivedUserRepo) Count(ctx context.Context) (int64, error) {
	return 0, nil
}
func (m *mockArchivedUserRepo) Restore(ctx context.Context, tx interface{}, archivedUser *entities.ArchivedUser, user *entities.User) error {
	return nil
}

// --- Mock EmailVerificationRepository ---

type mockEmailVerificationRepo struct {
	tokens    map[string]*entities.EmailVerificationToken
	createErr error
}

func newMockEmailVerificationRepo() *mockEmailVerificationRepo {
	return &mockEmailVerificationRepo{
		tokens: make(map[string]*entities.EmailVerificationToken),
	}
}

func (m *mockEmailVerificationRepo) Create(ctx context.Context, token *entities.EmailVerificationToken) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.tokens[token.Token] = token
	return nil
}
func (m *mockEmailVerificationRepo) ReadByToken(ctx context.Context, token string) (*entities.EmailVerificationToken, error) {
	t, ok := m.tokens[token]
	if !ok {
		return nil, errors.New("token not found")
	}
	return t, nil
}
func (m *mockEmailVerificationRepo) Update(ctx context.Context, token *entities.EmailVerificationToken) error {
	m.tokens[token.Token] = token
	return nil
}
func (m *mockEmailVerificationRepo) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return nil
}
func (m *mockEmailVerificationRepo) DeleteExpired(ctx context.Context) error { return nil }

// --- Mock UsernameChangeHistoryRepository ---

type mockUsernameChangeHistoryRepo struct{}

func (m *mockUsernameChangeHistoryRepo) Create(ctx context.Context, history *entities.UsernameChangeHistory) error {
	return nil
}
func (m *mockUsernameChangeHistoryRepo) ReadListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entities.UsernameChangeHistory, error) {
	return nil, nil
}
func (m *mockUsernameChangeHistoryRepo) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	return 0, nil
}

// --- Mock PasswordChangeHistoryRepository ---

type mockPasswordChangeHistoryRepo struct{}

func (m *mockPasswordChangeHistoryRepo) Create(ctx context.Context, history *entities.PasswordChangeHistory) error {
	return nil
}
func (m *mockPasswordChangeHistoryRepo) ReadListByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entities.PasswordChangeHistory, error) {
	return nil, nil
}
func (m *mockPasswordChangeHistoryRepo) CountByUserID(ctx context.Context, userID uuid.UUID) (int64, error) {
	return 0, nil
}

// --- Mock FileStorageService ---

type mockFileStorageService struct {
	savedPath string
	saveErr   error
	deleteErr error
	avatarURL string
}

func (m *mockFileStorageService) SaveAvatar(userID, fileName string, file io.Reader, size int64) (string, error) {
	if m.saveErr != nil {
		return "", m.saveErr
	}
	if m.savedPath != "" {
		return m.savedPath, nil
	}
	return "avatars/" + userID + "/" + fileName, nil
}
func (m *mockFileStorageService) DeleteAvatar(path string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	return nil
}
func (m *mockFileStorageService) GetAvatarURL(path string) string {
	if m.avatarURL != "" {
		return m.avatarURL
	}
	return "/uploads/" + path
}

// --- Mock EmailService ---

type mockEmailService struct {
	sendVerificationErr  error
	sentVerificationAddr string
}

func (m *mockEmailService) SendVerificationEmail(email, token string) error {
	m.sentVerificationAddr = email
	return m.sendVerificationErr
}
func (m *mockEmailService) SendPasswordChangeNotification(email string) error {
	return nil
}
func (m *mockEmailService) SendPasswordChangedNotification(email string) error {
	return nil
}
func (m *mockEmailService) SendAccountDeletedNotification(email string) error {
	return nil
}

// ========================================
// Tests
// ========================================

// --- UpdateProfile ---

func TestUserSettingsInteractor_UpdateProfile(t *testing.T) {
	setup := func() (*ctxTrackingUserRepo, *mockUserSettingsRepo, inputport.UserSettingsInputPort) {
		userRepo := newCtxTrackingUserRepo()
		settingsRepo := newMockUserSettingsRepo()
		sut := interactor.NewUserSettingsInteractor(
			&ctxTrackingTxManager{}, userRepo, settingsRepo,
			&mockArchivedUserRepo{}, newMockEmailVerificationRepo(),
			&mockUsernameChangeHistoryRepo{}, &mockPasswordChangeHistoryRepo{},
			&mockFileStorageService{}, &mockPasswordService{verifyOK: true},
			&mockEmailService{}, &mockLogger{},
		)
		return userRepo, settingsRepo, sut
	}

	t.Run("正常にプロフィールを更新できる", func(t *testing.T) {
		userRepo, _, sut := setup()
		user := createTestUserWithBalance(t, "settingsuser", 1000, "user")
		userRepo.setUser(user)

		resp, err := sut.UpdateProfile(context.Background(), &inputport.UpdateProfileRequest{
			UserID: user.ID, DisplayName: "New Display Name",
			Email: user.Email, FirstName: "新太郎", LastName: "新田中",
		})
		require.NoError(t, err)
		assert.NotNil(t, resp.User)
		assert.Equal(t, "New Display Name", resp.User.DisplayName)
	})

	t.Run("ユーザーが存在しない場合エラー", func(t *testing.T) {
		_, _, sut := setup()

		_, err := sut.UpdateProfile(context.Background(), &inputport.UpdateProfileRequest{
			UserID: uuid.New(), DisplayName: "test",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
	})

	t.Run("バージョン競合の場合エラー", func(t *testing.T) {
		userRepo, settingsRepo, sut := setup()
		user := createTestUserWithBalance(t, "conflict", 1000, "user")
		userRepo.setUser(user)
		settingsRepo.updateProfileOK = false

		_, err := sut.UpdateProfile(context.Background(), &inputport.UpdateProfileRequest{
			UserID: user.ID, DisplayName: "test", Email: user.Email,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "version conflict")
	})
}

// --- UpdateUsername ---

func TestUserSettingsInteractor_UpdateUsername(t *testing.T) {
	setup := func() (*ctxTrackingUserRepo, *mockUserSettingsRepo, inputport.UserSettingsInputPort) {
		userRepo := newCtxTrackingUserRepo()
		settingsRepo := newMockUserSettingsRepo()
		sut := interactor.NewUserSettingsInteractor(
			&ctxTrackingTxManager{}, userRepo, settingsRepo,
			&mockArchivedUserRepo{}, newMockEmailVerificationRepo(),
			&mockUsernameChangeHistoryRepo{}, &mockPasswordChangeHistoryRepo{},
			&mockFileStorageService{}, &mockPasswordService{verifyOK: true},
			&mockEmailService{}, &mockLogger{},
		)
		return userRepo, settingsRepo, sut
	}

	t.Run("正常にユーザー名を変更できる", func(t *testing.T) {
		userRepo, _, sut := setup()
		user := createTestUserWithBalance(t, "oldname", 1000, "user")
		userRepo.setUser(user)

		err := sut.UpdateUsername(context.Background(), &inputport.UpdateUsernameRequest{
			UserID: user.ID, NewUsername: "newname",
		})
		assert.NoError(t, err)
	})

	t.Run("既に使われているユーザー名の場合エラー", func(t *testing.T) {
		userRepo, settingsRepo, sut := setup()
		user := createTestUserWithBalance(t, "oldname", 1000, "user")
		userRepo.setUser(user)
		settingsRepo.usernameExists = true

		err := sut.UpdateUsername(context.Background(), &inputport.UpdateUsernameRequest{
			UserID: user.ID, NewUsername: "taken",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already")
	})
}

// --- ChangePassword ---

func TestUserSettingsInteractor_ChangePassword(t *testing.T) {
	setup := func() (*ctxTrackingUserRepo, *mockPasswordService, inputport.UserSettingsInputPort) {
		userRepo := newCtxTrackingUserRepo()
		pwService := &mockPasswordService{verifyOK: true}
		sut := interactor.NewUserSettingsInteractor(
			&ctxTrackingTxManager{}, userRepo, newMockUserSettingsRepo(),
			&mockArchivedUserRepo{}, newMockEmailVerificationRepo(),
			&mockUsernameChangeHistoryRepo{}, &mockPasswordChangeHistoryRepo{},
			&mockFileStorageService{}, pwService,
			&mockEmailService{}, &mockLogger{},
		)
		return userRepo, pwService, sut
	}

	t.Run("正常にパスワードを変更できる", func(t *testing.T) {
		userRepo, _, sut := setup()
		user := createTestUserWithBalance(t, "pwuser", 1000, "user")
		userRepo.setUser(user)

		err := sut.ChangePassword(context.Background(), &inputport.ChangePasswordRequest{
			UserID: user.ID, CurrentPassword: "oldpass", NewPassword: "newpass123",
		})
		assert.NoError(t, err)
	})

	t.Run("現在のパスワードが間違っている場合エラー", func(t *testing.T) {
		userRepo, pwService, sut := setup()
		pwService.verifyOK = false
		user := createTestUserWithBalance(t, "pwuser", 1000, "user")
		userRepo.setUser(user)

		err := sut.ChangePassword(context.Background(), &inputport.ChangePasswordRequest{
			UserID: user.ID, CurrentPassword: "wrong", NewPassword: "newpass123",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "password")
	})
}

// --- UploadAvatar ---

func TestUserSettingsInteractor_UploadAvatar(t *testing.T) {
	setup := func() (*ctxTrackingUserRepo, *mockFileStorageService, inputport.UserSettingsInputPort) {
		userRepo := newCtxTrackingUserRepo()
		fsService := &mockFileStorageService{}
		sut := interactor.NewUserSettingsInteractor(
			&ctxTrackingTxManager{}, userRepo, newMockUserSettingsRepo(),
			&mockArchivedUserRepo{}, newMockEmailVerificationRepo(),
			&mockUsernameChangeHistoryRepo{}, &mockPasswordChangeHistoryRepo{},
			fsService, &mockPasswordService{verifyOK: true},
			&mockEmailService{}, &mockLogger{},
		)
		return userRepo, fsService, sut
	}

	t.Run("正常にアバターをアップロードできる", func(t *testing.T) {
		userRepo, _, sut := setup()
		user := createTestUserWithBalance(t, "avatar_user", 1000, "user")
		userRepo.setUser(user)

		resp, err := sut.UploadAvatar(context.Background(), &inputport.UploadAvatarRequest{
			UserID: user.ID, FileData: []byte("fake-image-data"),
			FileName: "avatar.png", ContentType: "image/png",
		})
		require.NoError(t, err)
		assert.NotEmpty(t, resp.AvatarURL)
	})

	t.Run("ファイル保存に失敗した場合エラー", func(t *testing.T) {
		userRepo, fsService, sut := setup()
		fsService.saveErr = errors.New("storage error")
		user := createTestUserWithBalance(t, "avatar_user", 1000, "user")
		userRepo.setUser(user)

		_, err := sut.UploadAvatar(context.Background(), &inputport.UploadAvatarRequest{
			UserID: user.ID, FileData: []byte("fake-image-data"),
			FileName: "avatar.png", ContentType: "image/png",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save avatar")
	})
}

// --- DeleteAvatar ---

func TestUserSettingsInteractor_DeleteAvatar(t *testing.T) {
	setup := func() (*ctxTrackingUserRepo, inputport.UserSettingsInputPort) {
		userRepo := newCtxTrackingUserRepo()
		sut := interactor.NewUserSettingsInteractor(
			&ctxTrackingTxManager{}, userRepo, newMockUserSettingsRepo(),
			&mockArchivedUserRepo{}, newMockEmailVerificationRepo(),
			&mockUsernameChangeHistoryRepo{}, &mockPasswordChangeHistoryRepo{},
			&mockFileStorageService{}, &mockPasswordService{verifyOK: true},
			&mockEmailService{}, &mockLogger{},
		)
		return userRepo, sut
	}

	t.Run("正常にアバターを削除できる", func(t *testing.T) {
		userRepo, sut := setup()
		user := createTestUserWithBalance(t, "del_avatar", 1000, "user")
		userRepo.setUser(user)

		err := sut.DeleteAvatar(context.Background(), &inputport.DeleteAvatarRequest{
			UserID: user.ID,
		})
		assert.NoError(t, err)
	})

	t.Run("ユーザーが存在しない場合エラー", func(t *testing.T) {
		_, sut := setup()

		err := sut.DeleteAvatar(context.Background(), &inputport.DeleteAvatarRequest{
			UserID: uuid.New(),
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
	})
}

// --- SendEmailVerification ---

func TestUserSettingsInteractor_SendEmailVerification(t *testing.T) {
	setup := func() (*mockEmailService, *mockEmailVerificationRepo, inputport.UserSettingsInputPort) {
		emailService := &mockEmailService{}
		emailVerifRepo := newMockEmailVerificationRepo()
		sut := interactor.NewUserSettingsInteractor(
			&ctxTrackingTxManager{}, newCtxTrackingUserRepo(), newMockUserSettingsRepo(),
			&mockArchivedUserRepo{}, emailVerifRepo,
			&mockUsernameChangeHistoryRepo{}, &mockPasswordChangeHistoryRepo{},
			&mockFileStorageService{}, &mockPasswordService{verifyOK: true},
			emailService, &mockLogger{},
		)
		return emailService, emailVerifRepo, sut
	}

	t.Run("正常に認証メールを送信できる", func(t *testing.T) {
		emailService, _, sut := setup()

		err := sut.SendEmailVerification(context.Background(), &inputport.SendEmailVerificationRequest{
			Email: "new@example.com", TokenType: entities.TokenTypeRegistration,
		})
		assert.NoError(t, err)
		assert.Equal(t, "new@example.com", emailService.sentVerificationAddr)
	})

	t.Run("メール送信に失敗した場合エラー", func(t *testing.T) {
		emailService, _, sut := setup()
		emailService.sendVerificationErr = errors.New("smtp error")

		err := sut.SendEmailVerification(context.Background(), &inputport.SendEmailVerificationRequest{
			Email: "fail@example.com", TokenType: entities.TokenTypeRegistration,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to send email")
	})
}

// --- ArchiveAccount ---

func TestUserSettingsInteractor_ArchiveAccount(t *testing.T) {
	setup := func() (*ctxTrackingUserRepo, *mockPasswordService, inputport.UserSettingsInputPort) {
		userRepo := newCtxTrackingUserRepo()
		pwService := &mockPasswordService{verifyOK: true}
		sut := interactor.NewUserSettingsInteractor(
			&ctxTrackingTxManager{}, userRepo, newMockUserSettingsRepo(),
			&mockArchivedUserRepo{}, newMockEmailVerificationRepo(),
			&mockUsernameChangeHistoryRepo{}, &mockPasswordChangeHistoryRepo{},
			&mockFileStorageService{}, pwService,
			&mockEmailService{}, &mockLogger{},
		)
		return userRepo, pwService, sut
	}

	t.Run("正常にアカウントを削除できる", func(t *testing.T) {
		userRepo, _, sut := setup()
		user := createTestUserWithBalance(t, "archive_me", 1000, "user")
		userRepo.setUser(user)

		err := sut.ArchiveAccount(context.Background(), &inputport.ArchiveAccountRequest{
			UserID: user.ID, Password: "password123",
		})
		assert.NoError(t, err)
	})

	t.Run("パスワードが不正な場合エラー", func(t *testing.T) {
		userRepo, pwService, sut := setup()
		pwService.verifyOK = false
		user := createTestUserWithBalance(t, "archive_me", 1000, "user")
		userRepo.setUser(user)

		err := sut.ArchiveAccount(context.Background(), &inputport.ArchiveAccountRequest{
			UserID: user.ID, Password: "wrong",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "password is incorrect")
	})
}

// --- GetProfile ---

func TestUserSettingsInteractor_GetProfile(t *testing.T) {
	setup := func() (*ctxTrackingUserRepo, inputport.UserSettingsInputPort) {
		userRepo := newCtxTrackingUserRepo()
		sut := interactor.NewUserSettingsInteractor(
			&ctxTrackingTxManager{}, userRepo, newMockUserSettingsRepo(),
			&mockArchivedUserRepo{}, newMockEmailVerificationRepo(),
			&mockUsernameChangeHistoryRepo{}, &mockPasswordChangeHistoryRepo{},
			&mockFileStorageService{}, &mockPasswordService{verifyOK: true},
			&mockEmailService{}, &mockLogger{},
		)
		return userRepo, sut
	}

	t.Run("正常にプロフィールを取得できる", func(t *testing.T) {
		userRepo, sut := setup()
		user := createTestUserWithBalance(t, "profile_user", 5000, "user")
		userRepo.setUser(user)

		resp, err := sut.GetProfile(context.Background(), &inputport.GetProfileRequest{
			UserID: user.ID,
		})
		require.NoError(t, err)
		assert.Equal(t, user.ID, resp.User.ID)
	})

	t.Run("ユーザーが存在しない場合エラー", func(t *testing.T) {
		_, sut := setup()

		_, err := sut.GetProfile(context.Background(), &inputport.GetProfileRequest{
			UserID: uuid.New(),
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
	})
}
