/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package router

import (
	"net/http"

	"github.com/fidelity/theliv/pkg/auth/authmiddleware"
	com "github.com/fidelity/theliv/pkg/common"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func Userinfo(r chi.Router) {
	r.Get("/", func(w http.ResponseWriter, req *http.Request) {
		user, err := authmiddleware.GetUser(req, false)
		if err != nil {
			http.Error(w, com.NoUserInfo, http.StatusInternalServerError)
		} else if empty := processEmpty(w, req, user); !empty {
			render.Respond(w, req, user)
		}
	})
}
