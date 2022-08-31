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
	com "github.com/fidelity/theliv/pkg/common"
	v1 "k8s.io/api/apps/v1"
)

const (
	NotAvailableSolution = `
1. Deployment {{.Name}} is not available.
2. Please check the replica(s) status in this deployment.
3. Cmd: kubectl describe deploy {{.Name}} -n {{ .ObjectMeta.Namespace }}
`
	ResourceQuotaSolution = `
1. Deployment {{.Name}} has Insufficient quota.
2. Please check the requests/limits of your deployment.
3. Cmd: kubectl describe deploy {{.Name}} -n {{ .ObjectMeta.Namespace }}
`
)

func DeploymentNotAvailableInvestigator(ctx context.Context, problem *problem.Problem,
	input *problem.DetectorCreationInput) {
	getDeployCommonSolution(problem)
}

func DeploymentGenerationMismatchInvestigator(ctx context.Context, problem *problem.Problem,
	input *problem.DetectorCreationInput) {
	getDeployCommonSolution(problem)
}

func DeploymentReplicasMismatchInvestigator(ctx context.Context, problem *problem.Problem,
	input *problem.DetectorCreationInput) {
	getDeployCommonSolution(problem)
}

func getDeployCommonSolution(problem *problem.Problem) {
	deploy := *problem.AffectedResources.Resource.(*v1.Deployment)
	logChecking(com.Deployment + com.Blank + deploy.Name)
	appendSolution(problem, getDeploySolution(deploy))
}

func getDeploySolution(deploy v1.Deployment) []string {

	if ok, cd := containsCdt(deploy.Status.Conditions, "ReplicaFailure"); ok {
		if cd.Status == "True" {
			if strings.Contains(strings.ToLower(cd.Message), "cpu") ||
				strings.Contains(strings.ToLower(cd.Message), "memory") ||
				strings.Contains(strings.ToLower(cd.Message), "exceeded quota") {
				return GetSolutionsByTemplate(ResourceQuotaSolution, deploy, true)
			}
		}
	} else if ok, cd := containsCdt(deploy.Status.Conditions, "Available"); ok {
		if cd.Status == "False" {
			return GetSolutionsByTemplate(NotAvailableSolution, deploy, true)
		}
	}
	return GetSolutionsByTemplate(NotAvailableSolution, deploy, true)
}

func containsCdt(conditions []v1.DeploymentCondition, cType string) (bool, *v1.DeploymentCondition) {
	for _, c := range conditions {
		if string(c.Type) == cType {
			return true, &c
		}
	}
	return false, nil
}
