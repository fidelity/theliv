/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package investigators

import (
	"context"
	"strings"

	"github.com/fidelity/theliv/internal/problem"
	v1 "k8s.io/api/core/v1"
)

func InitContainerImagePullBackoffInvestigator(ctx context.Context, problem *problem.Problem, input *problem.DetectorCreationInput) {
	// Load kubenetes resource details
	pod := *problem.AffectedResources.Resource.(*v1.Pod)
	for _, ctnstatus := range pod.Status.InitContainerStatuses {
		detail := ctnstatus.State.Waiting.Message
		var solutions, secretmsg []string
		for _, ctn := range pod.Spec.InitContainers {
			if ctnstatus.Name == ctn.Name {
				for msg := range ImagePullBackoffSolutions {
					if strings.Contains(strings.ToLower(ctnstatus.State.Waiting.Message), strings.ToLower(msg)) {
						solutions = ImagePullBackoffSolutions[msg](&ctn, &msg)
						secretmsg = checksecretmsg(msg, pod)
					}
				}
				if len(solutions) == 0 {
					msg := UnknownManifestMsg
					solutions = ImagePullBackoffSolutions[UnknownManifestMsg](&ctn, &msg)
					secretmsg = checksecretmsg(msg, pod)
				}
			}
		}
		problem.SolutionDetails = append(problem.SolutionDetails, detail)
		problem.SolutionDetails = append(problem.SolutionDetails, solutions...)
		problem.SolutionDetails = append(problem.SolutionDetails, secretmsg...)
	}
}
