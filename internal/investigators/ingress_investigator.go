/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package investigators

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/fidelity/theliv/internal/problem"
	networkv1 "k8s.io/api/networking/v1"
)

const (
	S3            = "s3"
	NoCertFound   = "no certificate found"
	CertNotFound  = "CertificateNotFound"
	Protocol      = "protocol"
	SecurityGroup = "securityGroups"

	Description        = "Ingress {{.Name}} has incorrect configuration detected by Ingress Controller."
	IngDefaultSolution = "1. Ingress configuration Error: %s"
	FixIngSolution     = "%d. Fix above issue then deploy again, could possibly make Ingress work."

	ProtocolSolution      = "%d. Check protocol config in Ingress annotations, correct protocol includes http, https."
	S3Solution            = "%d. Check s3.backet config in Ingress annotations, make sure S3 bucket exists and is accessible."
	SecurityGroupSolution = "%d. Check securityGroups config in Ingress annotations, pass the correct securityGroup associated with the cluster."
	CertNotFoundSolution  = "%d. Check certificate config in Ingress annotations, make sure the cert you passed exists with correct name and path. Or use some cert-manager to create one."

	IngressCommands = `
1. kubectl describe ing {{.Name}} -n {{.ObjectMeta.Namespace}}
2. kubectl get events --field-selector involvedObject.name={{.Name}} -n {{.ObjectMeta.Namespace}}`
)

var ingSolutions = map[string]string{
	S3:            S3Solution,
	NoCertFound:   CertNotFoundSolution,
	CertNotFound:  CertNotFoundSolution,
	Protocol:      ProtocolSolution,
	SecurityGroup: SecurityGroupSolution,
}

func IngressMisconfiguredInvestigator(ctx context.Context, wg *sync.WaitGroup,
	problem *problem.Problem, input *problem.DetectorCreationInput) {
	defer wg.Done()

	ing := *problem.AffectedResources.Resource.(*networkv1.Ingress)
	commands := GetSolutionsByTemplate(ctx, IngressCommands, ing, true)

	appendSolution(problem, IngressSolution(ctx, problem.Description, ing), commands)
	problem.Description = GetSolutionsByTemplate(ctx, Description, ing, true)[0]
}

func IngressSolution(ctx context.Context, reason string, ing networkv1.Ingress) []string {
	solutions := []string{fmt.Sprintf(IngDefaultSolution, reason)}

	for msg := range ingSolutions {
		matched, err := regexp.MatchString(strings.ToLower(msg), strings.ToLower(reason))
		if err == nil && matched {
			solutions = appendSeq(solutions, GetSolutionsByTemplate(ctx, ingSolutions[msg], ing, true)[0])
		}
	}
	solutions = appendSeq(solutions, FixIngSolution)

	return solutions

}
