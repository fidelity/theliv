package router

import (
	"net/http"
	"reflect"

	"github.com/go-chi/render"
)

func processEmpty(w http.ResponseWriter, r *http.Request, content interface{}) bool {
	value := reflect.ValueOf(content)
	if !value.IsZero() {
		return false
	}

	switch value.Kind() {
	case reflect.Slice:
		render.Respond(w, r, []string{})
	default:
		render.Respond(w, r, struct{}{})
	}
	return true
}
