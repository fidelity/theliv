package err

import (
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
func NewCommonError(kind ErrorType, msg string) error {
	err := CommonError{Kind: kind, Message: msg}
	log.S().Error(err.ErrorMsg())
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
	default:
		return "UNKNOWN"
	}
}

// Used in the controller, to catch unhandled Panic
// Return error message, Http status 500.
func PanicHandler(w http.ResponseWriter) {
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
		w.Write([]byte(message))
	}
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
