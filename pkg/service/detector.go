/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package service

import (
	"context"

	log "github.com/fidelity/theliv/pkg/log"

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
	"PodNotRunning":                      {investigators.PodNotRunningInvestigator, investigators.PodNotRunningSolutionsInvestigator},
	"ContainerWaitingAsImagePullBackOff": {investigators.ContainerImagePullBackoffInvestigator},
	// "InitContainerWaitingAsImagePullBackOff": {investigators.InitContainerImagePullBackoffInvestigator},
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
		log.S().Errorf("Got error when initiating kubeclient to load affected resource, error is %s", err)
	}
	for _, problem := range problems {
		switch problem.Tags["resourcetype"] {
		case "pod":
			loadNamespacedResource(client, ctx, problem, &corev1.Pod{}, "pod", "")
			problem.CauseLevel = 1
		case "container":
			loadNamespacedResource(client, ctx, problem, &corev1.Pod{}, "pod", "container")
			problem.CauseLevel = 1
		case "deployment":
			loadNamespacedResource(client, ctx, problem, &appsv1.Deployment{}, "deployment", "")
			problem.CauseLevel = 3
		case "replicaset":
			loadNamespacedResource(client, ctx, problem, &appsv1.ReplicaSet{}, "replicaset", "")
			problem.CauseLevel = 2
		case "statefulset":
			loadNamespacedResource(client, ctx, problem, &appsv1.StatefulSet{}, "statefulset", "")
			problem.CauseLevel = 2
		case "daemonset":
			loadNamespacedResource(client, ctx, problem, &appsv1.DaemonSet{}, "daemonset", "")
			problem.CauseLevel = 2
		case "node":
			loadNamespacedResource(client, ctx, problem, &corev1.Node{}, "node", "")
			problem.CauseLevel = 0
		case "job":
			loadNamespacedResource(client, ctx, problem, &batchv1.Job{}, "job", "")
			problem.CauseLevel = 4
		case "cronjob":
			loadNamespacedResource(client, ctx, problem, &batchv1.CronJob{}, "cronjob", "")
			problem.CauseLevel = 4
		case "service":
			loadNamespacedResource(client, ctx, problem, &corev1.Service{}, "service", "")
			problem.CauseLevel = 5
		case "ingress":
			loadNamespacedResource(client, ctx, problem, &networkv1.Ingress{}, "ingress", "")
			problem.CauseLevel = 6
		case "endpoint":
			loadNamespacedResource(client, ctx, problem, &corev1.Endpoints{}, "endpoint", "")
			problem.CauseLevel = 5
		default:
			log.S().Warnf("WARN - Not found affected resource for resource type %s: ", problem.Tags["resourcetype"])
		}
	}
	return problems
}

func loadNamespacedResource(client *kubeclient.KubeClient, ctx context.Context,
	problem *problem.Problem, obj runtime.Object, resourceType string, subType string) {
	namespace := kubeclient.NamespacedName{
		Namespace: problem.Tags["namespace"],
		Name:      problem.Tags[resourceType],
	}
	buildName := namespace.Name
	buildType := resourceType
	if subType != "" {
		buildName = problem.Tags[subType]
		buildType = subType
	}
	if client.Get(ctx, obj, namespace, metav1.GetOptions{}) == nil {
		buildAffectedResource(problem, buildName, buildType, obj)
	} else {
		log.S().Errorf("Not found affected resource for resource type %s: ", problem.Tags["resourcetype"])
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
