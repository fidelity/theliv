/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package router

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

var startTime time.Time

func HealthCheck(r chi.Router) {
	r.Get("/", func(w http.ResponseWriter, req *http.Request) {
		status := getStatus()
		s, err := json.Marshal(status)
		if err != nil {
			//TODO log & error handling
			http.Error(w, err.Error(), 500)
			return
		}
		w.Write(s)
	})
}

type HealthStatus struct {
	Status string
	Since  string
}

func getStatus() *HealthStatus {
	return &HealthStatus{
		Status: "RUNNING",
		Since:  startTime.Local().String(),
	}
}

func init() {
	startTime = time.Now()
}
