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
	//  "github.com/go-chi/render"
 )
 
 func Sample(r chi.Router) {
	 r.Get("/", hello)
 }

 func hello(w http.ResponseWriter, req *http.Request) {
	service.Hello(req.Context())
}