/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/fidelity/theliv/internal/investigators"
	in "github.com/fidelity/theliv/internal/investigators"
	"github.com/fidelity/theliv/internal/problem"
	"github.com/fidelity/theliv/pkg/common"
	com "github.com/fidelity/theliv/pkg/common"
	"github.com/fidelity/theliv/pkg/config"
	theErr "github.com/fidelity/theliv/pkg/err"
	"github.com/fidelity/theliv/pkg/kubeclient"
	log "github.com/fidelity/theliv/pkg/log"
	"github.com/fidelity/theliv/pkg/observability/k8s"
	"github.com/fidelity/theliv/pkg/prometheus"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type investigatorFunc func(ctx context.Context, wg *sync.WaitGroup, problem *problem.Problem, input *problem.DetectorCreationInput)

// modify this map when adding new investigator func for alert
// for each alert, you can define one or more func to call to build details or solutions
var alertInvestigatorMap = map[string][]investigatorFunc{
	"PodNotRunning": {in.PodNotRunningInvestigator, in.PodNotRunningSolutionsInvestigator},
	"PodNotReady":   {in.PodNotReadyInvestigator},

	"ContainerWaitingAsImagePullBackOff":     {in.ContainerImagePullBackoffInvestigator},
	"ContainerWaitingAsCrashLoopBackoff":     {in.ContainerCrashLoopBackoffInvestigator},
	"InitContainerWaitingAsImagePullBackOff": {in.InitContainerImagePullBackoffInvestigator},

	"NodeNotReady":           {in.NodeNotReadyInvestigator},
	"NodeDiskPressure":       {in.NodeDiskPressureInvestigator},
	"NodeMemoryPressure":     {in.NodeMemoryPressureInvestigator},
	"NodePIDPressure":        {in.NodePIDPressureInvestigator},
	"NodeNetworkUnavailable": {in.NodeNetworkUnavailableInvestigator},

	"EndpointAddressNotAvailable": {in.EndpointAddressNotAvailableInvestigator},

	"DeploymentNotAvailable":       {in.DeploymentNotAvailableInvestigator},
	"DeploymentGenerationMismatch": {in.DeploymentGenerationMismatchInvestigator},
	"DeploymentReplicasMismatch":   {in.DeploymentReplicasMismatchInvestigator},

	com.IngressMisconfigured: {in.IngressMisconfiguredInvestigator},
}

func DetectAlerts(ctx context.Context) (interface{}, error) {
	var wg sync.WaitGroup
	contact := fmt.Sprintf(com.Contact, config.GetThelivConfig().TeamName)
	input := GetDetectorInput(ctx)

	client, err := kubeclient.NewKubeClient(ctx, input.Kubeconfig)
	if err != nil {
		return nil, theErr.NewCommonError(ctx, 4, com.LoadKubeConfigFailed+contact)
	}
	log.SWithContext(ctx).Infof("Kube client successfully created")
	input.KubeClient = client

	eventRetriever := k8s.NewK8sEventRetriever(client)
	input.EventRetriever = eventRetriever

	ingress := getUnhealthyIngress(ctx, input)
	alerts, err := prometheus.GetAlerts(ctx, input)
	if err != nil {
		return nil, theErr.NewCommonError(ctx, 6, com.PrometheusNotAvailable+contact)
	}
	log.SWithContext(ctx).Infof("%d prometheus alerts found", len(alerts.Alerts))

	// build problems from  alerts, problem is investigator input
	problems := buildProblemsFromAlerts(alerts.Alerts)
	if len(ingress) > 0 {
		problems = append(problems, ingress...)
	}
	problems = filterProblems(ctx, problems, input)
	log.SWithContext(ctx).Infof("Generated %d problems after filtering", len(problems))
	if err = buildProblemAffectedResource(ctx, &wg, problems, input); err != nil {
		return nil, theErr.NewCommonError(ctx, 4, com.LoadResourceFailed+contact)
	}

	problemresults := make([]*problem.Problem, 0)
	for _, p := range problems {
		if p.AffectedResources.Resource != nil {
			problemresults = append(problemresults, p)
			// check investigator func map or use common investigator for each problem
			if funcs, ok := alertInvestigatorMap[p.Name]; ok {
				for _, fc := range funcs {
					wg.Add(1)
					go fc(ctx, &wg, p, input)
				}
			} else {
				wg.Add(1)
				go investigators.CommonInvestigator(ctx, &wg, p, input)
			}
		}
	}

	wg.Wait()
	log.SWithContext(ctx).Infof("Generated %d problem results", len(problemresults))

	// Convert problems to report cards
	return problem.Aggregate(ctx, problemresults, client)
}

func buildProblemsFromAlerts(alerts []v1.Alert) []*problem.Problem {
	problems := make([]*problem.Problem, 0)
	for _, alert := range alerts {
		p := initProblem()
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

func initProblem() problem.Problem {
	return problem.Problem{
		Name:              "",
		Description:       "",
		Tags:              make(map[string]string),
		Level:             0,
		CauseLevel:        0,
		SolutionDetails:   common.InitLockedSlice(),
		UsefulCommands:    common.InitLockedSlice(),
		AffectedResources: problem.ResourceDetails{},
	}
}

// Classifies problems as cluster or namespace level, filters all other problems.
func filterProblems(ctx context.Context, problems []*problem.Problem, input *problem.DetectorCreationInput) []*problem.Problem {
	thelivcfg := config.GetThelivConfig()
	managednamespaces := thelivcfg.ProblemLevel.ManagedNamespaces
	results := make([]*problem.Problem, 0)
	for _, p := range problems {
		if p.Tags[com.Resourcetype] == com.Node || contains(managednamespaces, p.Tags[com.Namespace]) {
			// node & managednamespaces are cluster level problem
			p.Level = problem.Cluster
		} else if p.Tags[com.Namespace] == input.Namespace {
			p.Level = problem.UserNamespace
		} else {
			// filter out other problems that not related to user namespace
			continue
		}
		results = append(results, p)
	}
	return results
}

func buildProblemAffectedResource(ctx context.Context, wg *sync.WaitGroup, problems []*problem.Problem, input *problem.DetectorCreationInput) error {
	client := input.KubeClient
	wg.Add(len(problems))
	for _, problem := range problems {
		go loadResourceByType(ctx, wg, client, problem)
	}
	wg.Wait()
	return nil
}

func loadResourceByType(ctx context.Context, wg *sync.WaitGroup, client *kubeclient.KubeClient, problem *problem.Problem) error {
	defer wg.Done()
	switch problem.Tags[com.Resourcetype] {
	case com.Pod:
		loadNamespacedResource(client, ctx, problem, &corev1.Pod{}, com.Pod, "")
		problem.CauseLevel = 2
	case com.Container:
		loadNamespacedResource(client, ctx, problem, &corev1.Pod{}, com.Pod, com.Container)
		problem.CauseLevel = 1
	case com.Initcontainer:
		loadNamespacedResource(client, ctx, problem, &corev1.Pod{}, com.Pod, com.Container)
		problem.CauseLevel = 1
	case com.Deployment:
		loadNamespacedResource(client, ctx, problem, &appsv1.Deployment{}, com.Deployment, "")
		problem.CauseLevel = 4
	case com.Replicaset:
		loadNamespacedResource(client, ctx, problem, &appsv1.ReplicaSet{}, com.Replicaset, "")
		problem.CauseLevel = 3
	case com.Statefulset:
		loadNamespacedResource(client, ctx, problem, &appsv1.StatefulSet{}, com.Statefulset, "")
		problem.CauseLevel = 3
	case com.Daemonset:
		loadNamespacedResource(client, ctx, problem, &appsv1.DaemonSet{}, com.Daemonset, "")
		problem.CauseLevel = 3
	case com.Node:
		loadNamespacedResource(client, ctx, problem, &corev1.Node{}, com.Node, "")
		problem.CauseLevel = 0
	case com.Job:
		loadNamespacedResource(client, ctx, problem, &batchv1.Job{}, com.Job, "")
		problem.CauseLevel = 5
	case com.Cronjob:
		loadNamespacedResource(client, ctx, problem, &batchv1.CronJob{}, com.Cronjob, "")
		problem.CauseLevel = 5
	case com.Service:
		loadNamespacedResource(client, ctx, problem, &corev1.Service{}, com.Service, "")
		problem.CauseLevel = 6
	case com.Ingress:
		problem.CauseLevel = 7
	case com.Endpoint:
		loadNamespacedResource(client, ctx, problem, &corev1.Endpoints{}, com.Endpoint, "")
		problem.CauseLevel = 6
	default:
		log.SWithContext(ctx).Warnf("Not found affected resource for resource type %s: ", problem.Tags[com.Resourcetype])
	}
	return nil
}

func loadNamespacedResource(client *kubeclient.KubeClient, ctx context.Context,
	problem *problem.Problem, obj runtime.Object, resourceType string, subType string) {
	namespace := kubeclient.NamespacedName{
		Namespace: problem.Tags[com.Namespace],
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
		log.SWithContext(ctx).Errorf("Not found affected resource for %s: %s", problem.Tags[com.Resourcetype], buildName)
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
