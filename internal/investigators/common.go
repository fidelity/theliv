/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package investigators

import (
	"bytes"
	"context"
	"regexp"
	"strings"
	"text/template"
	"time"

	log "github.com/fidelity/theliv/pkg/log"
	"github.com/fidelity/theliv/pkg/observability"
	v1 "k8s.io/api/core/v1"

	"github.com/fidelity/theliv/internal/problem"
)

// Default Timespan, used in Event Filtering.
var DefaultTimespan = problem.TimeSpan{
	Timespan:     48,
	TimespanType: time.Hour,
}

// A general template instance.
var solutionTemp = template.New("solutionTemp")

// Create event.FilterCriteria.
func CreateEventFilterCriteria(timespan problem.TimeSpan,
	filterCriteria map[string]string) observability.EventFilterCriteria {

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

func getPodSolutionFromEvents(ctx context.Context, problem *problem.Problem,
	input *problem.DetectorCreationInput,
	pod *v1.Pod, status *v1.ContainerStatus,
	solutions map[string]func(pod *v1.Pod, status *v1.ContainerStatus) []string) string {

	events, err := GetPodEvents(ctx, input, pod)
	if err != nil {
		log.SWithContext(ctx).Error("Got error when calling Kubernetes event API, error is %s", err)
	}

	if len(events) > 0 {
		for _, event := range events {
			for msg := range solutions {
				matched, err := regexp.MatchString(strings.ToLower(msg), strings.ToLower(event.Message))
				if err == nil && matched {
					log.SWithContext(ctx).Infof("Found event with error '%s', pod %s, container %s", msg, pod.Name, status.Name)
					addSolutionFromMap(problem, pod, status, msg, solutions)
					return msg
				}
			}
		}
	}

	log.SWithContext(ctx).Infof("Can not find event details for pod %s, container %s", pod.Name, status.Name)

	return ""
}

func GetPodEvents(ctx context.Context, input *problem.DetectorCreationInput, pod *v1.Pod) ([]observability.EventRecord, error) {

	filter := CreateEventFilterCriteria(DefaultTimespan, input.EventRetriever.AddFilters(pod.Name, pod.Namespace))
	eventDataRef := input.EventRetriever.Retrieve(filter)
	return eventDataRef.GetEvents(ctx)
}

func addSolutionFromMap(problem *problem.Problem, pod *v1.Pod, status *v1.ContainerStatus, msg string,
	solutions map[string]func(pod *v1.Pod, status *v1.ContainerStatus) []string) {

	appendSolution(problem, solutions[msg](pod, status))
}

// A general function used to parse go template.
// Go template passed in string type, parsed results returned in []string type.
// Parameter splitIt, if true, parsed results will be split by \n.
func GetSolutionsByTemplate(ctx context.Context, template string, object interface{}, splitIt bool) (solution []string) {
	solution = []string{}
	s, err := ExecGoTemplate(ctx, template, object)
	if err != nil {
		return
	}
	s1 := strings.TrimPrefix(strings.TrimSuffix(s, "\n"), "\n")
	if splitIt {
		solution = strings.Split(s1, "\n")
	} else {
		solution = append(solution, s1)
	}
	return
}

// Execute Go Template parse
func ExecGoTemplate(ctx context.Context, template string, object interface{}) (s string, err error) {
	t, err := solutionTemp.Parse(template)
	if err != nil {
		log.SWithContext(ctx).Errorf("Parse template got error: %s", err)
		return
	}
	var tpl bytes.Buffer
	err = t.Execute(&tpl, object)
	if err != nil {
		log.SWithContext(ctx).Errorf("Parse template with object got error: %s", err)
		return
	}
	s = tpl.String()
	return
}

func logChecking(ctx context.Context, res string) {
	log.SWithContext(ctx).Infof("Checking status with %s", res)
}

func appendSolution(problem *problem.Problem, solutions interface{}) {
	switch v := solutions.(type) {
	case string:
		problem.SolutionDetails = append(problem.SolutionDetails, v)
	case []string:
		problem.SolutionDetails = append(problem.SolutionDetails, v...)
	}
}
