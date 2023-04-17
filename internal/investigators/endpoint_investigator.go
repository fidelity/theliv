/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package investigators

import (
	"context"

	"github.com/fidelity/theliv/internal/problem"
	com "github.com/fidelity/theliv/pkg/common"
	"github.com/fidelity/theliv/pkg/kubeclient"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	NotReadyAddressSolution = `
There's NotReadyAddresses in the subsets.
Please check the health of the selected pods.
`
	NoPoSelectedSolution = `
There's no Pod exists that match the service selector.
Please set the service selector or add endpoint subsets properly.
`
	NoServiceFoundSolution = `
No Service found for Endpoint {{ .Name}}. And this Endpoint has No subsets.
Please check if this Endpoint is necessary.
Cmd: kubectl get endpoints {{ .Name}} -n {{ .ObjectMeta.Namespace }}
`
)

func EndpointAddressNotAvailableInvestigator(ctx context.Context,
	problem *problem.Problem, input *problem.DetectorCreationInput) {

	var solutions []string

	endpoint := *problem.AffectedResources.Resource.(*v1.Endpoints)
	svc := &v1.Service{}
	namespace := kubeclient.NamespacedName{
		Namespace: endpoint.Namespace,
		Name:      endpoint.Name,
	}
	if input.KubeClient.Get(ctx, svc, namespace, metav1.GetOptions{}) == nil {
		logChecking(ctx, com.Service + com.Blank + svc.Name)
		problem.AffectedResources.ResourceKind = com.Service
		problem.AffectedResources.Resource = svc
	} else {
		logChecking(ctx, com.Endpoint + com.Blank + endpoint.Name)
		appendSolution(problem,
			GetSolutionsByTemplate(ctx, NoServiceFoundSolution, endpoint, true))
	}

	if len(endpoint.Subsets) != 0 {
		solutions = GetSolutionsByTemplate(ctx, NotReadyAddressSolution, svc, true)
	} else {
		solutions = GetSolutionsByTemplate(ctx, NoPoSelectedSolution, svc, true)
	}

	appendSolution(problem, solutions)

}
