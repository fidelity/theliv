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

// LogKind stores the type of the value to log
type LogKind byte

const (
	// Str is an alias for string type values
	Str LogKind = iota
)

// LogField is value used as a context key
type LogField struct {
	// Label is the string we use as key in the log entry
	Label string

	// Kind is the type of the value
	Kind LogKind
}

func (f LogField) ToZapField(value string) zap.Field {
	return zap.String(f.Label, value)
}

var (
	// ReqestID is the context key for Request IDs
	RequestID LogField = LogField{
		Label: "RequestID",
		Kind:  Str,
	}
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
func L(opts ...LoggerOption) *zap.Logger {
	if DefaultLogger == nil {
		NewDefaultLogger(DefaultLogConfig(0))
	}

	// Apply all the options
	// Create copy of default logger to avoid 
	// overwriting it with its child created in opt()
	l := DefaultLogger
	for _, opt := range opts {
		l = opt(DefaultLogger)
	}

	return l
}

// Return default Logger.Sugar instance.
func S(opts ...LoggerOption) *zap.SugaredLogger {
	if DefaultLogger == nil {
		NewDefaultLogger(DefaultLogConfig(0))
	}

	// Apply all the options
	// Create copy of default logger to avoid 
	// overwriting it with its child created in opt()
	l := DefaultLogger
	for _, opt := range opts {
		l = opt(DefaultLogger)
	}

	return l.Sugar()
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

type LoggerOption func(*zap.Logger) *zap.Logger

func WithReqId(ctx context.Context) LoggerOption {
	return func(l *zap.Logger) *zap.Logger {
		return l.With(RequestID.ToZapField(middleware.GetReqID(ctx)))
	}
}