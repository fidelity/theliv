/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package investigators

import (
	"context"

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
		loadPodDetails(ctx, problem)
	case com.Container:
		loadContainerDetails(ctx, problem)
	case com.Initcontainer:
		loadContainerDetails(ctx, problem)
	case com.Deployment:
		loadDeploymentDetails(ctx, problem)
	case com.Replicaset:
		loadReplicaSetDetails(ctx, problem)
	case com.Statefulset:
		loadStatefulSetDetails(ctx, problem)
	case com.Daemonset:
		loadDaemonSetDetails(ctx, problem)
	case com.Node:
		loadNodeDetails(ctx, problem)
	case com.Job:
		loadJobDetails(ctx, problem)
	case com.Cronjob:
		loadCronJobDetails(ctx, problem)
	case com.Service:
		loadServiceDetails(ctx, problem)
	case com.Ingress:
		loadIngressDetails(ctx, problem)
	case com.Endpoint:
		loadEndpointsDetails(ctx, problem)
	default:
		log.SWithContext(ctx).Warnf("Not found investigator function for resource type %s", problem.Tags[com.Resourcetype])
	}
}

func loadPodDetails(ctx context.Context, problem *problem.Problem) {
	var ro runtime.Object = problem.AffectedResources.Resource
	pod := *ro.(*v1.Pod)
	logChecking(ctx, com.Pod + com.Blank + pod.Name)
	for _, condition := range pod.Status.Conditions {
		appendNonEmptyDetail(problem, string(condition.Type), string(condition.Status),
			condition.Message, condition.Reason)
	}
}

func loadContainerDetails(ctx context.Context, problem *problem.Problem) {
	var ro runtime.Object = problem.AffectedResources.Resource
	pod := *ro.(*v1.Pod)
	containerName := problem.Tags[com.Container]
	logChecking(ctx, "init container with " + com.Pod + com.Blank + pod.Name)
	for _, status := range pod.Status.InitContainerStatuses {
		if status.Name == containerName {
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

	logChecking(ctx, "container with " + com.Pod + com.Blank + pod.Name)
	for _, status := range pod.Status.ContainerStatuses {
		if status.Name == containerName {
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

func loadDeploymentDetails(ctx context.Context, problem *problem.Problem) {
	var ro runtime.Object = problem.AffectedResources.Resource
	deployment := *ro.(*appsv1.Deployment)
	logChecking(ctx, com.Deployment + com.Blank + deployment.Name)
	for _, condition := range deployment.Status.Conditions {
		appendNonEmptyDetail(problem, string(condition.Type), string(condition.Status),
			condition.Message, condition.Reason)
	}
}

func loadReplicaSetDetails(ctx context.Context, problem *problem.Problem) {
	var ro runtime.Object = problem.AffectedResources.Resource
	rs := *ro.(*appsv1.ReplicaSet)
	logChecking(ctx, com.Replicaset + com.Blank + rs.Name)
	for _, condition := range rs.Status.Conditions {
		appendNonEmptyDetail(problem, string(condition.Type), string(condition.Status),
			condition.Message, condition.Reason)
	}
}

func loadStatefulSetDetails(ctx context.Context, problem *problem.Problem) {
	var ro runtime.Object = problem.AffectedResources.Resource
	ss := *ro.(*appsv1.StatefulSet)
	logChecking(ctx, com.Statefulset + com.Blank + ss.Name)
	for _, condition := range ss.Status.Conditions {
		appendNonEmptyDetail(problem, string(condition.Type), string(condition.Status),
			condition.Message, condition.Reason)
	}
}

func loadDaemonSetDetails(ctx context.Context, problem *problem.Problem) {
	var ro runtime.Object = problem.AffectedResources.Resource
	ds := *ro.(*appsv1.DaemonSet)
	logChecking(ctx, com.Daemonset + com.Blank + ds.Name)
	for _, condition := range ds.Status.Conditions {
		appendNonEmptyDetail(problem, string(condition.Type), string(condition.Status),
			condition.Message, condition.Reason)
	}
}

func loadNodeDetails(ctx context.Context, problem *problem.Problem) {
	var ro runtime.Object = problem.AffectedResources.Resource
	node := *ro.(*v1.Node)
	logChecking(ctx, com.Node + com.Blank + node.Name)
	for _, condition := range node.Status.Conditions {
		appendNonEmptyDetail(problem, string(condition.Type), string(condition.Status),
			condition.Message, condition.Reason)
	}
}

func loadJobDetails(ctx context.Context, problem *problem.Problem) {
	var ro runtime.Object = problem.AffectedResources.Resource
	job := *ro.(*batchv1.Job)
	logChecking(ctx, com.Job + com.Blank + job.Name)
	for _, condition := range job.Status.Conditions {
		appendNonEmptyDetail(problem, string(condition.Type), string(condition.Status),
			condition.Message, condition.Reason)
	}
}

func loadCronJobDetails(ctx context.Context, problem *problem.Problem) {
	var ro runtime.Object = problem.AffectedResources.Resource
	job := *ro.(*batchv1.CronJob)
	logChecking(ctx, com.Cronjob + com.Blank + job.Name)
	for _, job := range job.Status.Active {
		if job.Name != "" && job.Namespace != "" {
			detail := job.Name + "in " + job.Namespace + "is active."
			problem.SolutionDetails = append(problem.SolutionDetails, detail)
		}
	}
}

func loadServiceDetails(ctx context.Context, problem *problem.Problem) {
	var ro runtime.Object = problem.AffectedResources.Resource
	service := *ro.(*v1.Service)
	logChecking(ctx, com.Service + com.Blank + service.Name)
	for _, condition := range service.Status.Conditions {
		appendNonEmptyDetail(problem, string(condition.Type), string(condition.Status),
			condition.Message, condition.Reason)
	}
}

func loadIngressDetails(ctx context.Context, problem *problem.Problem) {
	var ro runtime.Object = problem.AffectedResources.Resource
	ingress := *ro.(*networkv1.Ingress)
	logChecking(ctx, com.Ingress + com.Blank + ingress.Name)
	for _, ingress := range ingress.Status.LoadBalancer.Ingress {
		for _, port := range ingress.Ports {
			if port.Error != nil {
				detail := *port.Error
				appendSolution(problem, detail)
			}
		}
	}
}

func loadEndpointsDetails(ctx context.Context, problem *problem.Problem) {
	var ro runtime.Object = problem.AffectedResources.Resource
	endpoints := *ro.(*v1.Endpoints)
	logChecking(ctx, com.Endpoint + com.Blank + endpoints.Name)
	for _, sub := range endpoints.Subsets {
		if sub.NotReadyAddresses != nil {
			for _, addr := range sub.NotReadyAddresses {
				detail := addr.IP + "not ready. Target object is " + addr.TargetRef.Name
				appendSolution(problem, detail)
			}
		}
	}
}

func buildReasonMsg(reason string, message string) string {
	var detail string
	if reason != "" && message != "" {
		detail = "Reason: " + reason + " Message: " + message
	} else if message != "" {
		detail = "Message: " + message
	} else if reason != "" {
		detail = "Reason: " + reason
	}
	return detail
}

func appendDetail(problem *problem.Problem, detail string,
	msg string, reason string) {
	detail = detail + buildReasonMsg(reason, msg)
	appendSolution(problem, detail)
}

func appendNonEmptyDetail(problem *problem.Problem, conType string,
	conMsg, msg string, reason string) {
	if msg != "" || reason != "" {
		detail := "Found Status: " + conType + "=" + conMsg + ". "
		appendDetail(problem, detail, msg, reason)
	}
}
