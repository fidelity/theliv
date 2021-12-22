package kubernetes

import (
	"context"
	"fmt"
	golog "log"
	"net/url"

	"github.com/fidelity/theliv/internal/problem"
	"github.com/fidelity/theliv/pkg/kubeclient"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type ServiceFailureDetector struct {
	DetectorInput *problem.DetectorCreationInput
	name          string
}

type ServiceProblemInput struct {
	Detector              *ServiceFailureDetector
	SolutionsLinks        map[string]func(Service *v1.Service, problemInput *ServiceProblemInput, msg *string) []string
	Title                 string
	Tags                  []string
	DocLink               string
	CheckEvent            *bool
	PossibleErrorMessages map[string][]string
}

const (
	SerivceFailureDetectorName = "ServiceFailureDetector"
	ServiceFailureDocLink      = "https://kubernetes.io/docs/tasks/debug-application-cluster/debug-service/"
	ServiceFailureTitle        = "ServiceNotAvailable"
	NoPodAvailableMsg          = "No pods available that matching the service selector"
	NoneEdpMsg                 = "Endpoints not created for the service"
	NotExposeNamedPortMsg      = "Pod's named port not matching with service targetPort"
	EdpNumberMismatchMsg       = "Endpoints number mismatch with pods number"
	SvcTargetPortMismatchMsg   = "Service target port not matching with container port"
	SvcHasNoSelectorMsg        = "Service has no selectors"
)

var _ problem.Detector = (*ServiceFailureDetector)(nil)

var ServiceFailureTags = []string{"servicefailure", "kubelet"}

var ServiceFailureSolutions = map[string]func(service *v1.Service, problemInput *ServiceProblemInput,
	msg *string) []string{
	NoPodAvailableMsg:        getSolutionNoPodAvailableMsg,
	NoneEdpMsg:               getSolutionNoneEdpMsg,
	NotExposeNamedPortMsg:    getSolutionNotExposeNamedPortMsg,
	EdpNumberMismatchMsg:     getSolutionEdpNumberMismatchMsg,
	SvcTargetPortMismatchMsg: getSolutionSvcTargetPortMismatchMsg,
	SvcHasNoSelectorMsg:      getSolutionSvcHasNoSelectorMsg,
}

func RegisterServiceFailureWithProblemDomain(regFunc func(problem.DetectorRegistration, problem.DomainName) error) error {
	err := regFunc(problem.DetectorRegistration{
		Registration: problem.Registration{
			Name:          problem.DetectorName(SerivceFailureDetectorName),
			Description:   "This detector will detect ServiceFailure error",
			Documentation: `Service not available`,
			Supports:      []problem.SupportedPlatform{problem.EKS_Platform, problem.AKS_Platform},
		},
		CreateFunc: NewServiceFailure,
	}, problem.ServiceFailuresDomain)
	return err
}

func NewServiceFailure(i *problem.DetectorCreationInput) (problem.Detector, error) {
	return ServiceFailureDetector{
		name:          SerivceFailureDetectorName,
		DetectorInput: i,
	}, nil
}

func (s ServiceFailureDetector) Detect(ctx context.Context) ([]problem.Problem, error) {
	fmt.Println("Running -> ServiceFailureDetector")
	problemInput := &ServiceProblemInput{
		Detector:       &s,
		SolutionsLinks: ServiceFailureSolutions,
		Title:          ServiceFailureTitle,
		Tags:           ServiceFailureTags,
		DocLink:        ServiceFailureDocLink,
	}
	return ServiceDetect(ctx, problemInput)
}

func ServiceDetect(ctx context.Context, problemInput *ServiceProblemInput) ([]problem.Problem, error) {
	client, err := kubeclient.NewKubeClient(problemInput.Detector.DetectorInput.Kubeconfig)
	if err != nil {
		golog.Printf("ERROR - Got error when getting service client with kubeclient, error is %s", err)
	}
	namespace := kubeclient.NamespacedName{
		Namespace: problemInput.Detector.DetectorInput.Namespace,
	}
	svclist := &v1.ServiceList{}
	listOptions := metav1.ListOptions{}
	getOptions := metav1.GetOptions{}
	client.List(ctx, svclist, namespace, listOptions)

	problems := make([]problem.Problem, 0)
	for _, svc := range svclist.Items {
		var msg string
		foundProblem := false
		edp := &v1.Endpoints{}
		resName := kubeclient.NamespacedName{
			Namespace: problemInput.Detector.DetectorInput.Namespace,
			Name:      svc.Name,
		}
		err = client.Get(ctx, edp, resName, getOptions)
		if err != nil {
			golog.Printf("ERROR - Got error when getting endpoints with kubeclient, error is %s", err)
		}
		if svc.Spec.Selector != nil {
			pods := &v1.PodList{}
			err := client.List(ctx, pods, kubeclient.NamespacedName{
				Namespace: problemInput.Detector.DetectorInput.Namespace,
			}, metav1.ListOptions{
				LabelSelector: labels.SelectorFromSet(svc.Spec.Selector).String(),
			})
			if err != nil {
				golog.Printf("ERROR - Got error when getting pods with kubeclient, error is %s", err)
			} else {
				foundProblem, msg = detectEndpointProblem(&svc, edp, pods)
			}
		} else {
			// svc has no selector, check if endpoint is create with IP address.
			foundProblem, msg = detectEndpointWithoutSelectorProblem(&svc, edp)
		}
		if foundProblem {
			addToServiceProblem(&svc, &msg, &problems, problemInput)
		}
	}
	return problems, nil
}

func detectEndpointProblem(svc *v1.Service, edp *v1.Endpoints, pods *v1.PodList) (bool, string) {
	var msg string
	var foundProblem bool
	if len(edp.Subsets) == 0 {
		// no endpoints found
		foundProblem = true
		msg = NoneEdpMsg
		if len(pods.Items) == 0 {
			// no pods available match with service selector
			msg = NoPodAvailableMsg
		} else {
			//check named port
			foundProblem, msg = checkTargetPort(svc, pods)
		}
	} else if len(edp.Subsets[0].Addresses) != len(pods.Items) {
		// edp number mismatch with pod numbers
		foundProblem = true
		msg = EdpNumberMismatchMsg
	} else {
		//check target ports
		foundProblem, msg = checkTargetPort(svc, pods)
	}
	return foundProblem, msg
}

func detectEndpointWithoutSelectorProblem(svc *v1.Service, edp *v1.Endpoints) (bool, string) {
	// svc has no selector, check if endpoint is create with IP address.
	var msg string
	var foundProblem bool = false
	if len(edp.Subsets) == 0 || len(edp.Subsets[0].Addresses) == 0 {
		// no endpoints found, or no IP
		foundProblem = true
		msg = NoneEdpMsg
	}
	return foundProblem, msg
}

func checkTargetPort(svc *v1.Service, pods *v1.PodList) (bool, string) {
	var msg string
	var foundProblem bool
	for _, svcport := range svc.Spec.Ports {
		for _, pod := range pods.Items {
			var containerports []int32
			var containerportnames []string
			for _, container := range pod.Spec.Containers {
				for _, port := range container.Ports {
					containerports = append(containerports, port.ContainerPort)
					containerportnames = append(containerportnames, port.Name)
				}
			}
			if svcport.TargetPort.IntVal != 0 && !containsInt(containerports, svcport.TargetPort.IntVal) {
				// container port not match with service target port
				foundProblem = true
				msg = SvcTargetPortMismatchMsg
			}
			if svcport.TargetPort.IntVal == 0 && !containsStr(containerportnames, svcport.TargetPort.StrVal) {
				// container port not match with service target port
				foundProblem = true
				msg = NotExposeNamedPortMsg
			}
		}
	}
	return foundProblem, msg
}

func addToServiceProblem(service *v1.Service, msg *string,
	problems *[]problem.Problem, problemInput *ServiceProblemInput) {
	affectedResources := make(map[string]problem.ResourceDetails)
	solutions := getServiceSolutionLinks(service, problemInput, msg)

	affectedResources[service.Name] = problem.ResourceDetails{
		Resource:  service.DeepCopyObject(),
		NextSteps: solutions,
	}

	doc, err := url.Parse(problemInput.DocLink)
	if err != nil {
		golog.Printf("WARN - error occurred creating Problem.Docs, error is %s", err)
	}
	prob := &problem.Problem{
		DomainName:        problemInput.Detector.Domain(),
		Name:              problemInput.Title,
		Description:       *msg,
		Tags:              problemInput.Tags,
		Docs:              []*url.URL{doc},
		AffectedResources: affectedResources,
		Level:             problem.UserNamespace,
	}
	*problems = append(*problems, *prob)
}

func getServiceSolutionLinks(service *v1.Service, problemInput *ServiceProblemInput, msg *string) []string {
	var nextSteps []string
	if solution, ok := problemInput.SolutionsLinks[*msg]; ok {
		if solution != nil {
			nextSteps = solution(service, problemInput, msg)
		} else {
			nextSteps = getServiceSolutionUnknown(service, problemInput, msg)
		}
	} else {
		nextSteps = getServiceSolutionUnknown(service, problemInput, msg)
	}
	return nextSteps
}

func getServiceSolutionUnknown(service *v1.Service, problemInput *ServiceProblemInput,
	msg *string) []string {
	return []string{"Unknown root cause."}
}

func getSolutionNoPodAvailableMsg(service *v1.Service, problemInput *ServiceProblemInput, msg *string) []string {
	solution1 := "No endpoints created for the service '%s'."
	solution1 = fmt.Sprintf(solution1, service.Name)
	solution2 := "No pods available matching the service selector: '%s'. Please make sure the service selector matches pod selector."
	solution2 = fmt.Sprintf(solution2, labels.SelectorFromSet(service.Spec.Selector).String())
	return []string{solution1, solution2}
}

func getSolutionNoneEdpMsg(service *v1.Service, problemInput *ServiceProblemInput, msg *string) []string {
	solution1 := "No endpoints created for the service '%s'. Or the endpoints has no IP address."
	solution1 = fmt.Sprintf(solution1, service.Name)
	solution2 := "When using service without selectors, please make sure the endpoint is created with IP address."
	return []string{solution1, solution2}
}

func getSolutionNotExposeNamedPortMsg(service *v1.Service, problemInput *ServiceProblemInput, msg *string) []string {
	solution1 := "No endpoints created for the service '%s'."
	solution1 = fmt.Sprintf(solution1, service.Name)
	solution2 := "Service targetPort not match with any of the Named Ports. Please make sure container has exposed the correct Named Port."
	return []string{solution1, solution2}
}

func getSolutionEdpNumberMismatchMsg(service *v1.Service, problemInput *ServiceProblemInput, msg *string) []string {
	solution1 := "The number of endpoints does not match the number of pods."
	solution2 := "Please check pods details to see if they are in running status."
	return []string{solution1, solution2}
}

func getSolutionSvcTargetPortMismatchMsg(service *v1.Service, problemInput *ServiceProblemInput, msg *string) []string {
	solution1 := "Service targetPort value doesn't match with any of the container ports. Please make sure service is using targetPort that matches the container port."
	return []string{solution1}
}

func getSolutionSvcHasNoSelectorMsg(service *v1.Service, problemInput *ServiceProblemInput, msg *string) []string {
	solution := "Service does not have any selectors. Please add Selector field in service."
	return []string{solution}
}

func containsInt(arr []int32, i int32) bool {
	for _, a := range arr {
		if a == i {
			return true
		}
	}
	return false
}

func containsStr(arr []string, i string) bool {
	for _, a := range arr {
		if a == i {
			return true
		}
	}
	return false
}

func (d ServiceFailureDetector) Name() string {
	return d.name
}

func (d ServiceFailureDetector) Domain() problem.DomainName {
	return problem.ServiceFailuresDomain
}
