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

const Registered = "Registered"

func Register(r chi.Router) {
	r.Post("/{cluster}", clusterRegister)
}

// Request body should be in valid Json format, or 400 will be returned.
// Path parameter cluster, should be a valid cluster name.
// Request body should include {"Url": "", "CA": "", "Token": ""}, or 400 will be returned.
// If backend DB operation failed, return 503.
// if backend DB operation success, return "Registered".
func clusterRegister(w http.ResponseWriter, r *http.Request) {

	cluster := chi.URLParam(r, "cluster")

	var basic service.ClusterBasic

	if err := decodeBody(r.Body, &basic); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	basic.Name = cluster
	if err := service.RegisterCluster(basic); err != nil {
		http.Error(w, SERVICE_UNAVAILABLE, http.StatusServiceUnavailable)
		return
	}
	render.Respond(w, r, Registered)
}
