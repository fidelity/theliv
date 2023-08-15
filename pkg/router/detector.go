/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package router

import (
	"context"
	"fmt"
	"net/http"

	"github.com/fidelity/theliv/internal/problem"
	com "github.com/fidelity/theliv/pkg/common"
	"github.com/fidelity/theliv/pkg/config"
	theErr "github.com/fidelity/theliv/pkg/err"
	"github.com/fidelity/theliv/pkg/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func Detector(r chi.Router) {
	r.Get("/{cluster}/{namespace}/detect", detectPrometheusAlerts)
	r.Get("/{cluster}/{namespace}/event", getK8sNsEvents)
}

func detectPrometheusAlerts(w http.ResponseWriter, r *http.Request) {

	ctx, err := createDetectorInputWithContext(r)
	if err != nil {
		processError(w, r, err)
	} else {
		con, err := service.DetectAlerts(ctx)
		if err != nil {
			processError(w, r, err)
		} else {
			render.JSON(w, r, con)
		}
	}
}

func getK8sNsEvents(w http.ResponseWriter, r *http.Request) {
	ctx, err := createDetectorInputWithContext(r)
	if err != nil {
		processError(w, r, err)
	} else {
		res, err := service.GetEvents(ctx)
		if err != nil {
			processError(w, r, err)
		} else {
			render.JSON(w, r, res)
		}
	}
}

func createDetectorInputWithContext(r *http.Request) (context.Context, error) {
	ctx := r.Context()
	// namespace
	cluster := chi.URLParam(r, "cluster")
	namespace := chi.URLParam(r, "namespace")

	contact := fmt.Sprintf(com.Contact, config.GetThelivConfig().TeamName)

	// Get kubeconfig for the specified cluster
	conf, err := config.GetConfigLoader().GetKubernetesConfig(r.Context(), cluster)
	if err != nil {
		return nil, theErr.NewCommonError(ctx, 4, com.LoadKubeConfigFailed+contact)

	}
	k8sconfig, err := conf.GetKubeConfig(r.Context())
	if err != nil {
		return nil, theErr.NewCommonError(ctx, 4, com.LoadKubeConfigFailed+contact)
	}
	// awsconfig := conf.GetAwsConfig()
	// var ac aws.Config
	// if awsconfig != nil {
	// 	ac = *awsconfig
	// }

	input := &problem.DetectorCreationInput{
		Kubeconfig:  k8sconfig,
		ClusterName: cluster,
		Namespace:   namespace,
		// AwsConfig:   ac,
	}

	return service.SetDetectorInput(ctx, input), nil
}
