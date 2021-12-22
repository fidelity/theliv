package router

import (
	"net/http"

	"github.com/fidelity/theliv/pkg/auth/authmiddleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func Userinfo(r chi.Router) {
	r.Get("/", func(w http.ResponseWriter, req *http.Request) {
		user, err := authmiddleware.GetUser(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		if empty := processEmpty(w, req, user); !empty {
			render.Respond(w, req, user)
		}
	})
}
