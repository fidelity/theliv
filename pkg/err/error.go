package err

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"runtime"

	log "github.com/fidelity/theliv/pkg/log"
)

func ErrorHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil && err != http.ErrAbortHandler {
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
		}()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

type Error struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func Panic(kind ErrorType, message string) {
	Panicf(kind, message)
}

func Panicf(kind ErrorType, message string, args ...interface{}) {
	if !reflect.ValueOf(args).IsZero() {
		message = fmt.Sprintf(message, args...)
	}

	error := Error{
		Type:    kind.String(),
		Message: message,
	}

	log.S().Error(kind.String(), ": ", message)

	e, _ := json.Marshal(error)
	panic(string(e))
}

type ErrorType int32

const (
	COMMON ErrorType = iota
	CONFIGURATION
	ETCD
	DATADOG
	KUBERNETES
	AWS
)

func (s ErrorType) String() string {
	switch s {
	case COMMON:
		return "COMMON"
	case CONFIGURATION:
		return "CONFIGURATION"
	case ETCD:
		return "ETCD"
	case DATADOG:
		return "DATA_DOG"
	case KUBERNETES:
		return "KUBERNETES"
	case AWS:
		return "AWS"
	default:
		return "UNKNOWN"
	}
}
