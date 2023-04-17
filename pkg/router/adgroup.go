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

const (
	MissingAdGroup = "adgroups field missing or empty in request body"
	Added          = "Cluster/NS added to adgroups"
	REMOVED        = "Cluster/NS removed from adgroups"
)

type AdGroups struct {
	AdGroups []string `json:"adgroups"`
}

func AdGroup(r chi.Router) {
	// POST /{cluster}/{namespace} operation
	r.Post("/{cluster}/{namespace}", addGroup)
	// DELETE /{cluster}/{namespace} operation
	r.Delete("/{cluster}/{namespace}", removeGroup)
}

// Request body should be in valid Json format, or 400 will be returned.
// Request body should include {"adgroups": [""]}, adgroups should be an non-empty collection,
// or 400 will be returned.
// If backend DB operation failed, return 503.
// if backend DB operation success, return normal response.
func addGroup(w http.ResponseWriter, r *http.Request) {
	cluster := chi.URLParam(r, "cluster")
	namespace := chi.URLParam(r, "namespace")

	var adgroups AdGroups

	if err := decodeBody(r.Body, &adgroups); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(adgroups.AdGroups) == 0 {
		http.Error(w, MissingAdGroup, http.StatusBadRequest)
		return
	}
	if err := service.AddGroup(r.Context(), cluster, namespace, adgroups.AdGroups); err != nil {
		http.Error(w, SERVICE_UNAVAILABLE, http.StatusServiceUnavailable)
		return
	}
	render.Respond(w, r, Added)
}

// Request body should be in valid Json format, or 400 will be returned.
// Request body should include {"adgroups": [""]}, adgroups should be an non-empty collection,
// or 400 will be returned.
// If backend DB operation failed, return 503.
// if backend DB operation success, return normal response.
func removeGroup(w http.ResponseWriter, r *http.Request) {
	cluster := chi.URLParam(r, "cluster")
	namespace := chi.URLParam(r, "namespace")

	var ads AdGroups

	if err := decodeBody(r.Body, &ads); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(ads.AdGroups) == 0 {
		http.Error(w, MissingAdGroup, http.StatusBadRequest)
		return
	}
	if err := service.RemoveGroup(r.Context(), cluster, namespace, ads.AdGroups); err != nil {
		http.Error(w, SERVICE_UNAVAILABLE, http.StatusServiceUnavailable)
		return
	}
	render.Respond(w, r, REMOVED)
}
