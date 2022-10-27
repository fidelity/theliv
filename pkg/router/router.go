package router

import (
	"github.com/go-chi/chi/v5"
)

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

}
