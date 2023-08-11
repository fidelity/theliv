/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package investigators

import (
	"context"
	"sync"

	"github.com/fidelity/theliv/internal/problem"
	"github.com/fidelity/theliv/pkg/eval"
	v1 "k8s.io/api/core/v1"
)

func InitContainerImagePullBackoffInvestigator(ctx context.Context, wg *sync.WaitGroup, problem *problem.Problem,
	input *problem.DetectorCreationInput) {
	defer eval.Timer("investigators - InitContainerImagePullBackoffInvestigator")()
	defer wg.Done()
	// Load kubernetes resource details
	pod := *problem.AffectedResources.Resource.(*v1.Pod)

	for _, status := range pod.Status.InitContainerStatuses {
		investigateContainerImgPullBackOff(ctx, problem, input, pod, status)
	}
}
