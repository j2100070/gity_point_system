package infrapassword

import (
	"github.com/gity/point-system/usecases/service"
	"golang.org/x/crypto/bcrypt"
)

// BcryptPasswordService はbcryptを使用したパスワードサービス
type BcryptPasswordService struct {
	cost int
}

// NewBcryptPasswordService は新しいBcryptPasswordServiceを作成
func NewBcryptPasswordService() service.PasswordService {
	return &BcryptPasswordService{
		cost: bcrypt.DefaultCost, // デフォルトは10
	}
}

// HashPassword はパスワードをハッシュ化
func (s *BcryptPasswordService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword はパスワードを検証
func (s *BcryptPasswordService) VerifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
