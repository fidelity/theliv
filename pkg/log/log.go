/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package logging

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var DefaultLogger *zap.Logger

// Build a *zap.Logger from zap.Config
func getLogger(config zap.Config) *zap.Logger {
	loggerMgr, err := config.Build()
	if err != nil {
		panic("Error building Zap logger")
	}
	return loggerMgr
}

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
	return L().Sugar()
}

// Return Logger instance with request id from context
func LWithContext(ctx context.Context) *zap.Logger {
	l := L()
	if reqID := middleware.GetReqID(ctx); reqID != "" {
		l = DefaultLogger.With(zap.String("RequestId", reqID))
	}
	return l
}

// Return Logger.Sugar instance with request id from context
func SWithContext(ctx context.Context) *zap.SugaredLogger {
	return LWithContext(ctx).Sugar()
}

// Used as middleware, to insert request id into the response header if present
func RequestIDHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if reqID := middleware.GetReqID(r.Context()); reqID != "" {
			w.Header().Set("X-Request-Id", reqID)
		}
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
