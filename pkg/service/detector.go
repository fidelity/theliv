/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package service

import (
	"context"

	"github.com/fidelity/theliv/internal/investigators"
	"github.com/fidelity/theliv/internal/problem"
	"github.com/fidelity/theliv/internal/problemdetectors"
	"github.com/fidelity/theliv/pkg/kubeclient"
	"github.com/fidelity/theliv/pkg/prometheus"
	"github.com/prometheus/common/model"
)

func Detect(ctx context.Context) (interface{}, error) {
	input := GetDetectorInput(ctx)

	pmgr := problem.DefaultProblemMgr()
	// Register detectors
	problemdetectors.Register(pmgr.DetectorRegistrationFunc())
	pbe, err := problem.NewDefaultProblemGraph(pmgr.Domains(), input)
	if err != nil {
		//TODO log
		return nil, err
	}
	r, err := pbe.Execute(ctx)
	if err != nil {
		return nil, err
	}

	problems := make([]problem.Problem, 0)
	for _, val := range r.ProblemMap {
		problems = append(problems, val...)
	}

	// Aggregator
	client, err := kubeclient.NewKubeClient(input.Kubeconfig)
	if err != nil {
		return nil, err
	}
	return problem.Aggregate(ctx, problems, client)
}

func DetectAlerts(ctx context.Context) ([]*problem.NewProblem, error) {
	input := GetDetectorInput(ctx)
	result, _ := prometheus.GetAlerts(input)

	// TODO: filter by namespace

	// build problems from  alerts, problem is detector input
	problems := make([]*problem.NewProblem, 0)
	for _, alert := range result.Alerts {
		p := problem.NewProblem{}
		p.Name = string(alert.Labels[model.LabelName("alertname")])
		p.Description = string(alert.Annotations[model.LabelName("description")])
		p.Tags = make(map[string]string)
		for ln, lv := range alert.Labels {
			p.Tags[string(ln)] = string(lv)
		}
		problems = append(problems, &p)
	}

	// TODO: register investigator func map
	// TODO: check investigator func or default
	for _, problem := range problems {
		investigators.CommonInvestigator(ctx, problem, input)
	}
	return problems, nil
}
