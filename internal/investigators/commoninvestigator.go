/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package investigators

import (
	"bytes"
	"context"
	"strings"
	"text/template"

	com "github.com/fidelity/theliv/pkg/common"
	log "github.com/fidelity/theliv/pkg/log"

	"github.com/fidelity/theliv/internal/problem"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	networkv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func CommonInvestigator(ctx context.Context, problem *problem.Problem, input *problem.DetectorCreationInput) {
	switch problem.Tags[com.Resourcetype] {
	case com.Pod:
		loadPodDetails(problem)
	case com.Container:
		loadContainerDetails(problem)
	case com.Deployment:
		loadDeploymentDetails(problem)
	case com.Replicaset:
		loadReplicaSetDetails(problem)
	case com.Statefulset:
		loadStatefulSetDetails(problem)
	case com.Daemonset:
		loadDaemonSetDetails(problem)
	case com.Node:
		loadNodeDetails(problem)
	case com.Job:
		loadJobDetails(problem)
	case com.Cronjob:
		loadCronJobDetails(problem)
	case com.Service:
		loadServiceDetails(problem)
	case com.Ingress:
		loadIngressDetails(problem)
	case com.Endpoint:
		loadEndpointsDetails(problem)
	default:
		log.S().Warnf("Not found investigator function for resource type %s", problem.Tags[com.Resourcetype])
	}
}

func loadPodDetails(problem *problem.Problem) {
	var ro runtime.Object = problem.AffectedResources.Resource
	pod := *ro.(*v1.Pod)
	logChecking(com.Pod + com.Blank + pod.Name)
	for _, condition := range pod.Status.Conditions {
		appendNonEmptyDetail(problem, string(condition.Type), string(condition.Status),
			condition.Message, condition.Reason)
	}
}

func loadContainerDetails(problem *problem.Problem) {
	var ro runtime.Object = problem.AffectedResources.Resource
	pod := *ro.(*v1.Pod)
	containername := problem.Tags[com.Container]
	logChecking("init container with " + com.Pod + com.Blank + pod.Name)
	for _, status := range pod.Status.InitContainerStatuses {
		if status.Name == containername {
			if status.State.Terminated != nil {
				appendDetail(problem, "", status.State.Terminated.Message,
					status.State.Terminated.Reason)
			}
			if status.State.Waiting != nil {
				appendDetail(problem, "", status.State.Waiting.Message,
					status.State.Waiting.Reason)
			}
			break
		}
	}

	logChecking("container with " + com.Pod + com.Blank + pod.Name)
	for _, status := range pod.Status.ContainerStatuses {
		if status.Name == containername {
			if status.State.Terminated != nil {
				appendDetail(problem, "", status.State.Terminated.Message,
					status.State.Terminated.Reason)
			}
			if status.State.Waiting != nil {
				appendDetail(problem, "", status.State.Waiting.Message,
					status.State.Waiting.Reason)
			}
			break
		}
	}
}

func loadDeploymentDetails(problem *problem.Problem) {
	var ro runtime.Object = problem.AffectedResources.Resource
	deployment := *ro.(*appsv1.Deployment)
	logChecking(com.Deployment + com.Blank + deployment.Name)
	for _, condition := range deployment.Status.Conditions {
		appendNonEmptyDetail(problem, string(condition.Type), string(condition.Status),
			condition.Message, condition.Reason)
	}
}

func loadReplicaSetDetails(problem *problem.Problem) {
	var ro runtime.Object = problem.AffectedResources.Resource
	rs := *ro.(*appsv1.ReplicaSet)
	logChecking(com.Replicaset + com.Blank + rs.Name)
	for _, condition := range rs.Status.Conditions {
		appendNonEmptyDetail(problem, string(condition.Type), string(condition.Status),
			condition.Message, condition.Reason)
	}
}

func loadStatefulSetDetails(problem *problem.Problem) {
	var ro runtime.Object = problem.AffectedResources.Resource
	ss := *ro.(*appsv1.StatefulSet)
	logChecking(com.Statefulset + com.Blank + ss.Name)
	for _, condition := range ss.Status.Conditions {
		appendNonEmptyDetail(problem, string(condition.Type), string(condition.Status),
			condition.Message, condition.Reason)
	}
}

func loadDaemonSetDetails(problem *problem.Problem) {
	var ro runtime.Object = problem.AffectedResources.Resource
	ds := *ro.(*appsv1.DaemonSet)
	logChecking(com.Daemonset + com.Blank + ds.Name)
	for _, condition := range ds.Status.Conditions {
		appendNonEmptyDetail(problem, string(condition.Type), string(condition.Status),
			condition.Message, condition.Reason)
	}
}

func loadNodeDetails(problem *problem.Problem) {
	var ro runtime.Object = problem.AffectedResources.Resource
	node := *ro.(*v1.Node)
	logChecking(com.Node + com.Blank + node.Name)
	for _, condition := range node.Status.Conditions {
		appendNonEmptyDetail(problem, string(condition.Type), string(condition.Status),
			condition.Message, condition.Reason)
	}
}

func loadJobDetails(problem *problem.Problem) {
	var ro runtime.Object = problem.AffectedResources.Resource
	job := *ro.(*batchv1.Job)
	logChecking(com.Job + com.Blank + job.Name)
	for _, condition := range job.Status.Conditions {
		appendNonEmptyDetail(problem, string(condition.Type), string(condition.Status),
			condition.Message, condition.Reason)
	}
}

func loadCronJobDetails(problem *problem.Problem) {
	var ro runtime.Object = problem.AffectedResources.Resource
	job := *ro.(*batchv1.CronJob)
	logChecking(com.Cronjob + com.Blank + job.Name)
	for _, job := range job.Status.Active {
		if job.Name != "" && job.Namespace != "" {
			detail := job.Name + "in " + job.Namespace + "is active."
			problem.SolutionDetails = append(problem.SolutionDetails, detail)
		}
	}
}

func loadServiceDetails(problem *problem.Problem) {
	var ro runtime.Object = problem.AffectedResources.Resource
	service := *ro.(*v1.Service)
	logChecking(com.Service + com.Blank + service.Name)
	for _, condition := range service.Status.Conditions {
		appendNonEmptyDetail(problem, string(condition.Type), string(condition.Status),
			condition.Message, condition.Reason)
	}
}

func loadIngressDetails(problem *problem.Problem) {
	var ro runtime.Object = problem.AffectedResources.Resource
	ingress := *ro.(*networkv1.Ingress)
	logChecking(com.Ingress + com.Blank + ingress.Name)
	for _, ingress := range ingress.Status.LoadBalancer.Ingress {
		for _, port := range ingress.Ports {
			if port.Error != nil {
				detail := *port.Error
				problem.SolutionDetails = append(problem.SolutionDetails, detail)
			}
		}
	}
}

func loadEndpointsDetails(problem *problem.Problem) {
	var ro runtime.Object = problem.AffectedResources.Resource
	endpoints := *ro.(*v1.Endpoints)
	logChecking(com.Endpoint + com.Blank + endpoints.Name)
	for _, sub := range endpoints.Subsets {
		if sub.NotReadyAddresses != nil {
			for _, addr := range sub.NotReadyAddresses {
				detail := addr.IP + "not ready. Target object is " + addr.TargetRef.Name
				problem.SolutionDetails = append(problem.SolutionDetails, detail)
			}
		}
	}
}

func buildReasonMsg(reason string, message string) string {
	var detail string
	if reason != "" && message != "" {
		detail = reason + ":" + message
	} else if message != "" {
		detail = message
	} else if reason != "" {
		detail = reason
	}
	return detail
}

// A general template instance.
var solutionTemp = template.New("solutionTemp")

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

func appendDetail(problem *problem.Problem, detail string,
	msg string, reason string) {
	detail = detail + buildReasonMsg(reason, msg)
	problem.SolutionDetails = append(problem.SolutionDetails, detail)
}

func appendNonEmptyDetail(problem *problem.Problem, conType string,
	conMsg, msg string, reason string) {
	if msg != "" || reason != "" {
		detail := conType + "=" + conMsg + ". "
		appendDetail(problem, detail, msg, reason)
	}
}

func logChecking(res string) {
	log.S().Infof("Checking status with %s", res)
}
