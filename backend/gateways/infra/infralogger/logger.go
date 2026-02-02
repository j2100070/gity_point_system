package infralogger

import (
	"fmt"
	"log"
	"os"

	"github.com/gity/point-system/entities"
)

// LoggerImpl はLoggerの実装
type LoggerImpl struct {
	logger *log.Logger
}

// NewLogger は新しいLoggerを作成
func NewLogger() entities.Logger {
	return &LoggerImpl{
		logger: log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile),
	}
}

// Debug はデバッグログを出力
func (l *LoggerImpl) Debug(msg string, fields ...entities.Field) {
	l.output("DEBUG", msg, fields...)
}

// Info は情報ログを出力
func (l *LoggerImpl) Info(msg string, fields ...entities.Field) {
	l.output("INFO", msg, fields...)
}

// Warn は警告ログを出力
func (l *LoggerImpl) Warn(msg string, fields ...entities.Field) {
	l.output("WARN", msg, fields...)
}

// Error はエラーログを出力
func (l *LoggerImpl) Error(msg string, fields ...entities.Field) {
	l.output("ERROR", msg, fields...)
}

// Fatal は致命的エラーログを出力してプログラムを終了
func (l *LoggerImpl) Fatal(msg string, fields ...entities.Field) {
	l.output("FATAL", msg, fields...)
	os.Exit(1)
}

// output はログを出力
func (l *LoggerImpl) output(level, msg string, fields ...entities.Field) {
	fieldStr := ""
	if len(fields) > 0 {
		fieldStr = " ["
		for i, field := range fields {
			if i > 0 {
				fieldStr += ", "
			}
			fieldStr += fmt.Sprintf("%s=%v", field.Key, field.Value)
		}
		fieldStr += "]"
	}
	l.logger.Printf("[%s] %s%s", level, msg, fieldStr)
}
