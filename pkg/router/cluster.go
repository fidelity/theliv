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

func Cluster(r chi.Router) {
	r.Get("/", listClusters)
	r.Route("/{clusterName}/namespaces", func(r chi.Router) {
		r.Get("/", listNamespaces)
	})
}

func listClusters(w http.ResponseWriter, r *http.Request) {
	clusters, err := service.GetClusters(r.Context())
	if err != nil {
		processError(w, r, err)
	} else if empty := processEmpty(w, r, clusters); !empty {
		render.Respond(w, r, clusters)
	}
}

func listNamespaces(w http.ResponseWriter, r *http.Request) {
	clusterName := chi.URLParam(r, "clusterName")
	names, err := service.ListNs(r.Context(), clusterName)

	if err != nil {
		processError(w, r, err)
	} else if empty := processEmpty(w, r, names); !empty {
		render.Respond(w, r, names)
	}
}
