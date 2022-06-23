/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package router

import (
	"context"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/fidelity/theliv/internal/problem"
	"github.com/fidelity/theliv/pkg/config"
	"github.com/fidelity/theliv/pkg/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func Detector(r chi.Router) {
	r.Get("/{cluster}/{namespace}/detect", detectPrometheusAlerts)
}

func detectPrometheusAlerts(w http.ResponseWriter, r *http.Request) {
	con, err := service.DetectAlerts(createDetectorInputWithContext(r))
	if err != nil {
		processError(w, r, err)
	} else {
		render.JSON(w, r, con)
	}
}

func createDetectorInputWithContext(r *http.Request) context.Context {
	ctx := r.Context()
	// namespace
	cluster := chi.URLParam(r, "cluster")
	namespace := chi.URLParam(r, "namespace")

	// Get kubeconfig for the specified cluster
	conf := config.GetConfigLoader().GetKubernetesConfig(cluster)
	if conf == nil {
		return ctx
	}

	k8sconfig := conf.GetKubeConfig()
	awsconfig := conf.GetAwsConfig()
	var ac aws.Config
	if awsconfig != nil {
		ac = *awsconfig
	}

	input := &problem.DetectorCreationInput{
		Kubeconfig:  k8sconfig,
		ClusterName: cluster,
		Namespace:   namespace,
		AwsConfig:   ac,
	}

	return service.SetDetectorInput(ctx, input)
}
