/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package investigators

import (
	"context"
	"strings"
	"sync"

	"github.com/fidelity/theliv/internal/problem"
	com "github.com/fidelity/theliv/pkg/common"
	v1 "k8s.io/api/apps/v1"
)

const (
	NotAvailableSolution = `
1. Deployment '{{.Name}}' is not available.
2. Please check the replica(s) status in this deployment.
`
	MemoryQuotaSolution = `
1. Deployment '{{.Name}}' has exceeded memory quotas.
2. Please check the memory requests/limits of your deployment.
`
	CPUQuotaSolution = `
1. Deployment '{{.Name}}' has exceeded CPU quotas.
2. Please check the CPU requests/limits of your deployment.
`
	ResourceQuotaSolution = `
1. Deployment '{{.Name}}' has exceeded resource quotas.
2. Please check the requests/limits of your deployment.
`
	DescribeCmd = `
kubectl describe deploy {{.Name}} -n {{ .ObjectMeta.Namespace }}
`
)

func DeploymentNotAvailableInvestigator(ctx context.Context, wg *sync.WaitGroup, problem *problem.Problem,
	input *problem.DetectorCreationInput) {
	defer wg.Done()

	getDeployCommonSolution(ctx, problem)
}

func DeploymentGenerationMismatchInvestigator(ctx context.Context, wg *sync.WaitGroup, problem *problem.Problem,
	input *problem.DetectorCreationInput) {
	defer wg.Done()

	getDeployCommonSolution(ctx, problem)
}

func DeploymentReplicasMismatchInvestigator(ctx context.Context, wg *sync.WaitGroup, problem *problem.Problem,
	input *problem.DetectorCreationInput) {
	defer wg.Done()

	getDeployCommonSolution(ctx, problem)
}

func getDeployCommonSolution(ctx context.Context, problem *problem.Problem) {
	deploy := *problem.AffectedResources.Resource.(*v1.Deployment)
	logChecking(ctx, com.Deployment+com.Blank+deploy.Name)
	appendSolution(problem, getDeploySolution(ctx, deploy),
		GetSolutionsByTemplate(ctx, DescribeCmd, deploy, true))
}

func getDeploySolution(ctx context.Context, deploy v1.Deployment) []string {
	if ok, cd := containsCdt(deploy.Status.Conditions, "ReplicaFailure"); ok {
		if cd.Status == "True" {
			msg := strings.ToLower(cd.Message)
			switch {
			case strings.Contains(msg, "cpu"):
				return GetSolutionsByTemplate(ctx, CPUQuotaSolution, deploy, true)
			case strings.Contains(msg, "memory"):
				return GetSolutionsByTemplate(ctx, MemoryQuotaSolution, deploy, true)
			case strings.Contains(msg, "exceeded quota"):
				return GetSolutionsByTemplate(ctx, ResourceQuotaSolution, deploy, true)
			}
		}
	} else if ok, cd := containsCdt(deploy.Status.Conditions, "Available"); ok {
		if cd.Status == "False" {
			return GetSolutionsByTemplate(ctx, NotAvailableSolution, deploy, true)
		}
	}
	return GetSolutionsByTemplate(ctx, NotAvailableSolution, deploy, true)
}

func containsCdt(conditions []v1.DeploymentCondition, cType string) (bool, *v1.DeploymentCondition) {
	for _, c := range conditions {
		if string(c.Type) == cType {
			return true, &c
		}
	}
	return false, nil
}
