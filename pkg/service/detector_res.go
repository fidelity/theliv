/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package service

import (
	"context"
	"strings"
	"sync"

	in "github.com/fidelity/theliv/internal/investigators"
	"github.com/fidelity/theliv/internal/problem"
	com "github.com/fidelity/theliv/pkg/common"
	theErr "github.com/fidelity/theliv/pkg/err"
	"github.com/fidelity/theliv/pkg/kubeclient"
	log "github.com/fidelity/theliv/pkg/log"
	"github.com/fidelity/theliv/pkg/observability"

	networkv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type IngressWithIssue struct {
	obj    *networkv1.Ingress
	events []observability.EventRecord
}

func getUnhealthyIngress(ctx context.Context, input *problem.DetectorCreationInput) []*problem.Problem {
	list := []*IngressWithIssue{}
	problems := []*problem.Problem{}
	var wg sync.WaitGroup
	ingList, err := listNamespacedResource(input.KubeClient, ctx, &networkv1.IngressList{}, input.Namespace, "ingress")
	if err != nil {
		return problems
	}
	obj := ingList.(*networkv1.IngressList)
	for _, ingress := range obj.Items {
		list = append(list, &IngressWithIssue{
			obj: &ingress,
		})
	}
	wg.Add(len(obj.Items))
	for _, ingress := range list {
		go GetIngressEvents(ctx, input, &wg, ingress)
	}
	wg.Wait()
	for _, ingress := range list {
		if len(ingress.events) > 0 {
			latest := observability.EventRecord{}
			for index, event := range ingress.events {
				if index == 0 {
					latest = event
				}
				if event.LastTimestamp.After(latest.LastTimestamp) {
					latest = event
				}
			}
			if latest.Type != "Normal" {
				ingress.events = []observability.EventRecord{latest}
				problems = append(problems, buildIngressProblem(ingress))
			}
		}
	}
	return problems
}

func buildIngressProblem(ingress *IngressWithIssue) *problem.Problem {
	p := initProblem()
	p.Name = com.IngressMisconfigured
	p.Description = strings.Replace(strings.Replace(ingress.events[0].Message,
		"Failed build model due to ", "", 1), "Failed deploy model due to", "", 1)
	p.Tags = make(map[string]string)
	p.AffectedResources.ResourceName = ingress.obj.Name
	p.AffectedResources.ResourceKind = com.Ingress
	p.AffectedResources.Resource = ingress.obj
	p.Tags[com.Name] = ingress.obj.Name
	p.Tags[com.Namespace] = ingress.obj.Namespace
	p.Tags["uid"] = string(ingress.obj.UID)
	p.Tags[com.Resourcetype] = com.Ingress
	p.Tags["reason"] = com.IngressMisconfigured
	return &p
}

func GetIngressEvents(ctx context.Context, input *problem.DetectorCreationInput, wg *sync.WaitGroup,
	ingress *IngressWithIssue) IngressWithIssue {
	defer wg.Done()
	events, err := in.GetResourceEvents(ctx, input, ingress.obj.Name, ingress.obj.Namespace)
	if err != nil {
		log.SWithContext(ctx).Error("Got error when calling Kubernetes event API, error is %s", err)
	}
	ingress.events = events
	return *ingress
}

func listNamespacedResource(client *kubeclient.KubeClient, ctx context.Context,
	obj runtime.Object, ns string, resourceType string) (runtime.Object, error) {
	namespace := kubeclient.NamespacedName{
		Namespace: ns,
	}
	if client.List(ctx, obj, namespace, metav1.ListOptions{}) != nil {
		log.SWithContext(ctx).Errorf("No %s resources found in namespace: %s", resourceType, ns)
		return nil, theErr.NewCommonError(ctx, 4, com.LoadResourceFailed+resourceType)
	}
	return obj, nil
}
