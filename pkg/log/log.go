package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var DefaultLogger *zap.Logger

func NewDefaultLogger(config zap.Config) *zap.Logger {
	loggerMgr := getLogger(config)
	DefaultLogger = loggerMgr
	return loggerMgr
}

func DefaultLogConfig(level int) zap.Config {
	logConfig := zap.NewProductionConfig()
	logConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	if level < -1 || level > 5 {
		logConfig.Level.SetLevel(zapcore.InfoLevel)
	} else {
		logConfig.Level.SetLevel(zapcore.Level(level))
	}
	return logConfig
}

func L() *zap.Logger {
	return DefaultLogger
}

func S() *zap.SugaredLogger {
	return DefaultLogger.Sugar()
}

func NewLogger(config zap.Config) *zap.Logger {
	return getLogger(config)
}

func getLogger(config zap.Config) *zap.Logger {
	loggerMgr, err := config.Build()
	if err != nil {
		panic("Error building Zap logger")
	}
	return loggerMgr
}
