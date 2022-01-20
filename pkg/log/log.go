/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var DefaultLogger *zap.Logger

// Return a new *zap.Logger instance, accept a zap.Config
func NewDefaultLogger(config zap.Config) *zap.Logger {
	loggerMgr := getLogger(config)
	DefaultLogger = loggerMgr
	return loggerMgr
}

// default zap.Config. accept level of int type, vaule between -1 and 5
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

// Return default Logger instance.
func L() *zap.Logger {
	return DefaultLogger
}

// Return default Logger.Sugar instance.
func S() *zap.SugaredLogger {
	return DefaultLogger.Sugar()
}

// Return a new *zap.Logger built from zap.Config
func NewLogger(config zap.Config) *zap.Logger {
	return getLogger(config)
}

// Build a *zap.Logger from zap.Config
func getLogger(config zap.Config) *zap.Logger {
	loggerMgr, err := config.Build()
	if err != nil {
		panic("Error building Zap logger")
	}
	return loggerMgr
}
