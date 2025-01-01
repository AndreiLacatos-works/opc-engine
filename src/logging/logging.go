package logging

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func MakeLogger(l zapcore.Level) *zap.Logger {
	rotatingLogger := &lumberjack.Logger{
		Filename:   "logs/app.log",
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   false,
	}

	consoleSyncer := zapcore.AddSync(os.Stdout)
	fileSyncer := zapcore.AddSync(rotatingLogger)
	stderrSyncer := zapcore.AddSync(os.Stderr)

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.ConsoleSeparator = " "
	encoderConfig.EncodeTime = customTimeEncoder
	encoderConfig.EncodeLevel = customLevelEncoder
	encoderConfig.EncodeName = customNameEncoder
	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	consoleCore := zapcore.NewCore(encoder, consoleSyncer, l)
	fileCore := zapcore.NewCore(encoder, fileSyncer, l)
	stderrCore := zapcore.NewCore(encoder, stderrSyncer, zapcore.ErrorLevel)

	core := zapcore.NewTee(consoleCore, fileCore, stderrCore)

	return zap.New(core).Named("OPCSIM")
}

func customLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(fmt.Sprintf("[%-5s]", level.CapitalString()))
}

func customNameEncoder(name string, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(fmt.Sprintf("[%s]", strings.ToUpper(name)))
}

func customTimeEncoder(time time.Time, enc zapcore.PrimitiveArrayEncoder) {
	template := "2006-01-02 15:04:05.999"
	formatted := time.Format(template)
	enc.AppendString(fmt.Sprintf("[%s]", formatted+strings.Repeat("0", len(template)-len(formatted))))
}
