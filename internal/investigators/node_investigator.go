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
	v1 "k8s.io/api/core/v1"
)

const (
	NotReadySolution = `
1. Node {{ .ObjectMeta.Name }} is not Ready.
2. NotReady message is: {{ range $index, $value := .Status.Conditions }}{{- if eq $value.Type "Ready" }}{{$value.Message}}{{- end}}{{end}}.
3. Please wait for sometime, if issue persists, you may need to ssh to the Node and check.
4. To restart Kubelet service, please refer to below UsefulCommands.
`

	MemPressSolution = `
1. Node {{ .ObjectMeta.Name }} is reporting memory pressure issue.
2. MemoryPressure message is: {{ range $index, $value := .Status.Conditions }}{{- if eq $value.Type "MemoryPressure" }}{{$value.Message}}{{- end}}{{end}}.
3. You may need to ssh to the Node, kill unnecessary processes.
4. You can find Pods running on this Node, re-allocate them to other Nodes.
5. To get pods running on the Node, refer to below command 1.
`

	DiskPressSolution = `
1. Node {{ .ObjectMeta.Name }} is reporting disk pressure issue.
2. DiskPressure message is: {{ range $index, $value := .Status.Conditions }}{{- if eq $value.Type "DiskPressure" }}{{$value.Message}}{{- end}}{{end}}.
3. You may need to ssh to the Node, delete unnecessary files, or add more storage resources.
`

	PidPressSolution = `
1. Node {{ .ObjectMeta.Name }} is reporting Pid pressure issue.
2. PIDPressure message is: {{ range $index, $value := .Status.Conditions }}{{- if eq $value.Type "PIDPressure" }}{{$value.Message}}{{- end}}{{end}}.
3. You may need to ssh to the Node, kill unnecessary processes.
4. You can find Pods running on this Node, re-allocate them to other Nodes. Refer to below command 1.
`

	NetUnAvailableSolution = `
1. Node {{ .ObjectMeta.Name }} is reporting network unavailable issue.
2. NetworkUnavailable message is: {{ range $index, $value := .Status.Conditions }}{{- if eq $value.Type "NetworkUnavailable" }}{{$value.Message}}{{- end}}{{end}}.
3. Please check if Kubelet service in node is started, and there's no network connection issue.
4. You may need to ssh to the Node and check the Kubelet service. Refer to below UsefulCommands.
`

	FindPoOnNoCmd = `
1. kubectl get pods -o wide -A | grep {{ .ObjectMeta.Name }}
`

	KubeletCmd = `
1. journalctl -u kubelet.
2. systemctl status kubelet.
3. systemctl restart kubelet.
`
)

func NodeNotReadyInvestigator(ctx context.Context, wg *sync.WaitGroup, problem *problem.Problem,
	input *problem.DetectorCreationInput) {
	defer wg.Done()
	getNodeCommonSolution(ctx, problem, NotReadySolution, KubeletCmd)
}

func NodeDiskPressureInvestigator(ctx context.Context, wg *sync.WaitGroup, problem *problem.Problem,
	input *problem.DetectorCreationInput) {
	defer wg.Done()
	getNodeCommonSolution(ctx, problem, DiskPressSolution, "")
}

func NodeMemoryPressureInvestigator(ctx context.Context, wg *sync.WaitGroup, problem *problem.Problem,
	input *problem.DetectorCreationInput) {
	defer wg.Done()
	getNodeCommonSolution(ctx, problem, MemPressSolution, FindPoOnNoCmd)
}

func NodePIDPressureInvestigator(ctx context.Context, wg *sync.WaitGroup, problem *problem.Problem,
	input *problem.DetectorCreationInput) {
	defer wg.Done()
	getNodeCommonSolution(ctx, problem, PidPressSolution, FindPoOnNoCmd)
}

func NodeNetworkUnavailableInvestigator(ctx context.Context, wg *sync.WaitGroup, problem *problem.Problem,
	input *problem.DetectorCreationInput) {
	defer wg.Done()
	getNodeCommonSolution(ctx, problem, NetUnAvailableSolution, KubeletCmd)
}

func getNodeCommonSolution(ctx context.Context, problem *problem.Problem, template string, cmdTemplate string) {
	no := *problem.AffectedResources.Resource.(*v1.Node)
	logChecking(ctx, com.Node+com.Blank+no.Name)
	solutions := GetSolutionsByTemplate(ctx, template, no, true)
	var cmd []string = nil
	if cmdTemplate != "" {
		cmd = GetSolutionsByTemplate(ctx, cmdTemplate, no, true)
	}

	appendSolution(problem, solutions, cmd)
}
