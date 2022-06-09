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
	// r.Get("/{cluster}/{namespace}/detect", detect)
	r.Get("/{cluster}/{namespace}/prometheus", detectPrometheusAlerts)
}

func detectPrometheusAlerts(w http.ResponseWriter, r *http.Request) {
	con, err := service.DetectAlerts(createDetectorInputWithContext(r))
	if err != nil {
		processError(w, r, err)
	} else {
		render.JSON(w, r, con)
	}
}

// func detect(w http.ResponseWriter, r *http.Request) {
// 	con, err := service.Detect(createDetectorInputWithContext(r))
// 	if err != nil {
// 		processError(w, r, err)
// 	} else {
// 		render.JSON(w, r, con)
// 	}
// }

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
	// thelivcfg := config.GetThelivConfig()

	// The kubeclient will be used for k8s logs/events driver.
	// client, err := kubeclient.NewKubeClient(k8sconfig)
	// if err != nil {
	// 	golog.Printf("ERROR - Got error when getting deployment client with kubeclient, error is %s", err)
	// }

	// eventClient, logClient := getLogDriver(thelivcfg, client)
	// eventDeeplinkClient, logDeeplinkClient := getDeeplinkDrivers(thelivcfg)

	input := &problem.DetectorCreationInput{
		Kubeconfig:  k8sconfig,
		ClusterName: cluster,
		Namespace:   namespace,
		// EventRetriever:         eventClient,
		// LogRetriever:           logClient,
		// EventDeeplinkRetriever: eventDeeplinkClient,
		// LogDeeplinkRetriever:   logDeeplinkClient,
		AwsConfig: ac,
	}

	return service.SetDetectorInput(ctx, input)
}
