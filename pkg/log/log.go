/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package logging

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

var DefaultLogger *slog.Logger

// SugaredSlogLogger wraps slog.Logger to provide Printf-style methods
type SugaredSlogLogger struct {
	logger *slog.Logger
}

// Infof logs an info message with Printf-style formatting
func (s *SugaredSlogLogger) Infof(format string, args ...interface{}) {
	s.logger.Info(fmt.Sprintf(format, args...))
}

// Info logs an info message without formatting
func (s *SugaredSlogLogger) Info(msg string) {
	s.logger.Info(msg)
}

// Errorf logs an error message with Printf-style formatting
func (s *SugaredSlogLogger) Errorf(format string, args ...interface{}) {
	s.logger.Error(fmt.Sprintf(format, args...))
}

// Error logs an error message without formatting
func (s *SugaredSlogLogger) Error(msg string) {
	s.logger.Error(msg)
}

// Warnf logs a warning message with Printf-style formatting
func (s *SugaredSlogLogger) Warnf(format string, args ...interface{}) {
	s.logger.Warn(fmt.Sprintf(format, args...))
}

// Warn logs a warning message without formatting
func (s *SugaredSlogLogger) Warn(msg string) {
	s.logger.Warn(msg)
}

// Debugf logs a debug message with Printf-style formatting
func (s *SugaredSlogLogger) Debugf(format string, args ...interface{}) {
	s.logger.Debug(fmt.Sprintf(format, args...))
}

// Debug logs a debug message without formatting
func (s *SugaredSlogLogger) Debug(msg string) {
	s.logger.Debug(msg)
}

// Fatalf logs an error message and exits with code 1
func (s *SugaredSlogLogger) Fatalf(format string, args ...interface{}) {
	s.logger.Error(fmt.Sprintf(format, args...))
	os.Exit(1)
}

// Panicf logs an error message and panics
func (s *SugaredSlogLogger) Panicf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	s.logger.Error(msg)
	panic(msg)
}

// With adds structured attributes to the logger
func (s *SugaredSlogLogger) With(attrs ...slog.Attr) *SugaredSlogLogger {
	args := make([]any, len(attrs))
	for i, attr := range attrs {
		args[i] = attr
	}
	return &SugaredSlogLogger{logger: s.logger.With(args...)}
}

// Return a new *slog.Logger instance, accept a *slog.HandlerOptions
func NewDefaultLogger(opts *slog.HandlerOptions) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)
	DefaultLogger = logger
	return logger
}

// default slog.HandlerOptions. accept level of int type, value between -1 and 5
func DefaultLogConfig(level int) *slog.HandlerOptions {
	var slogLevel slog.Level
	switch level {
	case -1:
		slogLevel = slog.LevelDebug
	case 0:
		slogLevel = slog.LevelInfo
	case 1:
		slogLevel = slog.LevelWarn
	case 2:
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	return &slog.HandlerOptions{
		Level: slogLevel,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Format timestamp as ISO8601 (RFC3339)
			if a.Key == slog.TimeKey {
				return slog.String(slog.TimeKey, a.Value.Time().Format(time.RFC3339))
			}
			return a
		},
	}
}

// Return default Logger instance.
func L() *slog.Logger {
	if DefaultLogger == nil {
		NewDefaultLogger(DefaultLogConfig(0))
	}

	return DefaultLogger
}

// Return default Logger.Sugar instance.
func S() *SugaredSlogLogger {
	return &SugaredSlogLogger{logger: L()}
}

// Return Logger instance with request id from context
func LWithContext(ctx context.Context) *slog.Logger {
	l := L()
	if reqID := middleware.GetReqID(ctx); reqID != "" {
		l = l.With(slog.String("RequestId", reqID))
	}
	return l
}

// Return Logger.Sugar instance with request id from context
func SWithContext(ctx context.Context) *SugaredSlogLogger {
	return &SugaredSlogLogger{logger: LWithContext(ctx)}
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
