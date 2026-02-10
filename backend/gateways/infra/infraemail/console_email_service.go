package infraemail

import (
	"fmt"

	"github.com/gity/point-system/entities"
	"github.com/gity/point-system/usecases/service"
)

// ConsoleEmailService はコンソールにメールを出力する実装（開発用）
type ConsoleEmailService struct {
	logger entities.Logger
}

// NewConsoleEmailService は新しいConsoleEmailServiceを作成
func NewConsoleEmailService(logger entities.Logger) service.EmailService {
	return &ConsoleEmailService{
		logger: logger,
	}
}

// SendVerificationEmail はメール認証用のメールを送信（コンソール出力）
func (s *ConsoleEmailService) SendVerificationEmail(to, token string) error {
	message := fmt.Sprintf(`
========================================
メール認証
========================================
宛先: %s
件名: メールアドレスの認証

以下のリンクをクリックしてメールアドレスを認証してください：
http://localhost:3000/verify-email?token=%s

このリンクは24時間有効です。
========================================
`, to, token)

	s.logger.Info("Sending verification email", entities.NewField("to", to))
	fmt.Println(message)

	return nil
}

// SendPasswordChangeNotification はパスワード変更通知メールを送信（コンソール出力）
func (s *ConsoleEmailService) SendPasswordChangeNotification(to string) error {
	message := fmt.Sprintf(`
========================================
パスワード変更通知
========================================
宛先: %s
件名: パスワードが変更されました

あなたのアカウントのパスワードが変更されました。

もしこの変更に覚えがない場合は、すぐにサポートに連絡してください。
========================================
`, to)

	s.logger.Info("Sending password change notification", entities.NewField("to", to))
	fmt.Println(message)

	return nil
}

// SendAccountDeletedNotification はアカウント削除通知メールを送信（コンソール出力）
func (s *ConsoleEmailService) SendAccountDeletedNotification(to string) error {
	message := fmt.Sprintf(`
========================================
アカウント削除通知
========================================
宛先: %s
件名: アカウントが削除されました

あなたのアカウントは正常に削除されました。

ご利用ありがとうございました。
========================================
`, to)

	s.logger.Info("Sending account deleted notification", entities.NewField("to", to))
	fmt.Println(message)

	return nil
}
