package main

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// // sugar
	// sugar := zap.NewExample().Sugar()
	// sugar.Infof("Hello name: %s, age: %d\n", "harrydo", 21) // like fmt.Printf(format, args...)

	// // logger
	// logger := zap.NewExample()
	// logger.Info("Hello", zap.String("name", "harrydo"), zap.Int("age", 21))

	// logger := zap.NewExample()
	// logger.Info("Hello")

	// // Development
	// logger, _ = zap.NewDevelopment()
	// logger.Info("Hello development")

	// // Production
	// logger, _ = zap.NewProduction()
	// logger.Info("Hello production")

	// 3. customize
	encoder := getEncoderLog()
	sync := getWriterSync()
	core := zapcore.NewCore(encoder, sync, zapcore.InfoLevel)
	logger := zap.New(core, zap.AddCaller())

	logger.Info("Hello custom", zap.Int("line", 1))
	logger.Error("Hello custom", zap.Int("line", 2))

}

// forrmat log
func getEncoderLog() zapcore.Encoder {
	encodeConfig := zap.NewProductionEncoderConfig()
	// định dạng time
	encodeConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	// định dạng level
	encodeConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	// định dạng caller
	encodeConfig.EncodeCaller = zapcore.ShortCallerEncoder
	return zapcore.NewJSONEncoder(encodeConfig)
}

func getWriterSync() zapcore.WriteSyncer {
	file, _ := os.OpenFile("./log/log.txt", os.O_CREATE|os.O_WRONLY, os.ModePerm)
	syncFile := zapcore.AddSync(file)
	syncConsole := zapcore.AddSync(os.Stderr)
	return zapcore.NewMultiWriteSyncer(syncConsole, syncFile)
}
