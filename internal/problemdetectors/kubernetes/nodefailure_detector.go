package kubernetes

import (
	"context"
	"net/url"
	"strings"

	log "github.com/fidelity/theliv/pkg/log"

	"github.com/fidelity/theliv/internal/problem"
	"github.com/fidelity/theliv/pkg/kubeclient"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ problem.Detector = (*NodeFailureDetector)(nil)

const (
	NodeFailure             = "NodeFailure"
	NodeFailureDetectorName = NodeFailure + "Detector"
	DocLink                 = "https://kubernetes.io/docs/concepts/architecture/nodes/#node-status"

	UnSchedulableTaint    = "node.kubernetes.io/unschedulable"
	UnSchedulableMsg      = "node is unschedulable."
	NotReadyTaint         = "node.kubernetes.io/not-ready"
	NotReadyMsg           = "node is not ready."
	UnReachableTaint      = "node.kubernetes.io/unreachable"
	UnReachableMsg        = "node is unreachable."
	MemPressTaint         = "node.kubernetes.io/memory-pressure"
	MemPressMsg           = "node got memory-pressure."
	DiskPressTaint        = "node.kubernetes.io/disk-pressure"
	DiskPressMsg          = "node got disk-pressure."
	PidPressTaint         = "node.kubernetes.io/pid-pressure"
	PidPressMsg           = "node got pid-pressure."
	NetUnAvailableTaint   = "node.kubernetes.io/network-unavailable"
	NetUnAvailableMsg     = "node got network-unavailable."
	UnInitializedTaint    = "node.cloudprovider.kubernetes.io/uninitialized"
	UnInitializedMsg      = "node is uninitialized."
	NodeFailureGeneralErr = "Node NotReady or UnSchedulable, root-cause not found."
)

var NodeFailureTags = []string{strings.ToLower(NodeFailure), "kubelet"}

func RegisterNodeFailureWithProblemDomain(regFunc func(problem.DetectorRegistration, problem.DomainName) error) error {

	err := regFunc(problem.DetectorRegistration{
		Registration: problem.Registration{
			Name:          problem.DetectorName(NodeFailureDetectorName),
			Description:   "This detector will detect Node Failure error",
			Documentation: `Node status Not Ready or UnSchedulable`,
			Supports:      []problem.SupportedPlatform{problem.EKS_Platform, problem.AKS_Platform},
		},
		CreateFunc: NewNodeFailureDetector,
	}, problem.NodeFailuresDomain)
	return err
}

func NewNodeFailureDetector(i *problem.DetectorCreationInput) (problem.Detector, error) {
	return NodeFailureDetector{
		ResourceCommonDetector{
			name:          NodeFailureDetectorName,
			DetectorInput: i,
		}}, nil
}

type NodeFailureDetector struct {
	ResourceCommonDetector
}

var NodeTaintsMsg = map[string]string{
	UnSchedulableTaint:  UnSchedulableMsg,
	NotReadyTaint:       NotReadyMsg,
	UnReachableTaint:    UnReachableMsg,
	MemPressTaint:       MemPressMsg,
	DiskPressTaint:      DiskPressMsg,
	PidPressTaint:       PidPressMsg,
	NetUnAvailableTaint: NetUnAvailableMsg,
	UnInitializedTaint:  UnInitializedMsg,
}

var NodeFailureSolutions = map[string]func(no v1.Node) ([]string, map[problem.DeeplinkType]*url.URL){
	UnSchedulableMsg:      getNodeFailureCommonSolution(UnSchedulableSolution),
	NotReadyMsg:           getNodeFailureCommonSolution(NotReadySolution),
	UnReachableMsg:        getNodeFailureCommonSolution(UnReachableSolution),
	MemPressMsg:           getNodeFailureCommonSolution(MemPressSolution),
	DiskPressMsg:          getNodeFailureCommonSolution(DiskPressSolution),
	PidPressMsg:           getNodeFailureCommonSolution(PidPressSolution),
	NetUnAvailableMsg:     getNodeFailureCommonSolution(NetUnAvailableSolution),
	UnInitializedMsg:      getNodeFailureCommonSolution(UnInitializedSolution),
	NodeFailureGeneralErr: getNodeFailureCommonSolution(GeneralErrSolution),
}

// Check Nodes if NotReady or UnSchedulable
// If Node Controller found Node not ready or not Healthy, taints will be added to Node.
// Detector will check predefined taints, to find the root cause.
// Problem will be Cluster level.
func (d NodeFailureDetector) Detect(ctx context.Context) ([]problem.Problem, error) {

	client, err := kubeclient.NewKubeClient(d.DetectorInput.Kubeconfig)
	if err != nil {
		log.S().Errorf("Got error when getting deployment client with kubeclient, error is %s", err)
	}
	nodes := &v1.NodeList{}
	listOptions := metav1.ListOptions{}
	client.List(ctx, nodes, kubeclient.NamespacedName{}, listOptions)

	problems := make([]problem.Problem, 0)

	for _, no := range nodes.Items {
		hasFoundRootCause := false
		if getNodeCondition(no, v1.NodeReady) != v1.ConditionTrue || no.Spec.Unschedulable {
			for _, taint := range toTaintMap(no.Spec.Taints) {
				if msg, ok := NodeTaintsMsg[taint.Key]; ok {
					hasFoundRootCause = true
					existingProblem := getProblemByMsg(msg, &problems)
					addToNodeProblem(d, no, existingProblem, msg, &problems, NodeFailureSolutions[msg])
				}
			}
			if !hasFoundRootCause {
				existingProblem := getProblemByMsg(NodeFailureGeneralErr, &problems)
				addToNodeProblem(d, no, existingProblem, NodeFailureGeneralErr, &problems, NodeFailureSolutions[NodeFailureGeneralErr])
			}
		}
	}
	return problems, nil
}

// Add the problem to existing problem if they have same root cause, otherwise create a new one.
func addToNodeProblem(d NodeFailureDetector, no v1.Node, prob *problem.Problem, msg string,
	problems *[]problem.Problem, solution func(no v1.Node) ([]string, map[problem.DeeplinkType]*url.URL)) {
	if prob == nil || isEmptyProblem(*prob) {
		affectedResources := make(map[string]problem.ResourceDetails)
		solutions, _ := solution(no)

		affectedResources[no.Name] = problem.ResourceDetails{
			Deeplink:  nil,
			Resource:  no.DeepCopyObject(),
			NextSteps: solutions,
		}

		doc, err := url.Parse(DocLink)
		if err != nil {
			log.S().Warnf("error occurred creating Problem.Docs, error is %s", err)
		}

		prob = &problem.Problem{
			DomainName:        d.Domain(),
			Name:              NodeFailure,
			Description:       msg,
			Tags:              NodeFailureTags,
			Docs:              []*url.URL{doc},
			AffectedResources: affectedResources,
			Level:             problem.Cluster,
		}

		*problems = append(*problems, *prob)
	} else {
		p := getProblemByMsg(msg, problems)
		solutions, _ := solution(no)
		p.AffectedResources[no.Name] = problem.ResourceDetails{
			Deeplink:  nil,
			Resource:  no.DeepCopyObject(),
			NextSteps: solutions,
		}
	}
}

func getNodeCondition(no v1.Node, conditionType v1.NodeConditionType) v1.ConditionStatus {
	for _, con := range no.Status.Conditions {
		if con.Type == conditionType {
			return con.Status
		}
	}
	return v1.ConditionUnknown
}

func toTaintMap(taints []v1.Taint) (result map[string]v1.Taint) {
	result = map[string]v1.Taint{}
	if len(taints) > 0 {
		for _, value := range taints {
			result[value.Key] = value
		}
	}
	return
}

func getNodeFailureCommonSolution(msg string) func(no v1.Node) ([]string, map[problem.DeeplinkType]*url.URL) {
	return func(no v1.Node) ([]string, map[problem.DeeplinkType]*url.URL) {
		return GetSolutionsByTemplate(msg, no, true), nil
	}
}

func (d NodeFailureDetector) Domain() problem.DomainName {
	return problem.NodeFailuresDomain
}
