package service

// EmailService はメール送信サービスのインターフェース
type EmailService interface {
	// SendVerificationEmail はメール認証用のメールを送信
	SendVerificationEmail(to, token string) error

	// SendPasswordChangeNotification はパスワード変更通知メールを送信
	SendPasswordChangeNotification(to string) error

	// SendAccountDeletedNotification はアカウント削除通知メールを送信
	SendAccountDeletedNotification(to string) error
}
