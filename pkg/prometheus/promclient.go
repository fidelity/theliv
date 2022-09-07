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

	"github.com/fidelity/theliv/internal/problem"
	"github.com/fidelity/theliv/pkg/config"
	errors "github.com/fidelity/theliv/pkg/err"
	log "github.com/fidelity/theliv/pkg/log"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	promconfig "github.com/prometheus/common/config"
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

func GetAlerts(input *problem.DetectorCreationInput) (result v1.AlertsResult, err error) {

	result = v1.AlertsResult{}
	promcfg := config.GetThelivConfig().Prometheus
	address := input.Kubeconfig.Host + "/api/v1/namespaces/" + promcfg.Namespace + "/services/https:" + promcfg.Name + ":" + promcfg.Port + "/proxy"

	client, err := api.NewClient(api.Config{
		Address: address,
		RoundTripper: promconfig.NewAuthorizationCredentialsRoundTripper("Bearer",
			promconfig.Secret(input.Kubeconfig.BearerToken),
			TLSRoundTripper),
	})
	if err != nil {
		log.S().Errorf("Got error when creating Prometheus client, error is %s", err)
		return
	}

	v1api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, err = v1api.Alerts(ctx)
	if err != nil {
		err = errors.NewCommonError(6, err.Error())
		log.S().Errorf("Got error when getting Prometheus alerts, error is %s", err)
	}
	return
}
