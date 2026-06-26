package logger

import (
	"go-app/pkg/setting"
	"os"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LoggerZap struct {
	*zap.Logger
}

func NewLogger(config setting.Logger) *LoggerZap {
	// debug - info - error - warn - fatal - panic - dpanic - derror
	level, err := zapcore.ParseLevel(config.Level)
	if err != nil {
		level = zapcore.InfoLevel // Mặc định là Info (log bình thường) nếu cấu hình không hợp lệ hoặc trống
	}

	// 1. Encoder có màu dành cho Console (Stdout) để dễ nhận diện trực quan khi chạy chương trình
	consoleEncoderConfig := zap.NewProductionEncoderConfig()
	consoleEncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	consoleEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // Màu sắc trực quan (Đỏ: Error, Vàng: Warn, Xanh: Info)
	consoleEncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(consoleEncoderConfig)

	// 2. Encoder KHÔNG có màu dành cho ghi File (để tránh rác ký tự mã màu ANSI như ^[[31m trong file log)
	fileEncoderConfig := zap.NewProductionEncoderConfig()
	fileEncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	fileEncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder // In hoa chữ thường (INFO, ERROR) không có màu
	fileEncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	fileEncoder := zapcore.NewConsoleEncoder(fileEncoderConfig)

	hook := &lumberjack.Logger{
		Filename:   config.Filename,
		MaxSize:    config.MaxSize,
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,
		Compress:   config.Compress,
	}

	// 3. Gộp 2 core: Console (có màu) và File (không màu) thông qua zapcore.NewTee
	consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level)
	fileCore := zapcore.NewCore(fileEncoder, zapcore.AddSync(hook), level)

	core := zapcore.NewTee(consoleCore, fileCore)

	return &LoggerZap{
		zap.New(core,
			zap.AddCaller(),
			zap.AddStacktrace(zap.ErrorLevel),
		),
	}
}
