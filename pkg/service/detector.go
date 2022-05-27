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
	"github.com/fidelity/theliv/internal/problemdetectors"
	"github.com/fidelity/theliv/pkg/config"
	"github.com/fidelity/theliv/pkg/kubeclient"
	"github.com/fidelity/theliv/pkg/prometheus"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"

	prob "github.com/fidelity/theliv/internal/problem"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Detect(ctx context.Context) (interface{}, error) {
	input := GetDetectorInput(ctx)

	pmgr := problem.DefaultProblemMgr()
	// Register detectors
	problemdetectors.Register(pmgr.DetectorRegistrationFunc())
	pbe, err := problem.NewDefaultProblemGraph(pmgr.Domains(), input)
	if err != nil {
		//TODO log
		return nil, err
	}
	r, err := pbe.Execute(ctx)
	if err != nil {
		return nil, err
	}

	problems := make([]problem.Problem, 0)
	for _, val := range r.ProblemMap {
		problems = append(problems, val...)
	}

	// Aggregator
	client, err := kubeclient.NewKubeClient(input.Kubeconfig)
	if err != nil {
		return nil, err
	}
	return problem.Aggregate(ctx, problems, client)
}

// ******************************
// New Investigator for Prometheus Alerts
// starts from here
// ******************************
type investigatorFunc func(ctx context.Context, problem *problem.NewProblem, input *problem.DetectorCreationInput)

// modify this map when adding new investigator func for alert
// for each alert, you can define one or more func to call to build details or solutions
var alertInvestigatorMap = map[string][]investigatorFunc{
	"AlertNameForCustomizedInvestigator": {PodNotRunningInvestigator, PodNotRunningSolutionsInvestigator},
}

func DetectAlerts(ctx context.Context) (interface{}, error) {
	thelivcfg := config.GetThelivConfig()
	managednamespaces := thelivcfg.ProblemLevel.ManagedNamespaces
	input := GetDetectorInput(ctx)
	alerts, _ := prometheus.GetAlerts(input)

	// build problems from  alerts, problem is investigator input
	problems := buildProblemsFromAlerts(alerts.Alerts) // no details
	buildProblemAffectedResource(ctx, problems, input) // client -> resource

	problemresults := make([]problem.NewProblem, 0)
	for _, problem := range problems {
		// check investigator func map or use common investigator for each problem
		if funcs, ok := alertInvestigatorMap[problem.Name]; ok {
			for _, fc := range funcs {
				fc(ctx, problem, input)
			}
		} else {
			investigators.CommonInvestigator(ctx, problem, input)
		}

		// build problem level
		if problem.Tags["resourcetype"] == "node" || contains(managednamespaces, problem.Tags["namespace"]) {
			// node & managednamespaces are cluster level problem
			problem.Level = prob.Cluster
		} else if problem.Tags["namespace"] == input.Namespace {
			problem.Level = prob.UserNamespace
		} else {
			// filter out other problems that not related to user namespace
			continue
		}
		problemresults = append(problemresults, *problem)
	}

	// Aggregator
	client, err := kubeclient.NewKubeClient(input.Kubeconfig)
	if err != nil {
		return nil, err
	}
	return problem.NewAggregate(ctx, problemresults, client)
}

func buildProblemsFromAlerts(alerts []v1.Alert) []*problem.NewProblem {
	problems := make([]*problem.NewProblem, 0)
	for _, alert := range alerts {
		p := problem.NewProblem{}
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

func buildProblemAffectedResource(ctx context.Context, problems []*problem.NewProblem, input *prob.DetectorCreationInput) []*problem.NewProblem {
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

func loadPodResource(client *kubeclient.KubeClient, ctx context.Context, problem *prob.NewProblem, input *prob.DetectorCreationInput) {
	namespace := kubeclient.NamespacedName{
		Namespace: input.Namespace,
		Name:      problem.Tags["pod"],
	}
	pod := &corev1.Pod{}
	getOptions := metav1.GetOptions{}
	client.Get(ctx, pod, namespace, getOptions)
	buildAffectedResource(problem, namespace.Name, "pod", pod)
}

func loadContainerResource(client *kubeclient.KubeClient, ctx context.Context, problem *prob.NewProblem, input *prob.DetectorCreationInput) {
	namespace := kubeclient.NamespacedName{
		Namespace: input.Namespace,
		Name:      problem.Tags["pod"],
	}
	containername := problem.Tags["container"]
	pod := &corev1.Pod{}
	getOptions := metav1.GetOptions{}
	client.Get(ctx, pod, namespace, getOptions)
	buildAffectedResource(problem, containername, "container", pod)
}

func loadDeploymentResource(client *kubeclient.KubeClient, ctx context.Context, problem *prob.NewProblem, input *prob.DetectorCreationInput) {
	namespace := kubeclient.NamespacedName{
		Namespace: input.Namespace,
		Name:      problem.Tags["deployment"],
	}
	deployment := &appsv1.Deployment{}
	getOptions := metav1.GetOptions{}
	client.Get(ctx, deployment, namespace, getOptions)
	buildAffectedResource(problem, namespace.Name, "deployment", deployment)
}

func loadReplicaSetResource(client *kubeclient.KubeClient, ctx context.Context, problem *prob.NewProblem, input *prob.DetectorCreationInput) {
	namespace := kubeclient.NamespacedName{
		Namespace: input.Namespace,
		Name:      problem.Tags["replicaset"],
	}
	rs := &appsv1.ReplicaSet{}
	getOptions := metav1.GetOptions{}
	client.Get(ctx, rs, namespace, getOptions)
	buildAffectedResource(problem, namespace.Name, "replicaset", rs)
}

func loadStatefulSetResource(client *kubeclient.KubeClient, ctx context.Context, problem *prob.NewProblem, input *prob.DetectorCreationInput) {
	namespace := kubeclient.NamespacedName{
		Namespace: input.Namespace,
		Name:      problem.Tags["statefulset"],
	}
	ss := &appsv1.StatefulSet{}
	getOptions := metav1.GetOptions{}
	client.Get(ctx, ss, namespace, getOptions)
	buildAffectedResource(problem, namespace.Name, "statefulset", ss)
}

func loadDaemonSetResource(client *kubeclient.KubeClient, ctx context.Context, problem *prob.NewProblem, input *prob.DetectorCreationInput) {
	namespace := kubeclient.NamespacedName{
		Namespace: input.Namespace,
		Name:      problem.Tags["daemonset"],
	}
	ds := &appsv1.DaemonSet{}
	getOptions := metav1.GetOptions{}
	client.Get(ctx, ds, namespace, getOptions)
	buildAffectedResource(problem, namespace.Name, "daemonset", ds)
}

func loadNodeResource(client *kubeclient.KubeClient, ctx context.Context, problem *prob.NewProblem, input *prob.DetectorCreationInput) {
	namespace := kubeclient.NamespacedName{
		Namespace: input.Namespace,
		Name:      problem.Tags["node"],
	}
	node := &corev1.Node{}
	getOptions := metav1.GetOptions{}
	client.Get(ctx, node, namespace, getOptions)
	buildAffectedResource(problem, namespace.Name, "node", node)
}

func loadJobResource(client *kubeclient.KubeClient, ctx context.Context, problem *prob.NewProblem, input *prob.DetectorCreationInput) {
	namespace := kubeclient.NamespacedName{
		Namespace: input.Namespace,
		Name:      problem.Tags["job"],
	}
	job := &batchv1.Job{}
	getOptions := metav1.GetOptions{}
	client.Get(ctx, job, namespace, getOptions)
	buildAffectedResource(problem, namespace.Name, "job", job)
}

func loadCronJobResource(client *kubeclient.KubeClient, ctx context.Context, problem *prob.NewProblem, input *prob.DetectorCreationInput) {
	namespace := kubeclient.NamespacedName{
		Namespace: input.Namespace,
		Name:      problem.Tags["cronjob"],
	}
	cronjob := &batchv1.CronJob{}
	getOptions := metav1.GetOptions{}
	client.Get(ctx, cronjob, namespace, getOptions)
	buildAffectedResource(problem, namespace.Name, "cronjob", cronjob)
}

func loadServiceResource(client *kubeclient.KubeClient, ctx context.Context, problem *prob.NewProblem, input *prob.DetectorCreationInput) {
	namespace := kubeclient.NamespacedName{
		Namespace: input.Namespace,
		Name:      problem.Tags["service"],
	}
	service := &corev1.Service{}
	getOptions := metav1.GetOptions{}
	client.Get(ctx, service, namespace, getOptions)
	buildAffectedResource(problem, namespace.Name, "service", service)
}

func loadIngressResource(client *kubeclient.KubeClient, ctx context.Context, problem *prob.NewProblem, input *prob.DetectorCreationInput) {
	namespace := kubeclient.NamespacedName{
		Namespace: input.Namespace,
		Name:      problem.Tags["ingress"],
	}
	ingress := &networkv1.Ingress{}
	getOptions := metav1.GetOptions{}
	client.Get(ctx, ingress, namespace, getOptions)
	buildAffectedResource(problem, namespace.Name, "ingress", ingress)
}

func loadEndpointsResource(client *kubeclient.KubeClient, ctx context.Context, problem *prob.NewProblem, input *prob.DetectorCreationInput) {
	namespace := kubeclient.NamespacedName{
		Namespace: input.Namespace,
		Name:      problem.Tags["endpoint"],
	}
	endpoints := &corev1.Endpoints{}
	getOptions := metav1.GetOptions{}
	client.Get(ctx, endpoints, namespace, getOptions)
	buildAffectedResource(problem, namespace.Name, "endpoint", endpoints)
}

func buildAffectedResource(problem *problem.NewProblem, resourceName string, resourceKind string, object runtime.Object) {
	problem.AffectedResources.ResourceName = resourceName
	problem.AffectedResources.ResourceKind = resourceKind
	problem.AffectedResources.Resource = object
	// if owner_kind, ok := problem.Tags["owner_kind"]; ok {
	// 	problem.AffectedResources.OwnerKind = problem.Tags[owner_kind]
	// }
	// if owner_name, ok := problem.Tags["owner_name"]; ok {
	// 	problem.AffectedResources.OwnerKind = problem.Tags[owner_name]
	// }
}

// create a seperate go file in ./internal/investigators for each investigator
func PodNotRunningInvestigator(ctx context.Context, problem *problem.NewProblem, input *problem.DetectorCreationInput) {
	// Load kubenetes resource details
	detail := "do something to kubeneter get details"
	problem.SolutionDetails = append(problem.SolutionDetails, &detail)
}

func PodNotRunningSolutionsInvestigator(ctx context.Context, problem *problem.NewProblem, input *problem.DetectorCreationInput) {
	// Generate solutions
	detail := "do something to provide solutions"
	problem.SolutionDetails = append(problem.SolutionDetails, &detail)
}

func contains(list []string, str string) bool {
	for _, l := range list {
		if str == l {
			return true
		}
	}
	return false
}
