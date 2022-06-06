/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package service

import (
	"context"

	golog "log"

	"github.com/fidelity/theliv/internal/investigators"
	"github.com/fidelity/theliv/internal/problem"
	"github.com/fidelity/theliv/pkg/config"
	"github.com/fidelity/theliv/pkg/kubeclient"
	"github.com/fidelity/theliv/pkg/prometheus"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type investigatorFunc func(ctx context.Context, problem *problem.Problem, input *problem.DetectorCreationInput)

// modify this map when adding new investigator func for alert
// for each alert, you can define one or more func to call to build details or solutions
var alertInvestigatorMap = map[string][]investigatorFunc{
	"PodNotRunning":                          {investigators.PodNotRunningInvestigator, investigators.PodNotRunningSolutionsInvestigator},
	"ContainerWaitingAsImagePullBackOff":     {investigators.ContainerImagePullBackoffInvestigator},
	"InitContainerWaitingAsImagePullBackOff": {investigators.InitContainerImagePullBackoffInvestigator},
}

func DetectAlerts(ctx context.Context) (interface{}, error) {
	input := GetDetectorInput(ctx)
	alerts, _ := prometheus.GetAlerts(input)

	// build problems from  alerts, problem is investigator input
	problems := buildProblemsFromAlerts(alerts.Alerts)
	problems = filterProblems(ctx, problems, input)
	buildProblemAffectedResource(ctx, problems, input)

	problemresults := make([]problem.Problem, 0)
	for _, p := range problems {
		// check investigator func map or use common investigator for each problem
		if funcs, ok := alertInvestigatorMap[p.Name]; ok {
			for _, fc := range funcs {
				fc(ctx, p, input)
			}
		} else {
			investigators.CommonInvestigator(ctx, p, input)
		}
		problemresults = append(problemresults, *p)
	}

	// Aggregator
	client, err := kubeclient.NewKubeClient(input.Kubeconfig)
	if err != nil {
		return nil, err
	}
	return problem.Aggregate(ctx, problemresults, client)
}

func buildProblemsFromAlerts(alerts []v1.Alert) []*problem.Problem {
	problems := make([]*problem.Problem, 0)
	for _, alert := range alerts {
		p := problem.Problem{}
		p.Name = string(alert.Labels[model.LabelName("alertname")])
		p.Description = string(alert.Annotations[model.LabelName("description")])
		p.Tags = make(map[string]string)
		for ln, lv := range alert.Labels {
			p.Tags[string(ln)] = string(lv)
		}
		problems = append(problems, &p)
	}
	return problems
}

func filterProblems(ctx context.Context, problems []*problem.Problem, input *problem.DetectorCreationInput) []*problem.Problem {
	thelivcfg := config.GetThelivConfig()
	managednamespaces := thelivcfg.ProblemLevel.ManagedNamespaces
	results := make([]*problem.Problem, 0)
	for _, p := range problems {
		if p.Tags["resourcetype"] == "node" || contains(managednamespaces, p.Tags["namespace"]) {
			// node & managednamespaces are cluster level problem
			p.Level = problem.Cluster
		} else if p.Tags["namespace"] == input.Namespace {
			p.Level = problem.UserNamespace
		} else {
			// filter out other problems that not related to user namespace
			continue
		}
		results = append(results, p)
	}
	return results
}

func buildProblemAffectedResource(ctx context.Context, problems []*problem.Problem, input *problem.DetectorCreationInput) []*problem.Problem {
	client, err := kubeclient.NewKubeClient(input.Kubeconfig)
	if err != nil {
		golog.Printf("ERROR - Got error when initiating kubeclient to load affected resource, error is %s", err)
	}
	for _, problem := range problems {
		switch problem.Tags["resourcetype"] {
		case "pod":
			loadPodResource(client, ctx, problem, input)
			problem.CauseLevel = 1
		case "container":
			loadContainerResource(client, ctx, problem, input)
			problem.CauseLevel = 1
		case "initcontainer":
			loadContainerResource(client, ctx, problem, input)
			problem.CauseLevel = 1
		case "deployment":
			loadDeploymentResource(client, ctx, problem, input)
			problem.CauseLevel = 3
		case "replicaset":
			loadReplicaSetResource(client, ctx, problem, input)
			problem.CauseLevel = 2
		case "statefulset":
			loadStatefulSetResource(client, ctx, problem, input)
			problem.CauseLevel = 2
		case "daemonset":
			loadDaemonSetResource(client, ctx, problem, input)
			problem.CauseLevel = 2
		case "node":
			loadNodeResource(client, ctx, problem, input)
			problem.CauseLevel = 0
		case "job":
			loadJobResource(client, ctx, problem, input)
			problem.CauseLevel = 4
		case "cronjob":
			loadCronJobResource(client, ctx, problem, input)
			problem.CauseLevel = 4
		case "service":
			loadServiceResource(client, ctx, problem, input)
			problem.CauseLevel = 5
		case "ingress":
			loadIngressResource(client, ctx, problem, input)
			problem.CauseLevel = 6
		case "endpoint":
			loadEndpointsResource(client, ctx, problem, input)
			problem.CauseLevel = 5
		default:
			golog.Printf("WARN - Not found affected resource for resource type %s: ", problem.Tags["resourcetype"])
		}
	}
	return problems
}

func loadPodResource(client *kubeclient.KubeClient, ctx context.Context, problem *problem.Problem, input *problem.DetectorCreationInput) {
	namespace := kubeclient.NamespacedName{
		Namespace: problem.Tags["namespace"],
		Name:      problem.Tags["pod"],
	}
	pod := &corev1.Pod{}
	getOptions := metav1.GetOptions{}
	if client.Get(ctx, pod, namespace, getOptions) == nil {
		buildAffectedResource(problem, namespace.Name, "pod", pod)
	} else {
		golog.Printf("ERROR - Not found affected resource for resource type %s: ", problem.Tags["resourcetype"])
	}
}

func loadContainerResource(client *kubeclient.KubeClient, ctx context.Context, problem *problem.Problem, input *problem.DetectorCreationInput) {
	namespace := kubeclient.NamespacedName{
		Namespace: problem.Tags["namespace"],
		Name:      problem.Tags["pod"],
	}
	containername := problem.Tags["container"]
	pod := &corev1.Pod{}
	getOptions := metav1.GetOptions{}
	if client.Get(ctx, pod, namespace, getOptions) == nil {
		buildAffectedResource(problem, containername, "container", pod)
	} else {
		golog.Printf("ERROR - Not found affected resource for resource type %s: ", problem.Tags["resourcetype"])
	}
}

func loadDeploymentResource(client *kubeclient.KubeClient, ctx context.Context, problem *problem.Problem, input *problem.DetectorCreationInput) {
	namespace := kubeclient.NamespacedName{
		Namespace: problem.Tags["namespace"],
		Name:      problem.Tags["deployment"],
	}
	deployment := &appsv1.Deployment{}
	getOptions := metav1.GetOptions{}
	if client.Get(ctx, deployment, namespace, getOptions) == nil {
		buildAffectedResource(problem, namespace.Name, "deployment", deployment)
	} else {
		golog.Printf("ERROR - Not found affected resource for resource type %s: ", problem.Tags["resourcetype"])
	}
}

func loadReplicaSetResource(client *kubeclient.KubeClient, ctx context.Context, problem *problem.Problem, input *problem.DetectorCreationInput) {
	namespace := kubeclient.NamespacedName{
		Namespace: problem.Tags["namespace"],
		Name:      problem.Tags["replicaset"],
	}
	rs := &appsv1.ReplicaSet{}
	getOptions := metav1.GetOptions{}
	if client.Get(ctx, rs, namespace, getOptions) == nil {
		buildAffectedResource(problem, namespace.Name, "replicaset", rs)
	} else {
		golog.Printf("ERROR - Not found affected resource for resource type %s: ", problem.Tags["resourcetype"])
	}
}

func loadStatefulSetResource(client *kubeclient.KubeClient, ctx context.Context, problem *problem.Problem, input *problem.DetectorCreationInput) {
	namespace := kubeclient.NamespacedName{
		Namespace: problem.Tags["namespace"],
		Name:      problem.Tags["statefulset"],
	}
	ss := &appsv1.StatefulSet{}
	getOptions := metav1.GetOptions{}
	if client.Get(ctx, ss, namespace, getOptions) == nil {
		buildAffectedResource(problem, namespace.Name, "statefulset", ss)
	} else {
		golog.Printf("ERROR - Not found affected resource for resource type %s: ", problem.Tags["resourcetype"])
	}
}

func loadDaemonSetResource(client *kubeclient.KubeClient, ctx context.Context, problem *problem.Problem, input *problem.DetectorCreationInput) {
	namespace := kubeclient.NamespacedName{
		Namespace: problem.Tags["namespace"],
		Name:      problem.Tags["daemonset"],
	}
	ds := &appsv1.DaemonSet{}
	getOptions := metav1.GetOptions{}
	if client.Get(ctx, ds, namespace, getOptions) == nil {
		buildAffectedResource(problem, namespace.Name, "daemonset", ds)
	} else {
		golog.Printf("ERROR - Not found affected resource for resource type %s: ", problem.Tags["resourcetype"])
	}
}

func loadNodeResource(client *kubeclient.KubeClient, ctx context.Context, problem *problem.Problem, input *problem.DetectorCreationInput) {
	namespace := kubeclient.NamespacedName{
		Namespace: problem.Tags["namespace"],
		Name:      problem.Tags["node"],
	}
	node := &corev1.Node{}
	getOptions := metav1.GetOptions{}
	if client.Get(ctx, node, namespace, getOptions) == nil {
		buildAffectedResource(problem, namespace.Name, "node", node)
	} else {
		golog.Printf("ERROR - Not found affected resource for resource type %s: ", problem.Tags["resourcetype"])
	}
}

func loadJobResource(client *kubeclient.KubeClient, ctx context.Context, problem *problem.Problem, input *problem.DetectorCreationInput) {
	namespace := kubeclient.NamespacedName{
		Namespace: problem.Tags["namespace"],
		Name:      problem.Tags["job"],
	}
	job := &batchv1.Job{}
	getOptions := metav1.GetOptions{}
	if client.Get(ctx, job, namespace, getOptions) == nil {
		buildAffectedResource(problem, namespace.Name, "job", job)
	} else {
		golog.Printf("ERROR - Not found affected resource for resource type %s: ", problem.Tags["resourcetype"])
	}
}

func loadCronJobResource(client *kubeclient.KubeClient, ctx context.Context, problem *problem.Problem, input *problem.DetectorCreationInput) {
	namespace := kubeclient.NamespacedName{
		Namespace: problem.Tags["namespace"],
		Name:      problem.Tags["cronjob"],
	}
	cronjob := &batchv1.CronJob{}
	getOptions := metav1.GetOptions{}
	if client.Get(ctx, cronjob, namespace, getOptions) == nil {
		buildAffectedResource(problem, namespace.Name, "cronjob", cronjob)
	} else {
		golog.Printf("ERROR - Not found affected resource for resource type %s: ", problem.Tags["resourcetype"])
	}
}

func loadServiceResource(client *kubeclient.KubeClient, ctx context.Context, problem *problem.Problem, input *problem.DetectorCreationInput) {
	namespace := kubeclient.NamespacedName{
		Namespace: problem.Tags["namespace"],
		Name:      problem.Tags["service"],
	}
	service := &corev1.Service{}
	getOptions := metav1.GetOptions{}
	if client.Get(ctx, service, namespace, getOptions) == nil {
		buildAffectedResource(problem, namespace.Name, "service", service)
	} else {
		golog.Printf("ERROR - Not found affected resource for resource type %s: ", problem.Tags["resourcetype"])
	}
}

func loadIngressResource(client *kubeclient.KubeClient, ctx context.Context, problem *problem.Problem, input *problem.DetectorCreationInput) {
	namespace := kubeclient.NamespacedName{
		Namespace: problem.Tags["namespace"],
		Name:      problem.Tags["ingress"],
	}
	ingress := &networkv1.Ingress{}
	getOptions := metav1.GetOptions{}
	if client.Get(ctx, ingress, namespace, getOptions) == nil {
		buildAffectedResource(problem, namespace.Name, "ingress", ingress)
	} else {
		golog.Printf("ERROR - Not found affected resource for resource type %s: ", problem.Tags["resourcetype"])
	}
}

func loadEndpointsResource(client *kubeclient.KubeClient, ctx context.Context, problem *problem.Problem, input *problem.DetectorCreationInput) {
	namespace := kubeclient.NamespacedName{
		Namespace: problem.Tags["namespace"],
		Name:      problem.Tags["endpoint"],
	}
	endpoints := &corev1.Endpoints{}
	getOptions := metav1.GetOptions{}
	if client.Get(ctx, endpoints, namespace, getOptions) == nil {
		buildAffectedResource(problem, namespace.Name, "endpoint", endpoints)
	} else {
		golog.Printf("ERROR - Not found affected resource for resource type %s: ", problem.Tags["resourcetype"])
	}
}

func buildAffectedResource(problem *problem.Problem, resourceName string, resourceKind string, object runtime.Object) {
	problem.AffectedResources.ResourceName = resourceName
	problem.AffectedResources.ResourceKind = resourceKind
	problem.AffectedResources.Resource = object
}

func contains(list []string, str string) bool {
	for _, l := range list {
		if str == l {
			return true
		}
	}
	return false
}
