/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package logging

import (
	"context"

	"github.com/go-chi/chi/v5/middleware"
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
	if DefaultLogger == nil {
		NewDefaultLogger(DefaultLogConfig(0))
	}
	return DefaultLogger
}

// Return default Logger.Sugar instance.
func S() *zap.SugaredLogger {
	if DefaultLogger == nil {
		NewDefaultLogger(DefaultLogConfig(0))
	}
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

// Use S().Infof, add contextID
func Infof(ctx context.Context, message string, args ...interface{}) {
	S().Infof("["+getContextID(ctx)+"] "+message, args...)
}

// Use S().Errorf, add contextID
func Errorf(ctx context.Context, message string, args ...interface{}) {
	S().Errorf("["+getContextID(ctx)+"] "+message, args...)
}

func getContextID(ctx context.Context) string {
	return middleware.GetReqID(ctx)
}
