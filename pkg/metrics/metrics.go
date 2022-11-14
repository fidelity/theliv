/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package metrics

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

var totalRequests = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Number of get requests.",
	},
	[]string{"method", "path"},
)

var responseStatus = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_response_status",
		Help: "Status of HTTP response",
	},
	[]string{"method", "path", "status"},
)

var httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "http_response_time_seconds",
	Help:    "Duration of HTTP requests.",
	Buckets: []float64{1, 5, 10, 30, 60},
}, []string{"method", "path"})

func PrometheusMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if strings.Contains(r.RequestURI, "/theliv-api/v1") {

			method := r.Method
			path := strings.Split(r.RequestURI, "/")[3]

			start := time.Now()

			rw := newResponseWriter(w)
			next.ServeHTTP(rw, r)

			statusCode := rw.statusCode
			duration := time.Since(start)

			httpDuration.WithLabelValues(method, path).Observe(duration.Seconds())
			responseStatus.WithLabelValues(method, path, strconv.Itoa(statusCode)).Inc()
			totalRequests.WithLabelValues(method, path).Inc()

		} else {
			next.ServeHTTP(w, r)
		}
	})
}

func init() {
	prometheus.Register(totalRequests)
	prometheus.Register(responseStatus)
	prometheus.Register(httpDuration)
}
