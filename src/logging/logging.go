package logging

import (
	"fmt"
	"os"
	"strings"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func MakeLogger() *zap.Logger {
	rotatingLogger := &lumberjack.Logger{
		Filename:   "logs/app.log",
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   false,
	}

	level := zapcore.DebugLevel
	consoleSyncer := zapcore.AddSync(os.Stdout)
	fileSyncer := zapcore.AddSync(rotatingLogger)
	stderrSyncer := zapcore.AddSync(os.Stderr)

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.ConsoleSeparator = " "
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("[2006-01-02 15:04:05.999]")
	encoderConfig.EncodeLevel = customLevelEncoder
	encoderConfig.EncodeName = customNameEncoder
	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	consoleCore := zapcore.NewCore(encoder, consoleSyncer, level)
	fileCore := zapcore.NewCore(encoder, fileSyncer, level)
	stderrCore := zapcore.NewCore(encoder, stderrSyncer, zapcore.ErrorLevel)

	core := zapcore.NewTee(consoleCore, fileCore, stderrCore)

	return zap.New(core).Named("default")
}

func customLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(fmt.Sprintf("[%-5s]", level.CapitalString()))
}

func customNameEncoder(name string, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(fmt.Sprintf("[%s]", strings.ToUpper(name)))
}
