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
	com "github.com/fidelity/theliv/pkg/common"
	"github.com/fidelity/theliv/pkg/kubeclient"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	NotReadyAddressSolution = `
1. At least one subset has 'NotReadyAddresses' status.
2. Please check the health of the selected pods.
`
	NoPodSelectedSolution = `
1. There are no pods that match the service selector(s):{{range $index, $value := .Spec.Selector}}
- {{$index}}: {{$value}}{{end}}
2. Please set the service selector or add endpoint subsets properly.
`
	NoServiceFoundSolution = `
1. No Service or subsets found for Endpoint {{.Name}}.
2. Please check if this Endpoint is necessary.
`
)

func EndpointAddressNotAvailableInvestigator(ctx context.Context, wg *sync.WaitGroup,
	problem *problem.Problem, input *problem.DetectorCreationInput) {
	defer wg.Done()

	var solutions []string

	endpoint := *problem.AffectedResources.Resource.(*v1.Endpoints)
	svc := &v1.Service{}
	namespace := kubeclient.NamespacedName{
		Namespace: endpoint.Namespace,
		Name:      endpoint.Name,
	}
	if input.KubeClient.Get(ctx, svc, namespace, metav1.GetOptions{}) == nil {
		logChecking(ctx, com.Service+com.Blank+svc.Name)
		problem.AffectedResources.ResourceKind = com.Service
		problem.AffectedResources.Resource = svc
	} else {
		logChecking(ctx, com.Endpoint+com.Blank+endpoint.Name)
		appendSolution(problem,
			GetSolutionsByTemplate(ctx, NoServiceFoundSolution, endpoint, true), nil)
	}

	if len(endpoint.Subsets) != 0 {
		solutions = GetSolutionsByTemplate(ctx, NotReadyAddressSolution, svc, true)
	} else {
		solutions = GetSolutionsByTemplate(ctx, NoPodSelectedSolution, svc, true)
	}

	appendSolution(problem, solutions, GetSolutionsByTemplate(ctx, GetEndpointsCmd, endpoint, true))
}
