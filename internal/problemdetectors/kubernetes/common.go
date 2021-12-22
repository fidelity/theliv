package kubernetes

import (
	"bytes"
	"fmt"
	"text/template"
	"time"

	"context"
	golog "log"
	"net/url"
	"strings"

	"github.com/fidelity/theliv/internal/problem"
	"github.com/fidelity/theliv/pkg/kubeclient"
	observability "github.com/fidelity/theliv/pkg/observability"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	LogsLink          = "Log deeplink: %s"
	EventsLink        = "Events deeplink: %s"
	EventPodQueryType = "pod_name"
	MaskedString      = "*"
	MaskedVisibleNo   = 5
	MaskedPrefixNo    = 5
)

var GeneralMsg = "Can not find detailed events to determine the root cause"

var DefaultTimespan = problem.TimeSpan{
	Timespan:     48,
	TimespanType: time.Hour,
}

var EventLogTimespan = problem.TimeSpan{
	Timespan:     10,
	TimespanType: time.Minute,
}

// A general template instance.
var solutionTemp = template.New("solutionTemp")

type ResourceCommonDetector struct {
	DetectorInput *problem.DetectorCreationInput
	name          string
}

type ProblemInput struct {
	Detector       *ResourceCommonDetector
	PodSkipDetect  PodSkipDetect
	PodDetect      PodDetect
	SolutionsLinks map[string]func(pod *v1.Pod, status *v1.ContainerStatus, problemInput *ProblemInput, msg *string,
		e *observability.EventRecord) ([]string, map[problem.DeeplinkType]*url.URL)
	Title                 string
	Tags                  []string
	DocLink               string
	CheckEvent            *bool
	PossibleErrorMessages map[string][]string
}

func (d ResourceCommonDetector) Name() string {
	return d.name
}

// Common Detector returns mostly used domain, PodFailures
// Detector with different domain should overwrite this.
func (d ResourceCommonDetector) Domain() problem.DomainName {
	return problem.PodFailuresDomain
}

func PodsDetect(ctx context.Context, problemInput *ProblemInput) ([]problem.Problem, error) {

	client, err := kubeclient.NewKubeClient(problemInput.Detector.DetectorInput.Kubeconfig)
	if err != nil {
		golog.Printf("ERROR - Got error when getting deployment client with kubeclient, error is %s", err)
	}
	namespace := kubeclient.NamespacedName{
		Namespace: problemInput.Detector.DetectorInput.Namespace,
	}

	pods := &v1.PodList{}
	podListOptions := metav1.ListOptions{}
	client.List(ctx, pods, namespace, podListOptions)

	checkEvent := true
	if problemInput.Detector.DetectorInput.EventRetriever == nil {
		golog.Println("INFO - EventRetriever is not provided, skip event analysis.")
		checkEvent = false
	}

	problemInput.CheckEvent = &checkEvent

	problems := make([]problem.Problem, 0)

	for _, pod := range pods.Items {
		if problemInput.PodSkipDetect(pod) {
			continue
		}

		golog.Printf("INFO - Checking if Unschedulable with pod %s", pod.Name)
		if podUnschedulableCheck(pod) {
			detectProblem(ctx, &pod, &v1.ContainerStatus{}, problemInput, &problems)
			continue
		}

		golog.Printf("INFO - Checking init container with pod %s", pod.Name)
		for _, status := range pod.Status.InitContainerStatuses {
			if detectProblem(ctx, &pod, &status, problemInput, &problems) {
				continue
			}
		}

		golog.Printf("INFO - Checking container with pod %s", pod.Name)
		for _, status := range pod.Status.ContainerStatuses {
			if detectProblem(ctx, &pod, &status, problemInput, &problems) {
				continue
			}
		}
	}
	return problems, nil
}

func checkPossibleErrorMessage(event *string, key *string, input *ProblemInput) bool {
	if len(input.PossibleErrorMessages) == 0 {
		return false
	}
	if msgs, ok := input.PossibleErrorMessages[*key]; ok {
		for _, msg := range msgs {
			if strings.Contains(strings.ToLower(*event), strings.ToLower(msg)) {
				return true
			}
		}
	}
	return false
}

func detectProblem(ctx context.Context, pod *v1.Pod, status *v1.ContainerStatus, problemInput *ProblemInput,
	problems *[]problem.Problem) bool {
	hasFoundRootCause := false
	if problemInput.PodDetect(*status) {
		if status.State.Waiting != nil {
			golog.Printf("INFO - Found %s with pod %s", status.State.Waiting.Reason, pod.GetName())
		}
		if status.State.Terminated != nil {
			golog.Printf("INFO - Found %s with pod %s", status.State.Terminated.Reason, pod.GetName())
		}
		if *problemInput.CheckEvent {
			golog.Println("INFO - Event client configured, checking details in events.")
			timespan := DefaultTimespan
			if problemInput.Detector.DetectorInput.EventTimespan != (problem.TimeSpan{}) {
				timespan = problemInput.Detector.DetectorInput.EventTimespan
			}
			retriever := problemInput.Detector.DetectorInput.EventRetriever
			filter := CreateEventFilterCriteria(timespan, retriever.AddFilters(pod.Name, pod.Namespace))
			eventDataRef := retriever.Retrieve(filter)
			events, err := eventDataRef.GetEvents(ctx)
			if err != nil {
				golog.Printf("ERROR - Got error when calling Datadog event API, error is %s", err)
			}
			for _, event := range events {
				for msg := range problemInput.SolutionsLinks {
					if strings.Contains(strings.ToLower(event.Message), strings.ToLower(msg)) ||
						checkPossibleErrorMessage(&event.Message, &msg, problemInput) {
						hasFoundRootCause = true
						golog.Printf("INFO - Found event with error '%s', pod %s", msg, pod.Name)
						existingProblem := getProblemByMsg(msg, problems)
						addToProblem(pod, status, existingProblem, &msg, problems, problemInput, &event)
					}
				}
			}
			if !hasFoundRootCause {
				golog.Printf("INFO - Can not find event details for pod %s", pod.Name)
				addProblemWithExitReason(pod, status, problems, problemInput)
			}
		} else {
			golog.Printf("INFO - Event client did not configured, can not analysis root cause, pod %s", pod.Name)
			addProblemWithExitReason(pod, status, problems, problemInput)
		}
	}
	return hasFoundRootCause
}

// Analysis detailed message in event if it presents, return solution.
func getSolutionLinks(pod *v1.Pod, status *v1.ContainerStatus, problemInput *ProblemInput, msg *string,
	e *observability.EventRecord) ([]string, map[problem.DeeplinkType]*url.URL) {
	var nextSteps []string
	var deeplinks map[problem.DeeplinkType]*url.URL
	if solution, ok := problemInput.SolutionsLinks[*msg]; ok {
		if solution != nil {
			nextSteps, deeplinks = solution(pod, status, problemInput, msg, e)
		} else {
			nextSteps, deeplinks = getSolutionUnknown(pod, status, problemInput, msg)
		}
	} else {
		nextSteps, deeplinks = getSolutionUnknown(pod, status, problemInput, msg)
	}
	return nextSteps, deeplinks
}

func getEventLink(pod *v1.Pod, problemInput *ProblemInput, startTime time.Time,
	endTime time.Time) *url.URL {
	var url *url.URL
	var err error
	// The cluster name and namespace may be empty in the v1.Pod, use them from DectectorInput.
	clusterName := problemInput.Detector.DetectorInput.ClusterName
	namespace := problemInput.Detector.DetectorInput.Namespace
	eventDeeplinkRetriever := problemInput.Detector.DetectorInput.EventDeeplinkRetriever
	if eventDeeplinkRetriever != nil {
		eventsLink := eventDeeplinkRetriever.GetEventDeepLink(EventPodQueryType, clusterName, namespace,
			pod.Name, startTime, endTime)
		if eventsLink != "" {
			url, err = url.Parse(eventsLink)
		}
		if err != nil {
			golog.Printf("WARN - Event url generation failed %s", err)
		}
	}
	return url

}

func getLogLink(pod *v1.Pod, problemInput *ProblemInput, isKubeletLog bool,
	kubeletLogSearchByNamespace bool, startTime time.Time, endTime time.Time) *url.URL {
	var url *url.URL
	var err error
	clusterName := problemInput.Detector.DetectorInput.ClusterName
	namespace := problemInput.Detector.DetectorInput.Namespace
	logDeeplinkRetriever := problemInput.Detector.DetectorInput.LogDeeplinkRetriever
	if logDeeplinkRetriever != nil {
		logsLink := logDeeplinkRetriever.GetLogDeepLink(
			clusterName, namespace, pod.Name, isKubeletLog, kubeletLogSearchByNamespace, startTime, endTime)
		if logsLink != "" {
			url, err = url.Parse(logsLink)
		}
		if err != nil {
			golog.Printf("WARN - Log url generation failed %s", err)
		}
	}
	return url
}

func getLogEventLinks(pod *v1.Pod, problemInput *ProblemInput, kubeletLog bool,
	kubeletLogSearchByNamespace bool, appLog bool, event *observability.EventRecord) map[problem.DeeplinkType]*url.URL {
	linkMap := map[problem.DeeplinkType]*url.URL{}
	var startTime, endTime time.Time
	if event == nil {
		startTime = pod.ObjectMeta.CreationTimestamp.Time
		endTime = time.Now()
	} else {
		startTime = SetStartTime(event.DateHappend, EventLogTimespan)
		endTime = SetEndTime(event.DateHappend, EventLogTimespan)
	}
	linkMap[problem.DeeplinkEvent] = getEventLink(pod, problemInput, startTime, endTime)
	if kubeletLog {
		linkMap[problem.DeeplinkKubeletLog] = getLogLink(pod, problemInput, true, kubeletLogSearchByNamespace, startTime, endTime)
	}
	if appLog {
		linkMap[problem.DeeplinkAppLog] = getLogLink(pod, problemInput, false, false, startTime, endTime)
	}
	return linkMap
}

// Add the problem to existing problem if they have same root cause, otherwise create a new one.
func addToProblem(pod *v1.Pod, status *v1.ContainerStatus, prob *problem.Problem, msg *string,
	problems *[]problem.Problem, problemInput *ProblemInput, e *observability.EventRecord) {
	if prob == nil || isEmptyProblem(*prob) {
		affectedResources := make(map[string]problem.ResourceDetails)
		solutions, deeplinks := getSolutionLinks(pod, status, problemInput, msg, e)

		affectedResources[pod.Name] = problem.ResourceDetails{
			Deeplink:  deeplinks,
			Resource:  pod.DeepCopyObject(),
			NextSteps: solutions,
		}

		doc, err := url.Parse(problemInput.DocLink)
		if err != nil {
			golog.Printf("WARN - error occurred creating Problem.Docs, error is %s", err)
		}

		prob = &problem.Problem{
			DomainName:        problemInput.Detector.Domain(),
			Name:              problemInput.Title,
			Description:       *msg,
			Tags:              problemInput.Tags,
			Docs:              []*url.URL{doc},
			AffectedResources: affectedResources,
			Level:             problem.UserNamespace,
		}

		*problems = append(*problems, *prob)
	} else {
		p := getProblemByMsg(*msg, problems)
		solutions, deeplinks := getSolutionLinks(pod, status, problemInput, msg, e)
		p.AffectedResources[pod.Name] = problem.ResourceDetails{
			Deeplink:  deeplinks,
			Resource:  pod.DeepCopyObject(),
			NextSteps: solutions,
		}
	}
}

func addProblemWithExitReason(pod *v1.Pod, status *v1.ContainerStatus, problems *[]problem.Problem, problemInput *ProblemInput) {
	var rootCauseByExitCode string
	if problemInput.Title == CrashLoopBackOff {
		rootCauseByExitCode = getRootCauseByExitCode(pod)
	}
	if rootCauseByExitCode != "" {
		existingProblem := getProblemByMsg(rootCauseByExitCode, problems)
		addToProblem(pod, status, existingProblem, &rootCauseByExitCode, problems, problemInput, nil)
	} else {
		existingProblem := getProblemByMsg(GeneralMsg, problems)
		addToProblem(pod, status, existingProblem, &GeneralMsg, problems, problemInput, nil)
	}
}

func getContainerName(pod *v1.Pod, reason string) string {
	if pod == nil {
		return ""
	}
	for _, status := range pod.Status.ContainerStatuses {
		if status.State.Waiting != nil && status.State.Waiting.Reason == reason {
			return status.Name
		}
	}
	return ""
}

func getSolutionUnknown(pod *v1.Pod, status *v1.ContainerStatus, problemInput *ProblemInput,
	msg *string) ([]string, map[problem.DeeplinkType]*url.URL) {
	return []string{"Unknown root cause."}, getLogEventLinks(pod, problemInput, true, false, false, nil)
}

// Get the problem from problem slice by Problem.Description
func getProblemByMsg(msg string, problems *[]problem.Problem) *problem.Problem {
	for _, problem := range *problems {
		if msg == string(problem.Description) {
			return &problem
		}
	}
	return nil
}

func isEmptyProblem(p problem.Problem) bool {
	return len(p.Name) == 0
}

type PodSkipDetect func(pod v1.Pod) bool

type PodDetect func(status v1.ContainerStatus) bool

// Create event.FilterCriteria.
func CreateEventFilterCriteria(timespan problem.TimeSpan, filterCriteria map[string]string) observability.EventFilterCriteria {
	now := time.Now()
	return observability.EventFilterCriteria{
		StartTime:      SetStartTime(now, timespan),
		EndTime:        now,
		FilterCriteria: filterCriteria,
	}
}

func SetStartTime(currentTime time.Time, timespan problem.TimeSpan) time.Time {
	return currentTime.Add(time.Duration(timespan.Timespan) * -timespan.TimespanType)
}

func SetEndTime(currentTime time.Time, timespan problem.TimeSpan) time.Time {
	return currentTime.Add(time.Duration(timespan.Timespan) * timespan.TimespanType)
}

func MaskString(str string, length ...int) string {
	n := MaskedVisibleNo
	if len(length) > 0 {
		n = length[0]
	}
	if len(str) <= n {
		return strings.Repeat(MaskedString, MaskedPrefixNo)
	} else {
		mask := strings.Repeat(MaskedString, MaskedPrefixNo)
		return fmt.Sprint(mask, str[len(str)-n:])
	}
}

// A general function used to parse go template.
// Go template passed in string type, parsed results returned in []string type.
// Parameter splitIt, if true, parsed results will be split by \n.
func GetSolutionsByTemplate(template string, object interface{}, splitIt bool) (solution []string) {
	solution = []string{}
	t, err := solutionTemp.Parse(template)
	if err != nil {
		return
	}
	var tpl bytes.Buffer
	err = t.Execute(&tpl, object)
	if err != nil {
		return
	}
	s := tpl.String()
	s1 := strings.TrimPrefix(strings.TrimSuffix(s, "\n"), "\n")
	if splitIt {
		solution = strings.Split(s1, "\n")
	} else {
		solution = append(solution, s1)
	}
	return
}
