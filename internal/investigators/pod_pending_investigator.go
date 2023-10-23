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
	v1 "k8s.io/api/core/v1"
)

const (
	NodesNotAvailable        = "0/.* nodes are available"
	PendingNodeSelector      = "didn't match .*selector"
	PendingNodeAffinity      = "didn't match .*affinity"
	PendingNodeTaint         = "untolerated taint"
	PendingNodeUnschedulable = "unschedulable"
	PendingInsufficient      = "Insufficient"
	PendingNoHostPort        = "node(s) didn't have free ports"

	PendingPVCGetErr      = "error getting PVC"
	PendingPVCNotFound    = "persistentvolumeclaim .* not found"
	PendingUnboundPVC     = "pod has unbound immediate PersistentVolumeClaims"
	PendingBindFailed     = "Failed to bind volumes"
	PendingCmNotFound     = "configmap .* not found"
	PendingSecretNotFound = "secret .* not found" //nolint:gosec
)

const (
	FailedSchedulingMessage          = "%d. Pod failed scheduling, message is: %s."
	NodeUnavailableSolution          = "%d. No node is available for the Pod, you may need to fix the issue in NotReady Node, or add new Node."
	NodeUnavailableRef               = "%d. Refer to: https://kubernetes.io/docs/tasks/debug-application-cluster/debug-cluster/"
	PendingNodeUnschedulableSolution = "%d. Some nodes are unschedulable, try to uncordon these nodes may fix this."

	PendingNodeSelectorSolution = "%d. Some nodes don't match the Pod node-selector/affinity, can check and adjust Pod node-selector/affinity."
	PendingNodeSelectorRef      = "%d. Refer to: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#node-affinity"
	PendingNodeTaintSolution    = "%d. Some node(s) had taints, that the pod didn't tolerate. Try to modify the pod to tolerate 1 of them."
	PendingNodeTaintRef         = "%d. Refer to: https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/"

	PendingInsufficientSolution = "%d. Some available node(s) has insufficient resources, check the resources that the pod requests or limits, try to modify them to applicable quota."
	PendingNoHostPortSolution   = "%d. Available node(s) didn't have free ports for the requested pod ports. Please check the HostPort used in the Pod, change/remove it is suggested."
	PVCNotFoundSolution         = "2. Pod {{ .ObjectMeta.Name }} is pending, used PVC not found." + KubectlPodAndPVC
	PVCUnboundSolution          = "2. Pod {{ .ObjectMeta.Name }} is pending, due to use an unbound PVC." + KubectlPodAndPVC
	KubectlPodAndPVC            = `
3. Please check PVC used by the pod, create new or choose an existing PVC may solve this problem.
4. Reference link: https://kubernetes.io/docs/concepts/storage/persistent-volumes/
`

	ContainerFailMount         = "%d. Container failed mount, message is: %s."
	ContainerFailMountSolution = "%d. Please check your volumes of the Pod, try to change to correct and existing resources may fix this problem."
	CmNotFoundSolution         = "%d. Please check the configMap that mount, try to change to an existing configMap may fix this issue."
	SecretNotFoundSolution     = "%d. Please check the secret that mount, try to change to an existing secret may fix this issue."

	PendingUnknownSolution = `
1. Pod {{ .ObjectMeta.Name }} is in Pending state for more than 5 mins.
2. The root cause can be any of below:
3. Your target node(s) may not be available, you may can restart or create new node.
4. Node(s) may have labels or taints, try to change your pod node-selector, affinity, tolerance, may fix this.
5. Existing node(s) may don't have enough resources, try to change your pod resource requests may fix this.
6. If you use any volumes, please make sure the resources you want to mount exists.
`
	KubeDescribePoCmd   = "%d. kubectl describe po {{.ObjectMeta.Name}} -n {{.ObjectMeta.Namespace}}"
	GetEventsCmd        = "%d. kubectl get events --field-selector involvedObject.name={{.ObjectMeta.Name}} -n {{.ObjectMeta.Namespace}}"
	GetNoAllCmd         = "%d. kubectl get no"
	GetNoAllocatableCmd = "%d. kubectl get no -o custom-columns=NAME:.metadata.name,ALLOCATABLE:.status.allocatable --no-headers"
	GetNoLabelCmd       = "%d. kubectl get no --show-labels"
	GetNoTaintCmd       = "%d. kubectl get no -o custom-columns=NAME:.metadata.name,TAINTS:.spec.taints --no-headers"
	UncordonCmd         = "%d. kubectl uncordon <node name>"

	GetPvcCmd    = "%d. kubectl get pvc -n {{ .ObjectMeta.Namespace }}"
	GetCmCmd     = "%d. kubectl get cm -n {{ .ObjectMeta.Namespace }}"
	GetSecretCmd = "%d. kubectl get secret -n {{ .ObjectMeta.Namespace }}"
)

func PodNotRunningInvestigator(ctx context.Context, wg *sync.WaitGroup, problem *problem.Problem, input *problem.DetectorCreationInput) {
	defer wg.Done()

	pod := *problem.AffectedResources.Resource.(*v1.Pod)

	var solutions []string
	var commands []string

	failScheduleEvent := getPoEventMsg(ctx, input, &pod, "FailedScheduling")
	failMount := getPoEventMsg(ctx, input, &pod, "FailedMount")

	if len(failScheduleEvent) > 0 {
		failSchedule := failScheduleEvent[0]
		solutions = []string{fmt.Sprintf(FailedSchedulingMessage, 1, failSchedule)}
		commands = []string{}
		if msgMatch(PendingPVCGetErr, failSchedule) {
			solutions, commands = appendPVSolution(ctx, pod, solutions, PVCNotFoundSolution)
		} else if msgMatch(PendingPVCNotFound, failSchedule) {
			solutions, commands = appendPVSolution(ctx, pod, solutions, PVCNotFoundSolution)
		} else if msgMatch(PendingUnboundPVC, failSchedule) {
			solutions, commands = appendPVSolution(ctx, pod, solutions, PVCUnboundSolution)
		} else if msgMatch(PendingBindFailed, failSchedule) {
			solutions, commands = appendPVSolution(ctx, pod, solutions, PVCUnboundSolution)
		} else if msgMatch(NodesNotAvailable, failSchedule) {
			solutions = appendSeq(solutions, NodeUnavailableSolution)
			solutions = appendSeq(solutions, NodeUnavailableRef)
			commands = appendSeq(commands, GetNoAllCmd)
			commands = appendSeq(commands, GetNoAllocatableCmd)
			if msgMatch(PendingNodeUnschedulable, failSchedule) {
				solutions = appendSeq(solutions, PendingNodeUnschedulableSolution)
				commands = appendSeq(commands, UncordonCmd)
			}
			if msgMatch(PendingNodeSelector, failSchedule) || msgMatch(PendingNodeAffinity, failSchedule) {
				solutions = appendSeq(solutions, PendingNodeSelectorSolution)
				solutions = appendSeq(solutions, PendingNodeSelectorRef)
				commands = appendSeq(commands, GetNoLabelCmd)
			}
			if msgMatch(PendingNodeTaint, failSchedule) {
				solutions = appendSeq(solutions, PendingNodeTaintSolution)
				solutions = appendSeq(solutions, PendingNodeTaintRef)
				commands = appendSeq(commands, GetNoTaintCmd)
			}
			if msgMatch(PendingInsufficient, failSchedule) {
				solutions = appendSeq(solutions, PendingInsufficientSolution)
			}
			if msgMatch(PendingNoHostPort, failSchedule) {
				solutions = appendSeq(solutions, PendingNoHostPortSolution)
			}
		} else {
			solutions, commands = getPendingPodUnknownSolution(ctx, pod)
		}
	} else if len(failMount) > 0 {
		commands = appendSeq(commands, GetSolutionsByTemplate(ctx, KubeDescribePoCmd, pod, true)[0])
		commands = appendSeq(commands, GetSolutionsByTemplate(ctx, GetEventsCmd, pod, true)[0])
		for _, event := range failMount {
			if msgMatch(PendingCmNotFound, event) {
				solutions = append(solutions, fmt.Sprintf(ContainerFailMount, 1, event))
				solutions = appendSeq(solutions, CmNotFoundSolution)
				commands = appendSeq(commands, GetSolutionsByTemplate(ctx, GetCmCmd, pod, true)[0])
			}
			if msgMatch(PendingSecretNotFound, event) {
				solutions = append(solutions, fmt.Sprintf(ContainerFailMount, 1, event))
				solutions = appendSeq(solutions, SecretNotFoundSolution)
				commands = appendSeq(commands, GetSolutionsByTemplate(ctx, GetSecretCmd, pod, true)[0])
			}
		}
		if len(solutions) == 0 {
			solutions = appendSeq(solutions, fmt.Sprintf(ContainerFailMount, 1, failMount[0]))
			solutions = appendSeq(solutions, ContainerFailMountSolution)
		}

	} else {
		solutions, commands = getPendingPodUnknownSolution(ctx, pod)
	}
	appendSolution(problem, solutions, commands)
}

func appendPVSolution(ctx context.Context, po v1.Pod, solutions []string, solution string) ([]string, []string) {
	addSolutions := GetSolutionsByTemplate(ctx, solution, po, true)
	solutions = append(solutions, addSolutions...)
	var commands []string
	commands = appendSeq(commands, GetSolutionsByTemplate(ctx, KubeDescribePoCmd, po, true)[0])
	commands = appendSeq(commands, GetSolutionsByTemplate(ctx, GetEventsCmd, po, true)[0])
	commands = appendSeq(commands, GetSolutionsByTemplate(ctx, GetPvcCmd, po, true)[0])
	return solutions, commands
}

func appendSeq(solution []string, message string) []string {
	order := len(solution) + 1
	return append(solution, fmt.Sprintf(message, order))
}

func msgMatch(msg1 string, msg2 string) bool {
	matched, err := regexp.MatchString(strings.ToLower(msg1), strings.ToLower(msg2))
	if matched && err == nil {
		return true
	}
	return false
}

func getPoEventMsg(ctx context.Context, input *problem.DetectorCreationInput, pod *v1.Pod, reason string) (msg []string) {
	events, err := GetPodEvents(ctx, input, pod)
	if err != nil {
		return
	}
	if len(events) > 0 {
		for _, event := range events {
			if event.Reason == reason && event.Message != "" {
				msg = append(msg, event.Message)
			}
		}
	}
	return
}

func PodNotRunningSolutionsInvestigator(ctx context.Context, wg *sync.WaitGroup, problem *problem.Problem, input *problem.DetectorCreationInput) {
	defer wg.Done()
	// Generate solutions
	// detail := "do something to provide solutions"
	// problem.SolutionDetails = append(problem.SolutionDetails, detail)
}

func getPendingPodUnknownSolution(ctx context.Context, po v1.Pod) ([]string, []string) {
	solutions := GetSolutionsByTemplate(ctx, PendingUnknownSolution, po, true)
	var commands []string
	commands = appendSeq(commands, GetSolutionsByTemplate(ctx, KubeDescribePoCmd, po, true)[0])
	commands = appendSeq(commands, GetSolutionsByTemplate(ctx, GetEventsCmd, po, true)[0])
	return solutions, commands
}
