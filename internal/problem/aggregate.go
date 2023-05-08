/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package problem

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"sort"
	"strings"

	com "github.com/fidelity/theliv/pkg/common"
	"github.com/fidelity/theliv/pkg/kubeclient"
	log "github.com/fidelity/theliv/pkg/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Aggregate(ctx context.Context, problems []Problem, client *kubeclient.KubeClient) (interface{}, error) {
	cards := make([]*ReportCard, 0)
	// cluster & managed namespace level NewProblems, report card only has the root cause
	for _, p := range problems {
		if p.Level != UserNamespace {
			cards = append(cards, buildClusterReportCard(p))
		}
	}

	// user level namespace
	ucards := buildReportCards(ctx, problems, client)
	for _, val := range ucards {
		val.RootCause = rootCause(val.Resources)
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
func buildClusterReportCard(p Problem) *ReportCard {
	resources := []*ReportCardResource{}
	var kind string
	var rootCause *ReportCardIssue

	res := getReportCardResource(p, p.AffectedResources)
	resources = append(resources, res)
	if rootCause == nil {
		kind = p.AffectedResources.ResourceKind
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

func buildReportCards(ctx context.Context, problems []Problem, client *kubeclient.KubeClient) map[string]*ReportCard {
	cards := make(map[string]*ReportCard)
	for _, p := range problems {
		// ignore cluster & managed namespace level NewProblems
		if p.Level != UserNamespace {
			continue
		}
		switch v := p.AffectedResources.Resource.(type) {
		case metav1.Object:
			top, h, argo := getTopResource(ctx, v, client)
			cr := getReportCardResource(p, p.AffectedResources)
			if argo != nil {
				appendCards(cards, cr, p, argo.Instance, com.Argo)
			} else if h != nil {
				appendCards(cards, cr, p, h.toString(), com.Helm)
			} else {
				topType := ""
				if obj, ok := top.(runtime.Object); ok {
					topType = obj.GetObjectKind().GroupVersionKind().Kind
				}
				appendCards(cards, cr, p, top.GetName(), topType)
			}
		default:
			// TODO log
		}
	}
	return cards
}

func rootCause(res []*ReportCardResource) *ReportCardIssue {
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

// getHelmChart returns the helm chart info if
func getHelmChart(meta metav1.Object) *helmChart {
	chart := helmChart{
		instance: meta.GetLabels()["app.kubernetes.io/instance"],
		version:  meta.GetLabels()["app.kubernetes.io/version"],
		chart:    meta.GetLabels()["helm.sh/chart"],
		release:  meta.GetAnnotations()["meta.helm.sh/release-name"],
	}
	return &chart
}

// getTopResource returns the top resource for the specified resource,
// e.g. Deployment --> ReplicaSet --> Pod, so the top resource for Pod is Deployment
// if any level of resource has Argo Instance info, then returns Argo Instance.
// if any level of resource has helm chart info, then returns helm
func getTopResource(ctx context.Context, mo metav1.Object, client *kubeclient.KubeClient) (metav1.Object, *helmChart, *ArgoInstance) {
	argo := getArgoInstance(mo)
	if argo.Instance != "" {
		return nil, nil, argo
	}
	if argo.RolloutTemplate == "" {
		chart := getHelmChart(mo)
		if !chart.isEmpty() {
			return nil, chart, nil
		}
	}
	oref := getControlOwner(mo)
	// if there is no parent resource
	if oref == nil {
		return mo, nil, nil
	}
	owner, err := client.GetOwner(ctx, *oref, mo.GetNamespace())
	if err != nil {
		fmt.Printf("Failed to get owner resource from owner reference, %v", err)
		// return the resource itself if cannot get its owner
		return mo, nil, nil
	}
	return getTopResource(ctx, owner, client)
}

// Assume only 1 owner which controls the resource
func getControlOwner(mo metav1.Object) *metav1.OwnerReference {
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

func getReportCardResource(p Problem, resource ResourceDetails) *ReportCardResource {
	cr := createReportCardResource(p, resource.Resource.(metav1.Object), resource.ResourceKind)
	cr.Issue.Solutions = append(cr.Issue.Solutions, p.SolutionDetails...)

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

func createReportCardResource(p Problem, v metav1.Object, kind string) *ReportCardResource {
	issue := ReportCardIssue{
		Name:        p.Name,
		Description: p.Description,
		Tags:        p.Tags,
		// DomainName:  p.DomainName,
		CauseLevel:  p.CauseLevel,
		CreatedTime: v.GetCreationTimestamp().String(),
	}
	name := v.GetName()
	if strings.Contains(p.Name, "Container") {
		name = p.Tags["container"]
	}
	return &ReportCardResource{
		Name:        name,
		Type:        kind,
		Labels:      v.GetLabels(),
		Annotations: v.GetAnnotations(),
		Metadata:    convertMetadata(v),
		Issue:       &issue,
	}
}

func hashcode(s string) string {
	v := int(crc32.ChecksumIEEE([]byte(s)))
	if v < 0 {
		v = -v
	}
	return fmt.Sprint(v)
}

func convertMetadata(obj metav1.Object) map[string]interface{} {
	b, err := json.Marshal(obj)
	if err != nil {
		log.S().Errorf("Marshal json error: %s", err)
	}
	m := make(map[string]interface{})
	err = json.Unmarshal(b, &m)
	if err != nil {
		log.S().Errorf("Unmarshal json error: %s", err)
	}
	return cleanFieldNotRequired(m)
}

// toString returns the helm chart release, it checks annotation first
func (chart *helmChart) toString() string {
	// annotation is managed by Helm itself, but there is no version info
	if chart.release != "" {
		return chart.release
	}
	// if helm.sh/chart is present, returns it as helm chart
	// format should be release-version
	if chart.chart != "" {
		return chart.chart
	}
	// if labels are not present, returns annotation "meta.helm.sh/release-name"
	if chart.instance == "" || chart.version == "" {
		return ""
	}
	return chart.instance + "-" + chart.version
}

// Check if the helm chart is empty
func (chart *helmChart) isEmpty() bool {
	return chart.toString() == ""
}

func cleanFieldNotRequired(data map[string]interface{}) map[string]interface{} {
	removeFields := []string{"selfLink", "uid", "resourceVersion", "creationTimestamp", "managedFields"}

	if meta, ok := data["metadata"]; ok {
		switch v := meta.(type) {
		case map[string]interface{}:
			for _, f := range removeFields {
				delete(v, f)
			}
		}
	}
	return data
}

// If card exists, append to card.Resources, or append new card into whole cards.
func appendCards(cards map[string]*ReportCard, cr *ReportCardResource, p Problem, name string, topType string) {
	if rd, ok := cards[name]; ok {
		rd.Resources = append(rd.Resources, cr)
	} else {
		cards[name] = &ReportCard{
			Name:            name,
			Level:           p.Level,
			Resources:       []*ReportCardResource{cr},
			TopResourceType: topType,
		}
	}
}

// Returns the ArgoInstance info if exists.
func getArgoInstance(meta metav1.Object) *ArgoInstance {
	return &ArgoInstance{
		Instance:        meta.GetLabels()["argocd.argoproj.io/instance"],
		RolloutTemplate: meta.GetLabels()["rollouts-pod-template-hash"],
	}
}
