/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package investigators

import (
	"context"
	"regexp"
	"strings"
	"sync"

	"github.com/fidelity/theliv/internal/problem"
	log "github.com/fidelity/theliv/pkg/log"
	v1 "k8s.io/api/core/v1"
)

const (
	ReadinessProbeFailedSolution = `1. Why the pod is not ready is because of the reason: ContainersNotReady.
2. It appears that your readiness probe is having trouble in this instance.
3. {{.}}.
4. For more details, you can refer to the Events section. You can click on the pod name above to see the Pod configurations.
`

	ReadinessGateFailedSolution = `1. Why the pod is not ready is because of the reason: ReadinessGatesNotReady.
2. It appears that your readiness Gates are having trouble in this instance.
3. Message: {{.}}.
4. Please check the readiness gate configurations for your pod. You can click on the pod name above to see the Pod configurations.
`

	UsefulCommands = `
1. kubectl get pod {{.Name}} -n {{.ObjectMeta.Namespace}} -o yaml
2. kubectl get events -n {{.ObjectMeta.Namespace}}`
)

type PodNotReady struct {
	Pod     *v1.Pod
	Reason  string
	Message string
	Config  string
}

var PodNotReadyEventMessage = []string{
	"Readiness probe failed",
	// "Back-off restarting failed container",
}

func PodNotReadyInvestigator(ctx context.Context, wg *sync.WaitGroup, problem *problem.Problem, input *problem.DetectorCreationInput) {
	defer wg.Done()

	pod := *problem.AffectedResources.Resource.(*v1.Pod)
	for _, con := range pod.Status.Conditions {
		if con.Type == "Ready" {
			var solution []string
			if con.Reason == "ReadinessGatesNotReady" {
				solution = GetSolutionsByTemplate(ctx, ReadinessGateFailedSolution, con.Message, true)
			} else {
				events, err := GetPodEvents(ctx, input, &pod)
				if err != nil {
					log.SWithContext(ctx).Error("Got error when calling Kubernetes event API, error is %s", err)
				}

				msg := "Message: " + con.Message
				if len(events) > 0 {
					for _, event := range events {
						for _, eventMsg := range PodNotReadyEventMessage {
							matched, err := regexp.MatchString(strings.ToLower(eventMsg), strings.ToLower(event.Message))
							if err == nil && matched {
								log.SWithContext(ctx).Infof("Found event with error '%s', pod %s", eventMsg, pod.Name)
								msg = "Event: " + event.Message
							}
						}
					}
				}
				solution = GetSolutionsByTemplate(ctx, ReadinessProbeFailedSolution, msg, true)
			}
			appendSolution(problem, solution, GetSolutionsByTemplate(ctx, UsefulCommands, pod, true))
		}
	}
}
