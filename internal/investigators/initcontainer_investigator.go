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

func InitContainerImagePullBackoffInvestigator(ctx context.Context, problem *problem.Problem,
	input *problem.DetectorCreationInput) {
	// Load kubernetes resource details
	pod := *problem.AffectedResources.Resource.(*v1.Pod)

	for _, status := range pod.Status.InitContainerStatuses {
		investigateContainerImgPullBackOff(ctx, problem, input, pod, status)
	}
}
