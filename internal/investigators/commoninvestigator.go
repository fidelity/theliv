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
	"github.com/fidelity/theliv/pkg/kubeclient"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	networkv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CommonInvestigator(ctx context.Context, problem *problem.NewProblem, input *problem.DetectorCreationInput) {
	switch problem.Tags["resourcetype"] {
	case "pod":
		loadPodDetails(ctx, problem, input)
	case "container":
		loadContainerDetails(ctx, problem, input)
	case "deployment":
		loadDeploymentDetails(ctx, problem, input)
	case "replicaset":
		loadReplicaSetDetails(ctx, problem, input)
	case "statefulset":
		loadStatefulSetDetails(ctx, problem, input)
	case "daemonset":
		loadDaemonSetDetails(ctx, problem, input)
	case "node":
		loadNodeDetails(ctx, problem, input)
	case "job":
		loadJobDetails(ctx, problem, input)
	case "cronjob":
		loadCronJobDetails(ctx, problem, input)
	case "service":
		loadServiceDetails(ctx, problem, input)
	case "ingress":
		loadIngressDetails(ctx, problem, input)
	case "endpoint":
		loadEndpointsDetails(ctx, problem, input)
	default:
		golog.Printf("WARN - Not found investigator function for resource type %s", problem.Tags["resourcetype"])
	}
}

func loadPodDetails(ctx context.Context, problem *problem.NewProblem, input *problem.DetectorCreationInput) {
	client, err := kubeclient.NewKubeClient(input.Kubeconfig)
	if err != nil {
		golog.Printf("ERROR - Got error when initiating kubeclient to load pod details, error is %s", err)
	}
	namespace := kubeclient.NamespacedName{
		Namespace: input.Namespace,
		Name:      problem.Tags["pod"],
	}
	pod := &v1.Pod{}
	getOptions := metav1.GetOptions{}
	client.Get(ctx, pod, namespace, getOptions)

	golog.Printf("INFO - Checking status with pod %s", pod.Name)
	if pod.Status.Message == "" && pod.Status.Reason == "" {
		return
	}
	detail := buildReasonMsg(pod.Status.Reason, pod.Status.Message)
	problem.Details = append(problem.Details, &detail)
}

func loadContainerDetails(ctx context.Context, problem *problem.NewProblem, input *problem.DetectorCreationInput) {
	client, err := kubeclient.NewKubeClient(input.Kubeconfig)
	if err != nil {
		golog.Printf("ERROR - Got error when initiating kubeclient to load container details, error is %s", err)
	}
	namespace := kubeclient.NamespacedName{
		Namespace: input.Namespace,
		Name:      problem.Tags["pod"],
	}
	containername := problem.Tags["container"]
	pod := &v1.Pod{}
	getOptions := metav1.GetOptions{}
	client.Get(ctx, pod, namespace, getOptions)
	golog.Printf("INFO - Checking init container with pod %s", pod.Name)
	for _, status := range pod.Status.InitContainerStatuses {
		if status.Name == containername {
			if status.State.Terminated != nil {
				detail := buildReasonMsg(status.State.Terminated.Reason, status.State.Terminated.Message)
				problem.Details = append(problem.Details, &detail)
			}
			if status.State.Waiting != nil {
				detail := buildReasonMsg(status.State.Waiting.Reason, status.State.Waiting.Message)
				problem.Details = append(problem.Details, &detail)
			}
			break
		}
	}

	golog.Printf("INFO - Checking container with pod %s", pod.Name)
	for _, status := range pod.Status.ContainerStatuses {
		if status.Name == containername {
			if status.State.Terminated != nil {
				detail := buildReasonMsg(status.State.Terminated.Reason, status.State.Terminated.Message)
				problem.Details = append(problem.Details, &detail)
			}
			if status.State.Waiting != nil {
				detail := buildReasonMsg(status.State.Waiting.Reason, status.State.Waiting.Message)
				problem.Details = append(problem.Details, &detail)
			}
			break
		}
	}
}

func loadDeploymentDetails(ctx context.Context, problem *problem.NewProblem, input *problem.DetectorCreationInput) {
	client, err := kubeclient.NewKubeClient(input.Kubeconfig)
	if err != nil {
		golog.Printf("ERROR - Got error when initiating kubeclient to load deployment details, error is %s", err)
	}
	namespace := kubeclient.NamespacedName{
		Namespace: input.Namespace,
		Name:      problem.Tags["deployment"],
	}
	deployment := &appsv1.Deployment{}
	getOptions := metav1.GetOptions{}
	client.Get(ctx, deployment, namespace, getOptions)

	golog.Printf("INFO - Checking status with deployment %s", deployment.Name)
	for _, condition := range deployment.Status.Conditions {
		detail := string(condition.Type) + "=" + string(condition.Status)
		if condition.Message != "" || condition.Reason != "" {
			detail = detail + buildReasonMsg(condition.Reason, condition.Message)
			problem.Details = append(problem.Details, &detail)
		}
	}
}

func loadReplicaSetDetails(ctx context.Context, problem *problem.NewProblem, input *problem.DetectorCreationInput) {
	client, err := kubeclient.NewKubeClient(input.Kubeconfig)
	if err != nil {
		golog.Printf("ERROR - Got error when initiating kubeclient to load replicaset details, error is %s", err)
	}
	namespace := kubeclient.NamespacedName{
		Namespace: input.Namespace,
		Name:      problem.Tags["replicaset"],
	}
	rs := &appsv1.ReplicaSet{}
	getOptions := metav1.GetOptions{}
	client.Get(ctx, rs, namespace, getOptions)

	golog.Printf("INFO - Checking status with replicaset %s", rs.Name)
	for _, condition := range rs.Status.Conditions {
		detail := string(condition.Type) + "=" + string(condition.Status)
		if condition.Message != "" || condition.Reason != "" {
			detail = detail + buildReasonMsg(condition.Reason, condition.Message)
			problem.Details = append(problem.Details, &detail)
		}
	}
}

func loadStatefulSetDetails(ctx context.Context, problem *problem.NewProblem, input *problem.DetectorCreationInput) {
	client, err := kubeclient.NewKubeClient(input.Kubeconfig)
	if err != nil {
		golog.Printf("ERROR - Got error when initiating kubeclient to load statefulset details, error is %s", err)
	}
	namespace := kubeclient.NamespacedName{
		Namespace: input.Namespace,
		Name:      problem.Tags["statefulset"],
	}
	ss := &appsv1.StatefulSet{}
	getOptions := metav1.GetOptions{}
	client.Get(ctx, ss, namespace, getOptions)

	golog.Printf("INFO - Checking status with statefulset %s", ss.Name)
	for _, condition := range ss.Status.Conditions {
		detail := string(condition.Type) + "=" + string(condition.Status)
		if condition.Message != "" || condition.Reason != "" {
			detail = detail + buildReasonMsg(condition.Reason, condition.Message)
			problem.Details = append(problem.Details, &detail)
		}
	}
}

func loadDaemonSetDetails(ctx context.Context, problem *problem.NewProblem, input *problem.DetectorCreationInput) {
	client, err := kubeclient.NewKubeClient(input.Kubeconfig)
	if err != nil {
		golog.Printf("ERROR - Got error when initiating kubeclient to load daemonset details, error is %s", err)
	}
	namespace := kubeclient.NamespacedName{
		Namespace: input.Namespace,
		Name:      problem.Tags["daemonset"],
	}
	ds := &appsv1.DaemonSet{}
	getOptions := metav1.GetOptions{}
	client.Get(ctx, ds, namespace, getOptions)

	golog.Printf("INFO - Checking status with daemonset %s", ds.Name)
	for _, condition := range ds.Status.Conditions {
		detail := string(condition.Type) + "=" + string(condition.Status)
		if condition.Message != "" || condition.Reason != "" {
			detail = detail + buildReasonMsg(condition.Reason, condition.Message)
			problem.Details = append(problem.Details, &detail)
		}
	}
}

func loadNodeDetails(ctx context.Context, problem *problem.NewProblem, input *problem.DetectorCreationInput) {
	client, err := kubeclient.NewKubeClient(input.Kubeconfig)
	if err != nil {
		golog.Printf("ERROR - Got error when initiating kubeclient to load node details, error is %s", err)
	}
	namespace := kubeclient.NamespacedName{
		Namespace: input.Namespace,
		Name:      problem.Tags["node"],
	}
	node := &v1.Node{}
	getOptions := metav1.GetOptions{}
	client.Get(ctx, node, namespace, getOptions)

	golog.Printf("INFO - Checking status with node %s", node.Name)
	for _, condition := range node.Status.Conditions {
		detail := string(condition.Type) + "=" + string(condition.Status)
		if condition.Message != "" || condition.Reason != "" {
			detail = detail + buildReasonMsg(condition.Reason, condition.Message)
			problem.Details = append(problem.Details, &detail)
		}
	}
}

func loadJobDetails(ctx context.Context, problem *problem.NewProblem, input *problem.DetectorCreationInput) {
	client, err := kubeclient.NewKubeClient(input.Kubeconfig)
	if err != nil {
		golog.Printf("ERROR - Got error when initiating kubeclient to load job details, error is %s", err)
	}
	namespace := kubeclient.NamespacedName{
		Namespace: input.Namespace,
		Name:      problem.Tags["job"],
	}
	job := &batchv1.Job{}
	getOptions := metav1.GetOptions{}
	client.Get(ctx, job, namespace, getOptions)

	golog.Printf("INFO - Checking status with job %s", job.Name)
	for _, condition := range job.Status.Conditions {
		detail := string(condition.Type) + "=" + string(condition.Status)
		if condition.Message != "" || condition.Reason != "" {
			detail = detail + buildReasonMsg(condition.Reason, condition.Message)
			problem.Details = append(problem.Details, &detail)
		}
	}
}

func loadCronJobDetails(ctx context.Context, problem *problem.NewProblem, input *problem.DetectorCreationInput) {
	client, err := kubeclient.NewKubeClient(input.Kubeconfig)
	if err != nil {
		golog.Printf("ERROR - Got error when initiating kubeclient to load cron job details, error is %s", err)
	}
	namespace := kubeclient.NamespacedName{
		Namespace: input.Namespace,
		Name:      problem.Tags["cronjob"],
	}
	job := &batchv1.CronJob{}
	getOptions := metav1.GetOptions{}
	client.Get(ctx, job, namespace, getOptions)

	golog.Printf("INFO - Checking status with cron job %s", job.Name)
	for _, job := range job.Status.Active {
		if job.Name != "" && job.Namespace != "" {
			detail := job.Name + "in " + job.Namespace + "is active."
			problem.Details = append(problem.Details, &detail)
		}
	}
}

func loadServiceDetails(ctx context.Context, problem *problem.NewProblem, input *problem.DetectorCreationInput) {
	client, err := kubeclient.NewKubeClient(input.Kubeconfig)
	if err != nil {
		golog.Printf("ERROR - Got error when initiating kubeclient to load service details, error is %s", err)
	}
	namespace := kubeclient.NamespacedName{
		Namespace: input.Namespace,
		Name:      problem.Tags["service"],
	}
	service := &v1.Service{}
	getOptions := metav1.GetOptions{}
	client.Get(ctx, service, namespace, getOptions)

	golog.Printf("INFO - Checking status with service %s", service.Name)
	// TODO: if need to check service.Status.LoadBalancer.Ingress[].Ports[] ?????
	for _, condition := range service.Status.Conditions {
		detail := string(condition.Type) + "=" + string(condition.Status)
		if condition.Message != "" || condition.Reason != "" {
			detail = detail + buildReasonMsg(condition.Reason, condition.Message)
			problem.Details = append(problem.Details, &detail)
		}
	}
}

func loadIngressDetails(ctx context.Context, problem *problem.NewProblem, input *problem.DetectorCreationInput) {
	client, err := kubeclient.NewKubeClient(input.Kubeconfig)
	if err != nil {
		golog.Printf("ERROR - Got error when initiating kubeclient to load ingress details, error is %s", err)
	}
	namespace := kubeclient.NamespacedName{
		Namespace: input.Namespace,
		Name:      problem.Tags["ingress"],
	}
	ingress := &networkv1.Ingress{}
	getOptions := metav1.GetOptions{}
	client.Get(ctx, ingress, namespace, getOptions)

	golog.Printf("INFO - Checking status with ingress %s", ingress.Name)
	for _, ingress := range ingress.Status.LoadBalancer.Ingress {
		for _, port := range ingress.Ports {
			if port.Error != nil {
				detail := *port.Error
				problem.Details = append(problem.Details, &detail)
			}
		}
	}
}

func loadEndpointsDetails(ctx context.Context, problem *problem.NewProblem, input *problem.DetectorCreationInput) {
	client, err := kubeclient.NewKubeClient(input.Kubeconfig)
	if err != nil {
		golog.Printf("ERROR - Got error when initiating kubeclient to load endpoints details, error is %s", err)
	}
	namespace := kubeclient.NamespacedName{
		Namespace: input.Namespace,
		Name:      problem.Tags["endpoint"],
	}
	endpoints := &v1.Endpoints{}
	getOptions := metav1.GetOptions{}
	client.Get(ctx, endpoints, namespace, getOptions)

	golog.Printf("INFO - Checking status with endpoints %s", endpoints.Name)
	for _, sub := range endpoints.Subsets {
		if sub.NotReadyAddresses != nil {
			for _, addr := range sub.NotReadyAddresses {
				detail := addr.IP + "not ready. Target object is " + addr.TargetRef.Name
				problem.Details = append(problem.Details, &detail)
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
