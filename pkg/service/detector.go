/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package service

import (
	"context"

	"github.com/fidelity/theliv/internal/problem"
)

func DetectAlerts(ctx context.Context) ([]problem.NewProblem, error) {
	// input := GetDetectorInput(ctx)

	// pmgr := problem.DefaultProblemMgr()
	// Register detectors
	// problemdetectors.Register(pmgr.DetectorRegistrationFunc())
	// pbe, err := problem.NewDefaultProblemGraph(pmgr.Domains(), input)
	// if err != nil {
	// 	//TODO log
	// 	return nil, err
	// }
	// r, err := pbe.Execute(ctx)
	// if err != nil {
	// 	return nil, err
	// }

	// problems := make([]problem.Problem, 0)
	// for _, val := range r.ProblemMap {
	// 	problems = append(problems, val...)
	// }

	// Aggregator
	// client, err := kubeclient.NewKubeClient(input.Kubeconfig)
	// if err != nil {
	// 	return nil, err
	// }
	// return problem.Aggregate(ctx, problems, client)
	return nil, nil
}
