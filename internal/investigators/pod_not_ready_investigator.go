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
	"github.com/fidelity/theliv/pkg/observability"
	v1 "k8s.io/api/core/v1"
)

const (
	ReadinessProbeFailedSolution = `1. Pod has status 'ContainersNotReady'.
2. It appears there is an issue with your readiness probe.
3. Events:{{if not .}} no events found.{{else}}{{range .}}
- {{.}}.{{end}}{{end}}
4. For more details, please refer to the Events section. Click on the pod name above to see the Pod configuration.
`

	ReadinessGateFailedSolution = `1. Pod has status 'ReadinessGatesNotReady'.
2. It appears there is an issue with your readiness gates.
3. Message: {{if not .}}no message found{{else}}{{.}}.{{end}}
4. Please check the readiness gate configurations for your pod. Click on the pod name above to see the Pod configurations.
`

	UsefulCommands = `
1. kubectl get pod {{.Name}} -n {{.ObjectMeta.Namespace}} -o yaml
2. kubectl get events --field-selector involvedObject.name={{.Name}} -n {{.ObjectMeta.Namespace}}`
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
				events, err := GetResourceEvents(ctx, input, pod.Name, pod.Namespace)
				if err != nil {
					log.SWithContext(ctx).Error("Got error when calling Kubernetes event API, error is %s", err)
				}

				var msg []string
				if con.Message != "" {
					msg = []string{con.Message}
				} else {
					msg = []string{}
				}
				if len(events) > 0 {
					getMsgFromEvents(ctx, events, pod, &msg)
				}
				solution = GetSolutionsByTemplate(ctx, ReadinessProbeFailedSolution, msg, true)
			}
			appendSolution(problem, solution, GetSolutionsByTemplate(ctx, UsefulCommands, pod, true))
		}
	}
}

func getMsgFromEvents(ctx context.Context, events []observability.EventRecord, pod v1.Pod, msg *[]string) {
	for _, event := range events {
		for _, eventMsg := range PodNotReadyEventMessage {
			matched, err := regexp.MatchString(strings.ToLower(eventMsg), strings.ToLower(event.Message))
			if err == nil && matched {
				log.SWithContext(ctx).Infof("Found event with error '%s', pod %s", eventMsg, pod.Name)
				*msg = append(*msg, event.Message)
			}
		}
	}
}
