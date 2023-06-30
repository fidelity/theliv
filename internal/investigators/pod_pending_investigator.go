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

const (
	PendingContainer          = "containers with unready status"
	PendingNodeSelector       = "node(s) didn't match node selector"
	PendingNodeAffinity       = "node(s) didn't match pod affinity"
	PendingNodeTaint          = "the pod didn't tolerate."
	PendingNodeUnschedulable  = "node(s) were unschedulable"
	PendingPVCGetErr          = "error getting PVC"
	PendingPVCNotFound        = "persistentvolumeclaim .* not found"
	PendingUnboundPVC         = "pod has unbound immediate PersistentVolumeClaims"
	PendingPVCProvisionFailed = "Failed to bind volumes"
	PendingInsufficient       = "Insufficient"
	PendingNoHostPort         = "node(s) didn't have free ports for the requested pod ports"
)

const (
	PendingContainerSolution = `
1. Failed Schedule {{ .ObjectMeta.Name }}: containers with unready status.
2. Please check container status.
`
	PendingNodeSelectorSolution = `
1. Failed Schedule Pod {{ .ObjectMeta.Name }}: node(s) didn't match node selector.
2. Please check the Pod node-selector, affinity.
3. Reference link: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#node-affinity
`

	GetNoLabelCmd = `
2. kubectl get no --show-labels
`
	GetNoTaintCmd = `
3. kubectl get no -o custom-columns=NAME:.metadata.name,TAINTS:.spec.taints --no-headers
`
	GetNoAllCmd = `
1. kubectl get no
`
	UncordonCmd = `
2. kubectl uncordon <node name>
`
	GetPvcCmd = `
2. kubectl get pvc -n {{ .ObjectMeta.Namespace }}
`
	GetNoAllocatableCmd = `
2. kubectl get nodes -o custom-columns=NAME:.metadata.name,ALLOCATABLE:.status.allocatable --no-headers
`

	PendingNodeTaintSolution = `
1. Failed Schedule Pod {{ .ObjectMeta.Name }}: node(s) had taints, that the pod didn't tolerate.
2. Please check the each Node's taints, and modify the pod to tolerate 1 of them.
3. Reference link: https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/
`
	PendingNodeUnschedulableSolution = `
1. Failed Schedule Pod {{ .ObjectMeta.Name }}: node(s) were unschedulable.
2. Please check the Nodes STATUS, if all nodes Not Ready or SchedulingDisabled.
3. To fix this, you may need to fix the issue in NotReady Node, or uncordon the Node, or add 1 new Node.
4. Reference link: https://kubernetes.io/docs/tasks/debug-application-cluster/debug-cluster/
`
	PVCNotFoundSolution = "1. Pod {{ .ObjectMeta.Name }} is pending, used PVC not found." + KubectlPodAndPVC

	PVCUnboundSolution = "1. Pod {{ .ObjectMeta.Name }} is pending, due to use an unbound PVC." + KubectlPodAndPVC

	KubectlPodAndPVC = `
2. Please check PVC used by the pod.
3. Reference link: https://kubernetes.io/docs/concepts/storage/persistent-volumes/
`
	PendingInsufficientSolution = `
1. Failed Schedule Pod {{ .ObjectMeta.Name }}: available node(s) has insufficient resources.
2. Please check the resources that the pod requests or limits, try to modify them to applicable quota.
3. Reference link: https://kubernetes.io/docs/concepts/architecture/nodes/
`
	PendingNodeAffinitySolution = `
1. Failed Schedule Pod {{ .ObjectMeta.Name }}: available node(s) didn't match pod affinity/anti-affinity rules. Also some Nodes may include taints that the Pod didn't tolerate.
2. Please check the Pod affinity/anti-affinity. Or check the Nodes' taints, and make the Pod tolerate 1 of them.
3. Reference link: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#affinity-and-anti-affinity
`
	PendingNoHostPortSolution = `
1. Failed Schedule Pod {{ .ObjectMeta.Name }}: available node(s) didn't have free ports for the requested pod ports.
2. Please check the HostPort used in the Pod, change/remove it is suggested.
3. If the HostPort is necessary, use NodeSelector to assign Pod to the Node which the specified Port is available.
`
	PendingUnknownSolution = `
1. Pod {{ .ObjectMeta.Name }} is not Running.
2. Either pod not scheduled, or container not ready.
`
	KubeDescribePoCmd = `
kubectl describe po {{.ObjectMeta.Name}} -n {{.ObjectMeta.Namespace}}
`
)

var PendingPodsSolutions = map[string]func(ctx context.Context, pod *v1.Pod, status *v1.ContainerStatus) ([]string, []string){
	PendingContainer:          getPendingPodCommonSolution(PendingContainerSolution, GetPoCmd),
	PendingNodeSelector:       getPendingPodCommonSolution(PendingNodeSelectorSolution, GetPoCmd+GetNoLabelCmd),
	PendingNodeTaint:          getPendingPodCommonSolution(PendingNodeTaintSolution, GetPoCmd+GetNoLabelCmd),
	PendingNodeUnschedulable:  getPendingPodCommonSolution(PendingNodeUnschedulableSolution, GetNoAllCmd+UncordonCmd),
	PendingPVCGetErr:          getPendingPodCommonSolution(PVCNotFoundSolution, GetPoCmd+GetPvcCmd),
	PendingPVCNotFound:        getPendingPodCommonSolution(PVCNotFoundSolution, GetPoCmd+GetPvcCmd),
	PendingUnboundPVC:         getPendingPodCommonSolution(PVCUnboundSolution, GetPoCmd+GetPvcCmd),
	PendingPVCProvisionFailed: getPendingPodCommonSolution(PVCUnboundSolution, GetPoCmd+GetPvcCmd),
	PendingInsufficient:       getPendingPodCommonSolution(PendingInsufficientSolution, GetPoCmd+GetNoAllocatableCmd),
	PendingNodeAffinity:       getPendingPodCommonSolution(PendingNodeAffinitySolution, GetPoCmd+GetNoLabelCmd+GetNoTaintCmd),
	PendingNoHostPort:         getPendingPodCommonSolution(PendingNoHostPortSolution, GetPoCmd+GetNoLabelCmd+GetNoTaintCmd),
}

func PodNotRunningInvestigator(ctx context.Context, problem *problem.Problem, input *problem.DetectorCreationInput) {

	pod := *problem.AffectedResources.Resource.(*v1.Pod)
	container := &v1.ContainerStatus{}

	if getPodSolutionFromEvents(ctx, problem, input, &pod, container, PendingPodsSolutions) == "" {
		solution, commands := getPendingPodCommonSolution(PendingUnknownSolution, KubeDescribePoCmd)(ctx, &pod, container)
		appendSolution(problem, solution, commands)
	}

}

func PodNotRunningSolutionsInvestigator(ctx context.Context, problem *problem.Problem, input *problem.DetectorCreationInput) {
	// Generate solutions
	// detail := "do something to provide solutions"
	// problem.SolutionDetails = append(problem.SolutionDetails, detail)
}

func getPendingPodCommonSolution(solution string, cmd string) func(ctx context.Context, pod *v1.Pod, status *v1.ContainerStatus) ([]string, []string) {
	return func(ctx context.Context, pod *v1.Pod, status *v1.ContainerStatus) ([]string, []string) {
		return GetSolutionsByTemplate(ctx, solution, pod, true), GetSolutionsByTemplate(ctx, cmd, pod, true)
	}
}
