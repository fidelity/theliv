/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package investigators

import (
	"context"
	golog "log"

	"github.com/fidelity/theliv/internal/problem"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	networkv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func CommonInvestigator(ctx context.Context, problem *problem.Problem, input *problem.DetectorCreationInput) {
	switch problem.Tags["resourcetype"] {
	case "pod":
		loadPodDetails(problem)
	case "container":
		loadContainerDetails(problem)
	case "deployment":
		loadDeploymentDetails(problem)
	case "replicaset":
		loadReplicaSetDetails(problem)
	case "statefulset":
		loadStatefulSetDetails(problem)
	case "daemonset":
		loadDaemonSetDetails(problem)
	case "node":
		loadNodeDetails(problem)
	case "job":
		loadJobDetails(problem)
	case "cronjob":
		loadCronJobDetails(problem)
	case "service":
		loadServiceDetails(problem)
	case "ingress":
		loadIngressDetails(problem)
	case "endpoint":
		loadEndpointsDetails(problem)
	default:
		golog.Printf("WARN - Not found investigator function for resource type %s", problem.Tags["resourcetype"])
	}
}

func loadPodDetails(problem *problem.Problem) {
	var ro runtime.Object = problem.AffectedResources.Resource
	pod := *ro.(*v1.Pod)
	golog.Printf("INFO - Checking status with pod %s", pod.Name)
	if pod.Status.Message == "" && pod.Status.Reason == "" {
		return
	}
	detail := buildReasonMsg(pod.Status.Reason, pod.Status.Message)
	problem.SolutionDetails = append(problem.SolutionDetails, &detail)
}

func loadContainerDetails(problem *problem.Problem) {
	var ro runtime.Object = problem.AffectedResources.Resource
	pod := *ro.(*v1.Pod)
	containername := problem.Tags["container"]
	golog.Printf("INFO - Checking init container with pod %s", pod.Name)
	for _, status := range pod.Status.InitContainerStatuses {
		if status.Name == containername {
			if status.State.Terminated != nil {
				detail := buildReasonMsg(status.State.Terminated.Reason, status.State.Terminated.Message)
				problem.SolutionDetails = append(problem.SolutionDetails, &detail)
			}
			if status.State.Waiting != nil {
				detail := buildReasonMsg(status.State.Waiting.Reason, status.State.Waiting.Message)
				problem.SolutionDetails = append(problem.SolutionDetails, &detail)
			}
			break
		}
	}

	golog.Printf("INFO - Checking container with pod %s", pod.Name)
	for _, status := range pod.Status.ContainerStatuses {
		if status.Name == containername {
			if status.State.Terminated != nil {
				detail := buildReasonMsg(status.State.Terminated.Reason, status.State.Terminated.Message)
				problem.SolutionDetails = append(problem.SolutionDetails, &detail)
			}
			if status.State.Waiting != nil {
				detail := buildReasonMsg(status.State.Waiting.Reason, status.State.Waiting.Message)
				problem.SolutionDetails = append(problem.SolutionDetails, &detail)
			}
			break
		}
	}
}

func loadDeploymentDetails(problem *problem.Problem) {
	var ro runtime.Object = problem.AffectedResources.Resource
	deployment := *ro.(*appsv1.Deployment)
	golog.Printf("INFO - Checking status with deployment %s", deployment.Name)
	for _, condition := range deployment.Status.Conditions {
		detail := string(condition.Type) + "=" + string(condition.Status)
		if condition.Message != "" || condition.Reason != "" {
			detail = detail + buildReasonMsg(condition.Reason, condition.Message)
			problem.SolutionDetails = append(problem.SolutionDetails, &detail)
		}
	}
}

func loadReplicaSetDetails(problem *problem.Problem) {
	var ro runtime.Object = problem.AffectedResources.Resource
	rs := *ro.(*appsv1.ReplicaSet)
	golog.Printf("INFO - Checking status with replicaset %s", rs.Name)
	for _, condition := range rs.Status.Conditions {
		detail := string(condition.Type) + "=" + string(condition.Status)
		if condition.Message != "" || condition.Reason != "" {
			detail = detail + buildReasonMsg(condition.Reason, condition.Message)
			problem.SolutionDetails = append(problem.SolutionDetails, &detail)
		}
	}
}

func loadStatefulSetDetails(problem *problem.Problem) {
	var ro runtime.Object = problem.AffectedResources.Resource
	ss := *ro.(*appsv1.StatefulSet)
	golog.Printf("INFO - Checking status with statefulset %s", ss.Name)
	for _, condition := range ss.Status.Conditions {
		detail := string(condition.Type) + "=" + string(condition.Status)
		if condition.Message != "" || condition.Reason != "" {
			detail = detail + buildReasonMsg(condition.Reason, condition.Message)
			problem.SolutionDetails = append(problem.SolutionDetails, &detail)
		}
	}
}

func loadDaemonSetDetails(problem *problem.Problem) {
	var ro runtime.Object = problem.AffectedResources.Resource
	ds := *ro.(*appsv1.DaemonSet)
	golog.Printf("INFO - Checking status with daemonset %s", ds.Name)
	for _, condition := range ds.Status.Conditions {
		detail := string(condition.Type) + "=" + string(condition.Status)
		if condition.Message != "" || condition.Reason != "" {
			detail = detail + buildReasonMsg(condition.Reason, condition.Message)
			problem.SolutionDetails = append(problem.SolutionDetails, &detail)
		}
	}
}

func loadNodeDetails(problem *problem.Problem) {
	var ro runtime.Object = problem.AffectedResources.Resource
	node := *ro.(*v1.Node)
	golog.Printf("INFO - Checking status with node %s", node.Name)
	for _, condition := range node.Status.Conditions {
		detail := string(condition.Type) + "=" + string(condition.Status)
		if condition.Message != "" || condition.Reason != "" {
			detail = detail + buildReasonMsg(condition.Reason, condition.Message)
			problem.SolutionDetails = append(problem.SolutionDetails, &detail)
		}
	}
}

func loadJobDetails(problem *problem.Problem) {
	var ro runtime.Object = problem.AffectedResources.Resource
	job := *ro.(*batchv1.Job)
	golog.Printf("INFO - Checking status with job %s", job.Name)
	for _, condition := range job.Status.Conditions {
		detail := string(condition.Type) + "=" + string(condition.Status)
		if condition.Message != "" || condition.Reason != "" {
			detail = detail + buildReasonMsg(condition.Reason, condition.Message)
			problem.SolutionDetails = append(problem.SolutionDetails, &detail)
		}
	}
}

func loadCronJobDetails(problem *problem.Problem) {
	var ro runtime.Object = problem.AffectedResources.Resource
	job := *ro.(*batchv1.CronJob)
	golog.Printf("INFO - Checking status with cron job %s", job.Name)
	for _, job := range job.Status.Active {
		if job.Name != "" && job.Namespace != "" {
			detail := job.Name + "in " + job.Namespace + "is active."
			problem.SolutionDetails = append(problem.SolutionDetails, &detail)
		}
	}
}

func loadServiceDetails(problem *problem.Problem) {
	var ro runtime.Object = problem.AffectedResources.Resource
	service := *ro.(*v1.Service)
	golog.Printf("INFO - Checking status with service %s", service.Name)
	for _, condition := range service.Status.Conditions {
		detail := string(condition.Type) + "=" + string(condition.Status)
		if condition.Message != "" || condition.Reason != "" {
			detail = detail + buildReasonMsg(condition.Reason, condition.Message)
			problem.SolutionDetails = append(problem.SolutionDetails, &detail)
		}
	}
}

func loadIngressDetails(problem *problem.Problem) {
	var ro runtime.Object = problem.AffectedResources.Resource
	ingress := *ro.(*networkv1.Ingress)
	golog.Printf("INFO - Checking status with ingress %s", ingress.Name)
	for _, ingress := range ingress.Status.LoadBalancer.Ingress {
		for _, port := range ingress.Ports {
			if port.Error != nil {
				detail := *port.Error
				problem.SolutionDetails = append(problem.SolutionDetails, &detail)
			}
		}
	}
}

func loadEndpointsDetails(problem *problem.Problem) {
	var ro runtime.Object = problem.AffectedResources.Resource
	endpoints := *ro.(*v1.Endpoints)
	golog.Printf("INFO - Checking status with endpoints %s", endpoints.Name)
	for _, sub := range endpoints.Subsets {
		if sub.NotReadyAddresses != nil {
			for _, addr := range sub.NotReadyAddresses {
				detail := addr.IP + "not ready. Target object is " + addr.TargetRef.Name
				problem.SolutionDetails = append(problem.SolutionDetails, &detail)
			}
		}
	}
}

func buildReasonMsg(reason string, message string) string {
	detail := ""
	if reason != "" {
		detail = detail + "Reason: " + reason + ". "
	}
	if message != "" {
		detail = detail + "Message: " + message + ". "
	}
	return detail
}
