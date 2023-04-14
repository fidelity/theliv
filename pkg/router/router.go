/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package router

import (
	"github.com/fidelity/theliv/pkg/auth/authmiddleware"
	"github.com/fidelity/theliv/pkg/auth/samlmethod"
	"github.com/fidelity/theliv/pkg/err"
	"github.com/fidelity/theliv/pkg/metrics"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewRouter() *chi.Mux {

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)

	//set content type as json by default
	r.Use(middleware.SetHeader("Content-Type", "application/json"))

	// Authentication
	r.Use(authmiddleware.StartAuth)

	// use prometheus middleware
	r.Use(metrics.PrometheusMiddleware)

	// Add panic handling middleware
	r.Use(err.PanicHandler)

	// api route
	r.Route("/theliv-api/v1", Route)

	// saml route
	r.Handle("/auth/saml/*", samlmethod.GetSP())

	return r

}

// Router for /thliev-api/v1
func Route(r chi.Router) {
	// health check
	r.Route("/health", HealthCheck)

	// list cluster and namespaces
	r.Route("/clusters", Cluster)

	// detector
	r.Route("/detector", Detector)

	// userinfo
	r.Route("/userinfo", Userinfo)

	// feedback
	r.Route("/feedbacks", SubmitFeedback)

	// rbac
	r.Route("/rbac", Rbac)

	// add role for app team
	r.Route("/adgroup", AdGroup)

	// new cluster registration
	r.Route("/register", Register)

	// config for UI
	r.Route("/configinfo", ConfigInfo)

	// export prometheus metrics
	r.Handle("/metrics", promhttp.Handler())

}
