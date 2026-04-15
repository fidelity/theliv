/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package router

import (
	"log/slog"
	"net/http"

	"github.com/fidelity/theliv/pkg/err"
	log "github.com/fidelity/theliv/pkg/log"
	"github.com/fidelity/theliv/pkg/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

const (
	Registered             = "Registered"
	errClusterNameTooShort = "cluster name must be at least 3 characters long"
)

func Register(r chi.Router) {
	r.Post("/{cluster}", clusterRegister)
}

// Request body should be in valid Json format, or 400 will be returned.
// Path parameter cluster, should be a valid cluster name (minimum 3 characters).
// If cluster name is less than 3 characters, 400 will be returned.
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

	l := log.SWithContext(r.Context()).With(
		slog.String("cluster", basic.Name),
		slog.String("clusterUrl", basic.Url),
		slog.String("account", basic.Account),
	)

	// return 400 if provided cluster name length is less than 3
	if len(basic.Name) < 3 {
		l.With(slog.Int("length", len(basic.Name))).Error("cluster name is too short")
		processError(w, r, err.NewCommonError(r.Context(), err.API, errClusterNameTooShort))
		return
	}

	// Validate critical fields but only log errors to
	// prevent k8s jobs from retrying failed registrations
	if basic.Url == "" {
		l.Error("cluster URL is empty")
	}

	if basic.CA == "" {
		l.Error("cluster CA is empty")
	}

	if basic.Token == "" {
		l.Error("cluster token is empty")
	}

	if err := service.RegisterCluster(r.Context(), l, basic); err != nil {
		http.Error(w, SERVICE_UNAVAILABLE, http.StatusServiceUnavailable)
		return
	}
	render.Respond(w, r, Registered)
}
