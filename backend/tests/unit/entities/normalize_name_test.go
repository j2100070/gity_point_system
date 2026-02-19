package entities_test

import (
	"testing"

	"github.com/gity/point-system/entities"
	"github.com/stretchr/testify/assert"
)

func TestNormalizeName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"半角スペースを除去", "Photosynth 太郎", "photosynth太郎"},
		{"全角スペースを除去", "Photosynth　太郎", "photosynth太郎"},
		{"スペースなし", "Photosynth太郎", "photosynth太郎"},
		{"英語大文字を小文字化", "TARO YAMADA", "taroyamada"},
		{"前後の空白を除去", "  田中太郎  ", "田中太郎"},
		{"空文字", "", ""},
		{"複数スペース", "田中  太郎", "田中太郎"},
		{"全角半角混在スペース", "田中　 太郎", "田中太郎"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := entities.NormalizeName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
