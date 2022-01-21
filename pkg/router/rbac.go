/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package router

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/fidelity/theliv/pkg/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

const (
	MISS_PATHS_IN_BODY  = "paths field missing or empty in request body"
	SERVICE_UNAVAILABLE = "DB service Unavailable"
	ROLE_UPDATED        = "Role update success"
)

type Paths struct {
	Paths []string `json:"paths"`
}

// Add POST /{roleName}/addpath operation
// PathParameter {roleName} should be a valid role name.
func Rbac(r chi.Router) {
	r.Route("/{roleName}/addpath", func(r chi.Router) {
		r.Post("/", addPath)
	})
}

// Request body should be in valid Json format, or 400 will be returned.
// Request body should include {"paths": [""]}, paths should be an unempty collection,
// or 400 will be returned.
// If backend DB operation failed, return 503.
// if backend DB operation success, return normal response.
func addPath(w http.ResponseWriter, r *http.Request) {
	roleName := chi.URLParam(r, "roleName")
	var paths Paths

	if err := decodeBody(r.Body, &paths); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(paths.Paths) == 0 {
		http.Error(w, MISS_PATHS_IN_BODY, http.StatusBadRequest)
		return
	}
	if err := service.AddPath(roleName, paths.Paths); err != nil {
		http.Error(w, SERVICE_UNAVAILABLE, http.StatusServiceUnavailable)
		return
	}
	render.Respond(w, r, ROLE_UPDATED)
}

// This function will decode the http.Request.Body
// Decoded body will be unmarshaled to the object specified by the second parameter.
// If error occurred, return error.
func decodeBody(body io.ReadCloser, result interface{}) error {
	return json.NewDecoder(body).Decode(&result)
}
