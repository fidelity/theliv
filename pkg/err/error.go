/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package err

import (
	"context"
	"fmt"
	"net/http"
	"runtime"

	log "github.com/fidelity/theliv/pkg/log"
)

type ErrorType uint8

const (
	COMMON ErrorType = iota
	CONFIGURATION
	DATABASE
	LOGSERVER
	KUBERNETES
	CLOUD
	PROMETHEUS
)

// Customised Error, with Kind
type CommonError struct {
	Kind    ErrorType `json:"kind"`
	Message string    `json:"message"`
}

func (c CommonError) Error() string {
	return c.ErrorMsg()
}

// New CommonError function, will log error message and stacktrace.
func NewCommonError(ctx context.Context, kind ErrorType, msg string) error {
	err := CommonError{Kind: kind, Message: kind.String() + ": " + msg}
	log.SWithContext(ctx).Error(err.ErrorMsg())
	return err
}

func (c CommonError) ErrorMsg() string {
	return fmt.Sprintf("%s: %s", c.Kind.String(), c.Message)
}

func (s ErrorType) String() string {
	switch s {
	case COMMON:
		return "COMMON"
	case CONFIGURATION:
		return "CONFIGURATION"
	case DATABASE:
		return "DATABASE"
	case LOGSERVER:
		return "LOGSERVER"
	case KUBERNETES:
		return "KUBERNETES"
	case CLOUD:
		return "CLOUD"
	case PROMETHEUS:
		return "PROMETHEUS"
	default:
		return "UNKNOWN"
	}
}

// Used as middleware, to catch unhandled Panic
// Return error message, Http status 500.
func PanicHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				message := ""
				switch e := err.(type) {
				case string:
					message = e
				case runtime.Error:
					message = e.Error()
				case error:
					message = e.Error()
				}
				log.SWithContext(r.Context()).Error(message)
				w.Write([]byte(message))
			}
		}()
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// If CommonError and Kind is DB/Cloud/Log/K8s, return 503.
// Else return 500.
func GetStatusCode(err error) int {
	switch e := err.(type) {
	case CommonError:
		if e.Kind > 1 {
			return http.StatusServiceUnavailable
		} else {
			return http.StatusInternalServerError
		}
	default:
		return http.StatusInternalServerError
	}
}
