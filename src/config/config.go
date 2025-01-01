package config

import (
	"os"
	"strings"
	"time"

	"go.uber.org/zap/zapcore"
)

type Config struct {
	LogLevel    zapcore.Level
	Version     string
	BuildTime   time.Time
	ProjectPath string
}

func GetConfig() Config {
	return Config{
		LogLevel:    getLogLevel(),
		Version:     "0.0.1",
		BuildTime:   time.Now().UTC(),
		ProjectPath: getProjectPath(),
	}
}

func getLogLevel() zapcore.Level {
	l := getTrimmedEnvVar("OPC_ENGINE_SIMULATOR_LOG_LEVEL")

	switch strings.ToLower(l) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	}

	return zapcore.InfoLevel
}

func getProjectPath() string {
	p := getTrimmedEnvVar("OPC_ENGINE_SIMULATOR_PROJECT_PATH")
	if p != "" {
		return p
	}
	if len(os.Args) >= 2 {
		return os.Args[1]
	}
	return ""
}

func getTrimmedEnvVar(s string) string {
	return strings.Trim(os.Getenv(s), " \t")
}
