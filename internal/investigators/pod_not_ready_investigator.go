/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package investigators

import (
	"context"

	"github.com/fidelity/theliv/internal/problem"
	v1 "k8s.io/api/core/v1"
)

const (
	PodNotReadySolution = `Pod {{.Pod.ObjectMeta.Name}} has been in running phase but not ready for more than 5 mins.
The Reason is: {{.Reason}}
The Message is: {{.Message}}
If the issue persists, please check the {{ .Config }} set for the Pod.
Cmd to check Pod: kubectl get pod {{ .Pod.ObjectMeta.Name }} -n {{ .Pod.ObjectMeta.Namespace }} -o yaml
`
)

type PodNotReady struct {
	Pod     *v1.Pod
	Reason  string
	Message string
	Config  string
}

func PodNotReadyInvestigator(ctx context.Context, problem *problem.Problem, input *problem.DetectorCreationInput) {

	pod := *problem.AffectedResources.Resource.(*v1.Pod)

	for _, con := range pod.Status.Conditions {
		if con.Type == "Ready" {
			config := "container readinessProbe"
			if con.Reason == "ReadinessGatesNotReady" {
				config = "ReadinessGates"
			}
			podInf := PodNotReady{
				Pod:     &pod,
				Message: con.Message,
				Reason:  con.Reason,
				Config:  config,
			}
			solution := GetSolutionsByTemplate(PodNotReadySolution, podInf, true)
			appendSolution(problem, solution)
		}
	}

}
