/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package kubernetes

import (
	"context"
	"fmt"
	golog "log"
	"net/url"
	"strings"
	"time"

	"github.com/fidelity/theliv/internal/problem"
	"github.com/fidelity/theliv/pkg/kubeclient"
	observability "github.com/fidelity/theliv/pkg/observability"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	DeploymentFailureDetectorName = "DeploymentFailureDetector"
	DeploymentFailureDocLink      = "https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#failed-deployment"
	DeploymentFailureTitle        = "DeploymentNotAvailable"
	NotAvailableMsg               = "Deployment does not have minimum availability"
	InsufficientQuotaMsg          = "Insufficient quota"
	NotAvailableSolution          = "Please check the replica(s) status in this deployment."
	ResourceQuotaSolution         = "Please check the requests/limits of your deployment, whether it is in line with the ResourceQuota limits set by the cluster administrator."
	EventDeploymentQueryType      = "kube_deployment"
)

type DeploymentFailureDetector struct {
	ResourceCommonDetector
}

type DeploymentProblemInput struct {
	Detector       *DeploymentFailureDetector
	SolutionsLinks map[string]func(deployment *appsv1.Deployment, problemInput *DeploymentProblemInput, msg *string,
		e *observability.EventRecord) ([]string, map[problem.DeeplinkType]*url.URL)
	Title                 string
	Tags                  []string
	DocLink               string
	CheckEvent            *bool
	PossibleErrorMessages map[string][]string
}

// compiler to validate if the struct indeed implements the interface
var _ problem.Detector = (*DeploymentFailureDetector)(nil)

var DeploymentFailureTags = []string{"deploymentnotavailable", "kubelet"}

var DeploymentFailureSolutions = map[string]func(deployment *appsv1.Deployment, problemInput *DeploymentProblemInput,
	msg *string, e *observability.EventRecord) ([]string, map[problem.DeeplinkType]*url.URL){
	NotAvailableMsg:      getSolutionNotAvailableMsg,
	InsufficientQuotaMsg: getSolutionsInsufficientQuota,
}

func RegisterDeploymentFailureWithProblemDomain(regFunc func(problem.DetectorRegistration, problem.DomainName) error) error {
	err := regFunc(problem.DetectorRegistration{
		Registration: problem.Registration{
			Name:          problem.DetectorName(DeploymentFailureDetectorName),
			Description:   "This detector will detect DeploymentFailure error",
			Documentation: `Deployment not available`,
			Supports:      []problem.SupportedPlatform{problem.EKS_Platform, problem.AKS_Platform},
		},
		CreateFunc: NewDeploymentFailure,
	}, problem.DeploymentFailuresDomain)
	return err
}

func NewDeploymentFailure(i *problem.DetectorCreationInput) (problem.Detector, error) {
	return DeploymentFailureDetector{
		ResourceCommonDetector{
			name:          DeploymentFailureDetectorName,
			DetectorInput: i,
		}}, nil
}

func (d DeploymentFailureDetector) Detect(ctx context.Context) ([]problem.Problem, error) {
	fmt.Println("Running -> DeploymentFailureDetector")
	problemInput := &DeploymentProblemInput{
		Detector:       &d,
		SolutionsLinks: DeploymentFailureSolutions,
		Title:          DeploymentFailureTitle,
		Tags:           DeploymentFailureTags,
		DocLink:        DeploymentFailureDocLink,
	}
	return DeploymentDetect(ctx, problemInput)
}

func DeploymentDetect(ctx context.Context, problemInput *DeploymentProblemInput) ([]problem.Problem, error) {
	client, err := kubeclient.NewKubeClient(problemInput.Detector.DetectorInput.Kubeconfig)
	if err != nil {
		golog.Printf("ERROR - Got error when getting deployment client with kubeclient, error is %s", err)
	}
	namespace := kubeclient.NamespacedName{
		Namespace: problemInput.Detector.DetectorInput.Namespace,
	}
	deployments := &appsv1.DeploymentList{}
	deploymentListOptions := metav1.ListOptions{}
	client.List(ctx, deployments, namespace, deploymentListOptions)

	problems := make([]problem.Problem, 0)

	for _, deployment := range deployments.Items {
		if deployment.Status.UnavailableReplicas > 0 {
			// Should report the problem
			var msg string
			var condition appsv1.DeploymentCondition
			foundRootCause := false
			golog.Printf("INFO - Found %s has %d unavailable replicas", deployment.Name, deployment.Status.UnavailableReplicas)
			if ok, cd := containsCdt(deployment.Status.Conditions, "ReplicaFailure"); ok {
				if cd.Status == "True" {
					golog.Printf("INFO - Found %s has ReplicaFailure. Message: %s", deployment.Name, cd.Message)
					if strings.Contains(strings.ToLower(cd.Message), "cpu") ||
						strings.Contains(strings.ToLower(cd.Message), "memory") ||
						strings.Contains(strings.ToLower(cd.Message), "exceeded quota") {
						msg = InsufficientQuotaMsg
						condition = *cd
						foundRootCause = true
					}
				}
			} else if ok, cd := containsCdt(deployment.Status.Conditions, "Available"); ok {
				if cd.Status == "False" {
					msg = NotAvailableMsg
					condition = *cd
					foundRootCause = true
				}
			}
			if foundRootCause {
				addToDeploymentProblem(&deployment, &condition, &msg, &problems, problemInput, nil)
			}
		}
	}
	return problems, err
}

func addToDeploymentProblem(deployment *appsv1.Deployment, condition *appsv1.DeploymentCondition, msg *string,
	problems *[]problem.Problem, problemInput *DeploymentProblemInput, e *observability.EventRecord) {
	affectedResources := make(map[string]problem.ResourceDetails)
	solutions, deeplinks := getDeploymentSolutionLinks(deployment, problemInput, msg, e)

	affectedResources[deployment.Name] = problem.ResourceDetails{
		Deeplink:  deeplinks,
		Resource:  deployment.DeepCopyObject(),
		NextSteps: solutions,
	}

	doc, err := url.Parse(problemInput.DocLink)
	if err != nil {
		golog.Printf("WARN - error occurred creating Problem.Docs, error is %s", err)
	}
	discription := condition.Message
	prob := &problem.Problem{
		DomainName:        problemInput.Detector.Domain(),
		Name:              problemInput.Title,
		Description:       discription,
		Tags:              problemInput.Tags,
		Docs:              []*url.URL{doc},
		AffectedResources: affectedResources,
		Level:             problem.UserNamespace,
	}

	*problems = append(*problems, *prob)
}

func getDeploymentSolutionLinks(deployment *appsv1.Deployment, problemInput *DeploymentProblemInput, msg *string,
	e *observability.EventRecord) ([]string, map[problem.DeeplinkType]*url.URL) {
	var nextSteps []string
	if solution, ok := problemInput.SolutionsLinks[*msg]; ok {
		if solution != nil {
			nextSteps, _ = solution(deployment, problemInput, msg, e)
		} else {
			nextSteps, _ = getDeploymentSolutionUnknown(deployment, problemInput, msg)
		}
	} else {
		nextSteps, _ = getDeploymentSolutionUnknown(deployment, problemInput, msg)
	}
	return nextSteps, nil
}

func getSolutionNotAvailableMsg(deployment *appsv1.Deployment, problemInput *DeploymentProblemInput, msg *string,
	e *observability.EventRecord) ([]string, map[problem.DeeplinkType]*url.URL) {
	return []string{NotAvailableSolution}, getDeploymentLogEventLinks(deployment, problemInput, e)
}

func getSolutionsInsufficientQuota(deployment *appsv1.Deployment, problemInput *DeploymentProblemInput, msg *string,
	e *observability.EventRecord) ([]string, map[problem.DeeplinkType]*url.URL) {
	return []string{ResourceQuotaSolution}, getDeploymentLogEventLinks(deployment, problemInput, e)
}

func getDeploymentLogEventLinks(deployment *appsv1.Deployment, problemInput *DeploymentProblemInput, event *observability.EventRecord) map[problem.DeeplinkType]*url.URL {
	linkMap := map[problem.DeeplinkType]*url.URL{}
	var startTime, endTime time.Time
	if event == nil {
		startTime = deployment.ObjectMeta.CreationTimestamp.Time
		endTime = time.Now()
	} else {
		startTime = SetStartTime(event.DateHappend, EventLogTimespan)
		endTime = SetEndTime(event.DateHappend, EventLogTimespan)
	}
	linkMap[problem.DeeplinkEvent] = getDeploymentEventLink(deployment, problemInput, startTime, endTime)
	return linkMap
}

func getDeploymentEventLink(deployment *appsv1.Deployment, problemInput *DeploymentProblemInput, startTime time.Time,
	endTime time.Time) *url.URL {
	var url *url.URL
	var err error
	// The cluster name and namespace may be empty in the v1.Pod, use them from DectectorInput.
	clusterName := problemInput.Detector.DetectorInput.ClusterName
	namespace := problemInput.Detector.DetectorInput.Namespace
	eventDeeplinkRetriever := problemInput.Detector.DetectorInput.EventDeeplinkRetriever
	if eventDeeplinkRetriever != nil {
		eventsLink := eventDeeplinkRetriever.GetEventDeepLink(
			EventDeploymentQueryType, clusterName, namespace, deployment.Name, startTime, endTime)
		url, err = url.Parse(eventsLink)
		if err != nil {
			golog.Printf("WARN - Event url generation failed %s", err)
		}
	}
	return url
}

func getDeploymentSolutionUnknown(deployment *appsv1.Deployment, problemInput *DeploymentProblemInput,
	msg *string) ([]string, map[problem.DeeplinkType]*url.URL) {
	return []string{"Unkown root cause."}, getDeploymentLogEventLinks(deployment, problemInput, nil)
}

func containsCdt(conditions []appsv1.DeploymentCondition, ctype string) (bool, *appsv1.DeploymentCondition) {
	for _, c := range conditions {
		if string(c.Type) == ctype {
			return true, &c
		}
	}
	return false, nil
}

func (d DeploymentFailureDetector) Domain() problem.DomainName {
	return problem.DeploymentFailuresDomain
}
