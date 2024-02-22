/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package router

import (
	"net/http"

	"github.com/fidelity/theliv/pkg/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type Prompt struct {
	Prompt string `json:"prompt,omitempty"`
}

func Ai(r chi.Router) {
	r.Post("/", completion)
}

func completion(w http.ResponseWriter, r *http.Request) {

	var prompt Prompt

	if err := decodeBody(r.Body, &prompt); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp, err := service.Completion(r.Context(), prompt.Prompt)
	if err != nil {
		processError(w, r, err)
	} else if empty := processEmpty(w, r, resp); !empty {
		render.Respond(w, r, resp)
	}
}
