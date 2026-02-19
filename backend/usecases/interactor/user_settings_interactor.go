package interactor

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/usecases/inputport"
	"github.com/gity/point-system/usecases/repository"
	"github.com/gity/point-system/usecases/service"
)

// UserSettingsInteractor はユーザー設定のユースケース実装
type UserSettingsInteractor struct {
	txManager                 repository.TransactionManager
	userRepo                  repository.UserRepository
	userSettingsRepo          repository.UserSettingsRepository
	archivedUserRepo          repository.ArchivedUserRepository
	emailVerificationRepo     repository.EmailVerificationRepository
	usernameChangeHistoryRepo repository.UsernameChangeHistoryRepository
	passwordChangeHistoryRepo repository.PasswordChangeHistoryRepository
	fileStorageService        service.FileStorageService
	passwordService           service.PasswordService
	emailService              service.EmailService
	logger                    entities.Logger
}

// NewUserSettingsInteractor は新しいUserSettingsInteractorを作成
func NewUserSettingsInteractor(
	txManager repository.TransactionManager,
	userRepo repository.UserRepository,
	userSettingsRepo repository.UserSettingsRepository,
	archivedUserRepo repository.ArchivedUserRepository,
	emailVerificationRepo repository.EmailVerificationRepository,
	usernameChangeHistoryRepo repository.UsernameChangeHistoryRepository,
	passwordChangeHistoryRepo repository.PasswordChangeHistoryRepository,
	fileStorageService service.FileStorageService,
	passwordService service.PasswordService,
	emailService service.EmailService,
	logger entities.Logger,
) inputport.UserSettingsInputPort {
	return &UserSettingsInteractor{
		txManager:                 txManager,
		userRepo:                  userRepo,
		userSettingsRepo:          userSettingsRepo,
		archivedUserRepo:          archivedUserRepo,
		emailVerificationRepo:     emailVerificationRepo,
		usernameChangeHistoryRepo: usernameChangeHistoryRepo,
		passwordChangeHistoryRepo: passwordChangeHistoryRepo,
		fileStorageService:        fileStorageService,
		passwordService:           passwordService,
		emailService:              emailService,
		logger:                    logger,
	}
}

// UpdateProfile はプロフィールを更新
func (i *UserSettingsInteractor) UpdateProfile(ctx context.Context, req *inputport.UpdateProfileRequest) (*inputport.UpdateProfileResponse, error) {
	i.logger.Info("Updating user profile", entities.NewField("user_id", req.UserID))

	// ユーザーを取得
	user, err := i.userRepo.Read(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	emailChanged := false

	// メールアドレスが変更された場合は一意性チェック
	if req.Email != "" && req.Email != user.Email {
		exists, err := i.userSettingsRepo.CheckEmailExists(ctx, req.Email, user.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to check email existence: %w", err)
		}
		if exists {
			return nil, errors.New("email already exists")
		}
		emailChanged = true
	}

	// プロフィールを更新
	if err := user.UpdateProfile(req.DisplayName, req.Email, req.FirstName, req.LastName); err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	// データベースに保存
	success, err := i.userSettingsRepo.UpdateProfile(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to save profile: %w", err)
	}
	if !success {
		return nil, errors.New("profile update failed due to version conflict")
	}

	// メールアドレスが変更された場合は認証メールを送信
	emailVerificationSent := false
	if emailChanged {
		// 古いトークンを削除
		_ = i.emailVerificationRepo.DeleteByUserID(ctx, user.ID)

		// 新しいトークンを作成
		token, err := entities.NewEmailVerificationToken(&user.ID, req.Email, entities.TokenTypeEmailChange)
		if err != nil {
			i.logger.Error("Failed to create email verification token", entities.NewField("error", err))
		} else {
			if err := i.emailVerificationRepo.Create(ctx, token); err != nil {
				i.logger.Error("Failed to save email verification token", entities.NewField("error", err))
			} else {
				// メール送信
				if err := i.emailService.SendVerificationEmail(req.Email, token.Token); err != nil {
					i.logger.Error("Failed to send verification email", entities.NewField("error", err))
				} else {
					emailVerificationSent = true
				}
			}
		}
	}

	i.logger.Info("Profile updated successfully",
		entities.NewField("user_id", user.ID),
		entities.NewField("email_changed", emailChanged))

	return &inputport.UpdateProfileResponse{
		User:                  user,
		EmailVerificationSent: emailVerificationSent,
	}, nil
}

// UpdateUsername はユーザー名を変更
func (i *UserSettingsInteractor) UpdateUsername(ctx context.Context, req *inputport.UpdateUsernameRequest) error {
	i.logger.Info("Updating username", entities.NewField("user_id", req.UserID))

	// ユーザーを取得
	user, err := i.userRepo.Read(ctx, req.UserID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	oldUsername := user.Username

	// 一意性チェック
	exists, err := i.userSettingsRepo.CheckUsernameExists(ctx, req.NewUsername, user.ID)
	if err != nil {
		return fmt.Errorf("failed to check username existence: %w", err)
	}
	if exists {
		return errors.New("username already exists")
	}

	// ユーザー名を更新
	if err := user.UpdateUsername(req.NewUsername); err != nil {
		return fmt.Errorf("failed to update username: %w", err)
	}

	// データベースに保存
	success, err := i.userSettingsRepo.UpdateUsername(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to save username: %w", err)
	}
	if !success {
		return errors.New("username update failed due to version conflict")
	}

	// 変更履歴を記録
	history := entities.NewUsernameChangeHistory(user.ID, oldUsername, req.NewUsername, &user.ID, req.IPAddress)
	if err := i.usernameChangeHistoryRepo.Create(ctx, history); err != nil {
		i.logger.Error("Failed to create username change history", entities.NewField("error", err))
		// 履歴の保存に失敗してもエラーにしない（ユーザー名の変更は成功）
	}

	i.logger.Info("Username updated successfully",
		entities.NewField("user_id", user.ID),
		entities.NewField("old_username", oldUsername),
		entities.NewField("new_username", req.NewUsername))

	return nil
}

// ChangePassword はパスワードを変更
func (i *UserSettingsInteractor) ChangePassword(ctx context.Context, req *inputport.ChangePasswordRequest) error {
	i.logger.Info("Changing password", entities.NewField("user_id", req.UserID))

	// ユーザーを取得
	user, err := i.userRepo.Read(ctx, req.UserID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// 現在のパスワードを検証
	if !i.passwordService.VerifyPassword(user.PasswordHash, req.CurrentPassword) {
		return errors.New("current password is incorrect")
	}

	// 新しいパスワードをハッシュ化
	newHash, err := i.passwordService.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// パスワードを更新
	if err := user.UpdatePassword(newHash); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// データベースに保存
	success, err := i.userSettingsRepo.UpdatePassword(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to save password: %w", err)
	}
	if !success {
		return errors.New("password update failed due to version conflict")
	}

	// 変更履歴を記録
	history := entities.NewPasswordChangeHistory(user.ID, req.IPAddress, req.UserAgent)
	if err := i.passwordChangeHistoryRepo.Create(ctx, history); err != nil {
		i.logger.Error("Failed to create password change history", entities.NewField("error", err))
	}

	// パスワード変更通知メールを送信
	if err := i.emailService.SendPasswordChangeNotification(user.Email); err != nil {
		i.logger.Error("Failed to send password change notification", entities.NewField("error", err))
	}

	i.logger.Info("Password changed successfully", entities.NewField("user_id", user.ID))

	return nil
}

// UploadAvatar はアバター画像をアップロード
func (i *UserSettingsInteractor) UploadAvatar(ctx context.Context, req *inputport.UploadAvatarRequest) (*inputport.UploadAvatarResponse, error) {
	i.logger.Info("Uploading avatar", entities.NewField("user_id", req.UserID))

	// ユーザーを取得
	user, err := i.userRepo.Read(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// 古いアバターのパスを保存（後で削除）
	var oldAvatarPath *string
	if user.AvatarType == entities.AvatarTypeUploaded && user.AvatarURL != nil {
		oldAvatarPath = user.AvatarURL
	}

	// ファイルを保存
	fileReader := bytes.NewReader(req.FileData)
	filePath, err := i.fileStorageService.SaveAvatar(
		req.UserID.String(),
		req.FileName,
		fileReader,
		int64(len(req.FileData)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to save avatar file: %w", err)
	}

	// アバターURLを取得
	avatarURL := i.fileStorageService.GetAvatarURL(filePath)

	// ユーザーのアバターを更新
	if err := user.UpdateAvatar(avatarURL, entities.AvatarTypeUploaded); err != nil {
		// ファイルの削除を試みる
		_ = i.fileStorageService.DeleteAvatar(filePath)
		return nil, fmt.Errorf("failed to update avatar: %w", err)
	}

	// データベースに保存
	success, err := i.userSettingsRepo.UpdateProfile(ctx, user)
	if err != nil {
		// ファイルの削除を試みる
		_ = i.fileStorageService.DeleteAvatar(filePath)
		return nil, fmt.Errorf("failed to save avatar: %w", err)
	}
	if !success {
		// ファイルの削除を試みる
		_ = i.fileStorageService.DeleteAvatar(filePath)
		return nil, errors.New("avatar update failed due to version conflict")
	}

	// 古いアバターを削除
	if oldAvatarPath != nil {
		if err := i.fileStorageService.DeleteAvatar(*oldAvatarPath); err != nil {
			i.logger.Error("Failed to delete old avatar", entities.NewField("error", err))
		}
	}

	i.logger.Info("Avatar uploaded successfully",
		entities.NewField("user_id", user.ID),
		entities.NewField("avatar_url", avatarURL))

	return &inputport.UploadAvatarResponse{
		AvatarURL: avatarURL,
	}, nil
}

// DeleteAvatar はアバターを削除（自動生成に戻す）
func (i *UserSettingsInteractor) DeleteAvatar(ctx context.Context, req *inputport.DeleteAvatarRequest) error {
	i.logger.Info("Deleting avatar", entities.NewField("user_id", req.UserID))

	// ユーザーを取得
	user, err := i.userRepo.Read(ctx, req.UserID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// 削除するアバターのパスを保存
	var avatarPath *string
	if user.AvatarType == entities.AvatarTypeUploaded && user.AvatarURL != nil {
		avatarPath = user.AvatarURL
	}

	// アバターを削除
	user.DeleteAvatar()

	// データベースに保存
	success, err := i.userSettingsRepo.UpdateProfile(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to save avatar deletion: %w", err)
	}
	if !success {
		return errors.New("avatar deletion failed due to version conflict")
	}

	// ファイルを削除
	if avatarPath != nil {
		if err := i.fileStorageService.DeleteAvatar(*avatarPath); err != nil {
			i.logger.Error("Failed to delete avatar file", entities.NewField("error", err))
		}
	}

	i.logger.Info("Avatar deleted successfully", entities.NewField("user_id", user.ID))

	return nil
}

// SendEmailVerification はメール認証メールを送信
func (i *UserSettingsInteractor) SendEmailVerification(ctx context.Context, req *inputport.SendEmailVerificationRequest) error {
	i.logger.Info("Sending email verification", entities.NewField("email", req.Email))

	// 既存のトークンを削除（登録時はUserIDがnil、メール変更時はUserIDあり）
	if req.UserID != nil {
		_ = i.emailVerificationRepo.DeleteByUserID(ctx, *req.UserID)
	}

	// 新しいトークンを作成
	token, err := entities.NewEmailVerificationToken(req.UserID, req.Email, req.TokenType)
	if err != nil {
		return fmt.Errorf("failed to create token: %w", err)
	}

	// トークンを保存
	if err := i.emailVerificationRepo.Create(ctx, token); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	// メールを送信
	if err := i.emailService.SendVerificationEmail(req.Email, token.Token); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	i.logger.Info("Email verification sent successfully", entities.NewField("email", req.Email))

	return nil
}

// VerifyEmail はメールアドレスを認証
func (i *UserSettingsInteractor) VerifyEmail(ctx context.Context, req *inputport.VerifyEmailRequest) (*inputport.VerifyEmailResponse, error) {
	i.logger.Info("Verifying email", entities.NewField("token", req.Token[:10]+"..."))

	// トークンを取得
	token, err := i.emailVerificationRepo.ReadByToken(ctx, req.Token)
	if err != nil {
		return nil, errors.New("invalid or expired token")
	}

	// トークンの有効期限をチェック
	if token.IsExpired() {
		return nil, errors.New("token has expired")
	}

	// 既に検証済みかチェック
	if token.IsVerified() {
		return nil, errors.New("token has already been used")
	}

	// トークンを検証済みにする
	if err := token.Verify(); err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}

	// トークンを更新
	if err := i.emailVerificationRepo.Update(ctx, token); err != nil {
		return nil, fmt.Errorf("failed to update token: %w", err)
	}

	var user *entities.User

	// ユーザーIDがある場合（メール変更）
	if token.UserID != nil {
		user, err = i.userRepo.Read(ctx, *token.UserID)
		if err != nil {
			return nil, fmt.Errorf("user not found: %w", err)
		}

		// メールアドレスを更新して認証
		if err := user.UpdateProfile("", token.Email, "", ""); err != nil {
			return nil, fmt.Errorf("failed to update email: %w", err)
		}

		user.VerifyEmail()

		// データベースに保存
		success, err := i.userSettingsRepo.UpdateProfile(ctx, user)
		if err != nil {
			return nil, fmt.Errorf("failed to save email: %w", err)
		}
		if !success {
			return nil, errors.New("email verification failed due to version conflict")
		}
	}

	i.logger.Info("Email verified successfully", entities.NewField("email", token.Email))

	return &inputport.VerifyEmailResponse{
		User:  user,
		Email: token.Email,
	}, nil
}

// ArchiveAccount はアカウントを削除（アーカイブ）
func (i *UserSettingsInteractor) ArchiveAccount(ctx context.Context, req *inputport.ArchiveAccountRequest) error {
	i.logger.Info("Archiving account", entities.NewField("user_id", req.UserID))

	// ユーザーを取得
	user, err := i.userRepo.Read(ctx, req.UserID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// パスワードを検証
	if !i.passwordService.VerifyPassword(user.PasswordHash, req.Password) {
		return errors.New("password is incorrect")
	}

	// トランザクション開始
	err = i.txManager.Do(ctx, func(ctx context.Context) error {
		// アーカイブユーザーを作成
		archivedUser := user.ToArchivedUser(&req.UserID, req.DeletionReason)

		// アーカイブユーザーを保存
		if err := i.archivedUserRepo.Create(ctx, archivedUser); err != nil {
			return fmt.Errorf("failed to archive user: %w", err)
		}

		// 元のユーザーを削除（論理削除ではなく物理削除）
		if err := i.userRepo.Delete(ctx, user.ID); err != nil {
			return fmt.Errorf("failed to delete user: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	// アバターファイルを削除（トランザクション外で実行）
	if user.AvatarType == entities.AvatarTypeUploaded && user.AvatarURL != nil {
		if err := i.fileStorageService.DeleteAvatar(*user.AvatarURL); err != nil {
			i.logger.Error("Failed to delete avatar file", entities.NewField("error", err))
		}
	}

	// アカウント削除通知メールを送信
	if err := i.emailService.SendAccountDeletedNotification(user.Email); err != nil {
		i.logger.Error("Failed to send account deleted notification", entities.NewField("error", err))
	}

	i.logger.Info("Account archived successfully", entities.NewField("user_id", req.UserID))

	return nil
}

// GetProfile はプロフィール情報を取得
func (i *UserSettingsInteractor) GetProfile(ctx context.Context, req *inputport.GetProfileRequest) (*inputport.GetProfileResponse, error) {
	i.logger.Info("Getting profile", entities.NewField("user_id", req.UserID))

	user, err := i.userRepo.Read(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return &inputport.GetProfileResponse{
		User: user,
	}, nil
}
