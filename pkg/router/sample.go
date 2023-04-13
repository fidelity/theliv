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
)

func Sample(r chi.Router) {
	r.Get("/", hello)
	r.Get("/hello1", hello1)
}

func hello(w http.ResponseWriter, req *http.Request) {
	service.Hello(req.Context())
}

func hello1(w http.ResponseWriter, req *http.Request) {
	service.Hello1(req.Context())
}