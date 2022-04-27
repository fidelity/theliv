/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package prometheus

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"time"

	golog "log"

	"github.com/fidelity/theliv/internal/problem"
	"github.com/fidelity/theliv/pkg/service"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	promconfig "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
)

var TLSRoundTripper http.RoundTripper = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext,
	TLSHandshakeTimeout: 10 * time.Second,
	TLSClientConfig: (&tls.Config{
		InsecureSkipVerify: true,
	}),
}

func DetectAlerts(ctx context.Context) ([]problem.NewProblem, error) {
	input := service.GetDetectorInput(ctx)
	client, err := api.NewClient(api.Config{
		Address: "https://tochange.prometheus.host:8443",
		RoundTripper: promconfig.NewAuthorizationCredentialsRoundTripper("Bearer",
			promconfig.Secret(input.Kubeconfig.BearerToken),
			TLSRoundTripper),
	})
	if err != nil {
		golog.Printf("ERROR - Got error when creating Prometheus client, error is %s", err)
	}

	v1api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, err := v1api.Alerts(ctx)
	if err != nil {
		golog.Printf("ERROR - Got error when getting Prometheus alerts, error is %s", err)
	}

	// TODO: filter by namespace

	// build problems from  alerts, problem is detector input
	problems := make([]problem.NewProblem, 0)
	for _, alert := range result.Alerts {
		p := problem.NewProblem{}
		p.Name = string(alert.Labels[model.LabelName("alertname")])
		p.Description = string(alert.Annotations[model.LabelName("description")])
		p.Tags = make(map[string]string)
		for ln, lv := range alert.Labels {
			p.Tags[string(ln)] = string(lv)
		}
		problems = append(problems, p)
	}
	return problems, err
}