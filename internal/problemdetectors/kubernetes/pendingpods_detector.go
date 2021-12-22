package kubernetes

import (
	"context"
	"net/url"
	"strings"

	"github.com/fidelity/theliv/internal/problem"
	observability "github.com/fidelity/theliv/pkg/observability"
	v1 "k8s.io/api/core/v1"
)

var _ problem.Detector = (*PendingPodsDetector)(nil)

const (
	PendingPods             = "PendingPods"
	PendingPodsDetectorName = PendingPods + "Detector"

	PendingNodeSelector       = "node(s) didn't match node selector"
	PendingNodeAffinity       = "node(s) didn't match pod affinity"
	PendingNodeTaint          = "the pod didn't tolerate."
	PendingNodeUnschedulable  = "node(s) were unschedulable"
	PendingPVCNotFound        = "error getting PVC"
	PendingUnboundPVC         = "pod has unbound immediate PersistentVolumeClaims"
	PendingPVCProvisionFailed = "Failed to bind volumes"
	PendingInsufficient       = "Insufficient"
	PendingNoHostPort         = "node(s) didn't have free ports for the requested pod ports"
	PendingPodsDocLink        = "https://www.datadoghq.com/blog/debug-kubernetes-pending-pods"
)

var PendingPodsTags = []string{strings.ToLower(PendingPods), "kubelet"}

func RegisterPendingPodsWithProblemDomain(regFunc func(problem.DetectorRegistration, problem.DomainName) error) error {

	err := regFunc(problem.DetectorRegistration{
		Registration: problem.Registration{
			Name:          problem.DetectorName(PendingPodsDetectorName),
			Description:   "This detector will detect PendingPods error",
			Documentation: `Pods pending scheduling`,
			Supports:      []problem.SupportedPlatform{problem.EKS_Platform, problem.AKS_Platform},
		},
		CreateFunc: NewPendingPodsDetector,
	}, problem.PodFailuresDomain)
	return err
}

func NewPendingPodsDetector(i *problem.DetectorCreationInput) (problem.Detector, error) {
	return PendingPodsDetector{
		ResourceCommonDetector{
			name:          PendingPodsDetectorName,
			DetectorInput: i,
		}}, nil
}

type PendingPodsDetector struct {
	ResourceCommonDetector
}

var PendingPodsSolutions = map[string]func(pod *v1.Pod, status *v1.ContainerStatus, problemInput *ProblemInput,
	msg *string, e *observability.EventRecord) ([]string, map[problem.DeeplinkType]*url.URL){
	PendingNodeSelector:       getPendingPodCommonSolution(PendingNodeSelectorSolution),
	PendingNodeTaint:          getPendingPodCommonSolution(PendingNodeTaintSolution),
	PendingNodeUnschedulable:  getPendingPodCommonSolution(PendingNodeUnschedulableSolution),
	PendingPVCNotFound:        getPendingPodCommonSolution(PVCNotFoundSolution),
	PendingUnboundPVC:         getPendingPodCommonSolution(PVCUnboundSolution),
	PendingPVCProvisionFailed: getPendingPodCommonSolution(PVCUnboundSolution),
	PendingInsufficient:       getPendingPodCommonSolution(PendingInsufficientSolution),
	PendingNodeAffinity:       getPendingPodCommonSolution(PendingNodeAffinitySolution),
	PendingNoHostPort:         getPendingPodCommonSolution(PendingNoHostPortSolution),
}

// Check all pods in Pending phase.
// When event client is provided, check the events to figure out root cause
// The problem with same root cause will be consolidated into one problem.Problem, the corresponding pods will be
// added to AffectedResources.
func (d PendingPodsDetector) Detect(ctx context.Context) ([]problem.Problem, error) {
	problemInput := &ProblemInput{
		Detector:       &d.ResourceCommonDetector,
		PodSkipDetect:  podSkipDetectPendingPods,
		PodDetect:      podDetectPendingPods,
		SolutionsLinks: PendingPodsSolutions,
		Title:          PendingPods,
		Tags:           PendingPodsTags,
		DocLink:        PendingPodsDocLink,
	}
	return PodsDetect(ctx, problemInput)
}

func podSkipDetectPendingPods(po v1.Pod) bool {
	return !podUnschedulableCheck(po)
}

func podDetectPendingPods(status v1.ContainerStatus) bool {
	return true
}

func podUnschedulableCheck(po v1.Pod) bool {
	if po.Status.Phase == v1.PodPending {
		for _, con := range po.Status.Conditions {
			if con.Type == "PodScheduled" && con.Status == "False" {
				return true
			}
		}
	}
	return false
}

func getPendingPodCommonSolution(solution string) func(pod *v1.Pod, status *v1.ContainerStatus, problemInput *ProblemInput,
	msg *string, e *observability.EventRecord) ([]string, map[problem.DeeplinkType]*url.URL) {
	return func(pod *v1.Pod, status *v1.ContainerStatus, problemInput *ProblemInput, msg *string,
		e *observability.EventRecord) ([]string, map[problem.DeeplinkType]*url.URL) {
		return GetSolutionsByTemplate(solution, pod, true), getLogEventLinks(pod, problemInput, true, false, false, e)
	}
}
