package service

// PasswordService はパスワード処理のサービスインターフェース
type PasswordService interface {
	// HashPassword はパスワードをハッシュ化
	HashPassword(password string) (string, error)

	// VerifyPassword はパスワードを検証
	VerifyPassword(hashedPassword, password string) bool
}
