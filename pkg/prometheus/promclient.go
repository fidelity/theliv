/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package prometheus

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/fidelity/theliv/internal/problem"
	"github.com/fidelity/theliv/pkg/config"
	errors "github.com/fidelity/theliv/pkg/err"
	log "github.com/fidelity/theliv/pkg/log"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	promconfig "github.com/prometheus/common/config"
	"k8s.io/client-go/rest"
)

func loadDataFromFileOrInline(inline []byte, filePath string) ([]byte, error) {
	if len(inline) > 0 {
		return inline, nil
	}
	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
		}
		return data, nil
	}
	return nil, nil
}

func buildTLSRoundTripper(kubeconfig *rest.Config) (http.RoundTripper, error) {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	caData, err := loadDataFromFileOrInline(
		kubeconfig.CAData, kubeconfig.CAFile)
	if err != nil {
		return nil, err
	}
	if len(caData) > 0 {
		caPool := x509.NewCertPool()
		if !caPool.AppendCertsFromPEM(caData) {
			return nil, fmt.Errorf("failed to parse CA certificate from kubeconfig")
		}
		tlsConfig.RootCAs = caPool
	}

	certData, err := loadDataFromFileOrInline(
		kubeconfig.CertData, kubeconfig.CertFile)
	if err != nil {
		return nil, err
	}
	keyData, err := loadDataFromFileOrInline(
		kubeconfig.KeyData, kubeconfig.KeyFile)
	if err != nil {
		return nil, err
	}
	if len(certData) > 0 && len(keyData) > 0 {
		cert, err := tls.X509KeyPair(certData, keyData)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig:     tlsConfig,
	}, nil
}

func GetAlerts(ctx context.Context, input *problem.DetectorCreationInput) (v1.AlertsResult, error) {
	promcfg := config.GetThelivConfig().Prometheus
	address := input.Kubeconfig.Host + "/api/v1/namespaces/" +
		promcfg.Namespace + "/services/https:" + promcfg.Name + ":" + promcfg.Port + "/proxy"

	transport, err := buildTLSRoundTripper(input.Kubeconfig)
	if err != nil {
		log.SWithContext(ctx).Errorf("Failed to build TLS transport from kubeconfig: %s", err)
		return v1.AlertsResult{}, err
	}

	client, err := api.NewClient(api.Config{
		Address: address,
		RoundTripper: promconfig.NewAuthorizationCredentialsRoundTripper("Bearer",
			promconfig.NewInlineSecret(input.Kubeconfig.BearerToken),
			transport),
	})
	if err != nil {
		log.SWithContext(ctx).Errorf("Got error when creating Prometheus client, error is %s", err)
		return v1.AlertsResult{}, err
	}

	v1api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, err := v1api.Alerts(ctx)
	if err != nil {
		err = errors.NewCommonError(ctx, 6, err.Error())
		log.SWithContext(ctx).Errorf("Got error when getting Prometheus alerts, error is %s", err)
		return v1.AlertsResult{}, err
	}
	return result, nil
}
