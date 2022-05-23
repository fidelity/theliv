/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package problem

import (
	"context"
	"fmt"
	"sort"

	"github.com/fidelity/theliv/pkg/kubeclient"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func NewAggregate(ctx context.Context, problems []NewProblem, client *kubeclient.KubeClient) (interface{}, error) {
	// TODO get client from context

	cards := make([]*ReportCard, 0)
	// cluster & managed namespace level NewProblems, report card only has the root cause
	for _, p := range problems {
		if p.Level != UserNamespace {
			cards = append(cards, buildNewClusterReportCard(p))
		}
	}

	// user level namespace
	ucards := buildNewReportCards(problems, client)
	for _, val := range ucards {
		val.RootCause = rootNewCause(val.Resources)
		// set ID
		val.ID = hashcode(val.TopResourceType + "/" + val.Name)
		cards = append(cards, val)
	}
	// Sort makes sure the cluster level report card is the first one,
	// then sort by id
	sort.Slice(cards, func(i, j int) bool {
		if cards[i].Level == cards[j].Level {
			return cards[i].ID < cards[j].ID
		}
		return cards[i].Level < cards[j].Level
	})
	return cards, nil
}

// Build report card for cluster level or managed namespace level
func buildNewClusterReportCard(p NewProblem) *ReportCard {
	resources := []*ReportCardResource{}
	var kind string
	var rootCause *ReportCardIssue

	res := getNewReportCardResource(p, p.AffectedResources)
	resources = append(resources, res)
	if rootCause == nil {
		kind = p.AffectedResources.Resource.GetObjectKind().GroupVersionKind().Kind
		rootCause = res.Issue
	}
	return &ReportCard{
		Name:            p.Description,
		Level:           p.Level,
		Resources:       resources,
		TopResourceType: kind,
		ID:              hashcode(kind + "/" + p.Description),
		RootCause:       rootCause,
	}
}

func buildNewReportCards(problems []NewProblem, client *kubeclient.KubeClient) map[string]*ReportCard {
	cards := make(map[string]*ReportCard)
	for _, p := range problems {
		// ignore cluster & managed namespace level NewProblems
		if p.Level != UserNamespace {
			continue
		}
		switch v := p.AffectedResources.Resource.(type) {
		case metav1.Object:
			top, h := getNewTopResource(v, client)
			cr := getNewReportCardResource(p, p.AffectedResources)
			if h != nil {
				if rd, ok := cards[h.toString()]; ok {
					rd.Resources = append(rd.Resources, cr)
				} else {
					cards[h.toString()] = &ReportCard{
						Name:            h.toString(),
						Level:           p.Level,
						Resources:       []*ReportCardResource{cr},
						TopResourceType: "Helm",
					}
				}
			} else {
				if rd, ok := cards[top.GetName()]; ok {
					rd.Resources = append(rd.Resources, cr)
				} else {
					card := &ReportCard{
						Name:      top.GetName(),
						Level:     p.Level,
						Resources: []*ReportCardResource{cr},
					}
					if obj, ok := top.(runtime.Object); ok {
						card.TopResourceType = obj.GetObjectKind().GroupVersionKind().Kind
					}
					cards[top.GetName()] = card
				}
			}
		default:
			// TODO log
		}
	}
	return cards
}

func rootNewCause(res []*ReportCardResource) *ReportCardIssue {
	var root *ReportCardIssue
	rootlevel := 100
	causelevelmap := make(map[int]*ReportCardResource)
	for _, r := range res {
		causelevelmap[r.Issue.CauseLevel] = r
		if r.Issue.CauseLevel < rootlevel {
			rootlevel = r.Issue.CauseLevel
		}
	}
	root = causelevelmap[rootlevel].Issue
	return root
}

// getNewHelmChart returns the helm chart info if
func getNewHelmChart(meta metav1.Object) *helmChart {
	chart := helmChart{
		instance: meta.GetLabels()["app.kubernetes.io/instance"],
		version:  meta.GetLabels()["app.kubernetes.io/version"],
		chart:    meta.GetLabels()["helm.sh/chart"],
		release:  meta.GetAnnotations()["meta.helm.sh/release-name"],
	}
	return &chart
}

// getNewTopResource returns the top resource for the specified resource,
// e.g. Deployment --> ReplicaSet --> Pod, so the top resource for Pod is Deployment
// if any level of resource has helm chart info, then returns helm
func getNewTopResource(mo metav1.Object, client *kubeclient.KubeClient) (metav1.Object, *helmChart) {
	chart := getNewHelmChart(mo)
	if !chart.isEmpty() {
		return nil, chart
	}
	oref := getNewControlOwner(mo)
	// if there is no parent resource
	if oref == nil {
		return mo, nil
	}
	owner, err := client.GetOwner(context.TODO(), *oref, mo.GetNamespace())
	if err != nil {
		fmt.Printf("Failed to get owner resource from owner reference, %v", err)
		// return the resource itself if cannot get its owner
		return mo, nil
	}
	return getNewTopResource(owner, client)
}

// Assume only 1 owner which controls the resource
func getNewControlOwner(mo metav1.Object) *metav1.OwnerReference {
	if mo.GetOwnerReferences() == nil {
		return nil
	}
	for _, owner := range mo.GetOwnerReferences() {
		if *owner.Controller {
			return &owner
		}
	}
	return nil
}

func getNewReportCardResource(p NewProblem, resource NewResourceDetails) *ReportCardResource {
	cr := createNewReportCardResource(p, resource.Resource.(metav1.Object), resource.Resource.GetObjectKind().GroupVersionKind().Kind)
	for _, s := range p.SolutionDetails {
		cr.Issue.Solutions = append(cr.Issue.Solutions, *s)
	}
	// cr.Issue.Documents = urlToStr(p.Docs)
	// if resource.Deeplink != nil {
	// 	links := make(map[string]string)
	// 	for k, v := range resource.Deeplink {
	// 		links[string(k)] = v.String()
	// 	}
	// 	cr.Deeplink = links
	// }
	return cr
}

func createNewReportCardResource(p NewProblem, v metav1.Object, kind string) *ReportCardResource {
	issue := ReportCardIssue{
		Name:        p.Name,
		Description: p.Description,
		// Tags:        p.Tags,
		// DomainName:  p.DomainName,
		CauseLevel:  p.CauseLevel,
		CreatedTime: v.GetCreationTimestamp().String(),
	}
	return &ReportCardResource{
		Name:        v.GetName(),
		Type:        kind,
		Labels:      v.GetLabels(),
		Annotations: v.GetAnnotations(),
		Metadata:    convertMetadata(v),
		Issue:       &issue,
	}
}
