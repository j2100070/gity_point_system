//go:build integration
// +build integration

package integration

import (
	"io"
)

// ========================================
// MockPasswordService
// ========================================

type mockPasswordService struct{}

func (m *mockPasswordService) HashPassword(password string) (string, error) {
	return "$2a$10$mock_hashed_" + password, nil
}

func (m *mockPasswordService) VerifyPassword(hashedPassword, password string) bool {
	return hashedPassword == "$2a$10$mock_hashed_"+password
}

// ========================================
// MockEmailService
// ========================================

type mockEmailService struct {
	sentEmails []sentEmail
}

type sentEmail struct {
	To    string
	Type  string
	Token string
}

func (m *mockEmailService) SendVerificationEmail(to, token string) error {
	m.sentEmails = append(m.sentEmails, sentEmail{To: to, Type: "verification", Token: token})
	return nil
}

func (m *mockEmailService) SendPasswordChangeNotification(to string) error {
	m.sentEmails = append(m.sentEmails, sentEmail{To: to, Type: "password_change"})
	return nil
}

func (m *mockEmailService) SendAccountDeletedNotification(to string) error {
	m.sentEmails = append(m.sentEmails, sentEmail{To: to, Type: "account_deleted"})
	return nil
}

// ========================================
// MockFileStorageService
// ========================================

type mockFileStorageService struct {
	savedFiles []string
}

func (m *mockFileStorageService) SaveAvatar(userID string, fileName string, file io.Reader, fileSize int64) (string, error) {
	path := "/avatars/" + userID + "/" + fileName
	m.savedFiles = append(m.savedFiles, path)
	return path, nil
}

func (m *mockFileStorageService) DeleteAvatar(filePath string) error {
	return nil
}

func (m *mockFileStorageService) GetAvatarURL(filePath string) string {
	return "http://localhost:8080" + filePath
}
