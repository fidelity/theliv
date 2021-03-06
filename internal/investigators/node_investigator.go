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
	v1 "k8s.io/api/core/v1"
)

const (
	NotReadySolution = `
1. Node {{ .ObjectMeta.Name }} is not Ready.
2. NotReady message is: {{ range $index, $value := .Status.Conditions }}{{- if eq $value.Type "Ready" }}{{$value.Message}}{{- end}}{{end}}.
3. Please wait for sometime, if issue persists, you may need to ssh to the Node and check.
4. Cmd to check Kubelet log: journalctl -u kubelet.
5. Cmd to check Kubelet service status: systemctl status kubelet.
6. Cmd to restart Kubelet service: systemctl restart kubelet.
`

	MemPressSolution = `
1. Node {{ .ObjectMeta.Name }} is reporting memory pressure issue.
2. MemoryPressure message is: {{ range $index, $value := .Status.Conditions }}{{- if eq $value.Type "MemoryPressure" }}{{$value.Message}}{{- end}}{{end}}.
3. You may need to ssh to the Node, kill unnecessary processes.
4. You can find Pods running on this Node, re-allocate some to other Nodes.
5. Cmd to find Pods on this Node: kubectl get pods -o wide -A | grep {{ .ObjectMeta.Name }}
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
4. You can find Pods running on this Node, re-allocate some to other Nodes.
5. Cmd to find Pods on this Node: kubectl get pods -o wide -A | grep {{ .ObjectMeta.Name }}
`

	NetUnAvailableSolution = `
1. Node {{ .ObjectMeta.Name }} is reporting network unavailable issue.
2. NetworkUnavailable message is: {{ range $index, $value := .Status.Conditions }}{{- if eq $value.Type "NetworkUnavailable" }}{{$value.Message}}{{- end}}{{end}}.
3. Please check if Kubelet service in node is started, and there's no network connection issue.
4. You may need to ssh to the Node and check the Kubelet service.
5. Cmd to check Kubelet log: journalctl -u kubelet.
6. Cmd to check Kubelet service status: systemctl status kubelet.
7. Cmd to restart Kubelet service: systemctl restart kubelet.
`
)

func NodeNotReadyInvestigator(ctx context.Context, problem *problem.Problem,
	input *problem.DetectorCreationInput) {
	getNodeCommonSolution(problem, NotReadySolution)
}

func NodeDiskPressureInvestigator(ctx context.Context, problem *problem.Problem,
	input *problem.DetectorCreationInput) {
	getNodeCommonSolution(problem, DiskPressSolution)
}

func NodeMemoryPressureInvestigator(ctx context.Context, problem *problem.Problem,
	input *problem.DetectorCreationInput) {
	getNodeCommonSolution(problem, MemPressSolution)
}

func NodePIDPressureInvestigator(ctx context.Context, problem *problem.Problem,
	input *problem.DetectorCreationInput) {
	getNodeCommonSolution(problem, PidPressSolution)
}

func NodeNetworkUnavailableInvestigator(ctx context.Context, problem *problem.Problem,
	input *problem.DetectorCreationInput) {
	getNodeCommonSolution(problem, NetUnAvailableSolution)
}

func getNodeCommonSolution(problem *problem.Problem, template string) {
	no := *problem.AffectedResources.Resource.(*v1.Node)
	logChecking(com.Node + com.Blank + no.Name)
	solutions := GetSolutionsByTemplate(template, no, true)
	appendSolution(problem, solutions)
}
