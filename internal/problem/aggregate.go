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

	"github.com/fidelity/theliv/pkg/kubeclient"
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
	ucards := buildReportCards(problems, client)
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

func buildReportCards(problems []Problem, client *kubeclient.KubeClient) map[string]*ReportCard {
	cards := make(map[string]*ReportCard)
	for _, p := range problems {
		// ignore cluster & managed namespace level NewProblems
		if p.Level != UserNamespace {
			continue
		}
		switch v := p.AffectedResources.Resource.(type) {
		case metav1.Object:
			top, h := getTopResource(v, client)
			cr := getReportCardResource(p, p.AffectedResources)
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
// if any level of resource has helm chart info, then returns helm
func getTopResource(mo metav1.Object, client *kubeclient.KubeClient) (metav1.Object, *helmChart) {
	chart := getHelmChart(mo)
	if !chart.isEmpty() {
		return nil, chart
	}
	oref := getControlOwner(mo)
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
	return getTopResource(owner, client)
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
	cr := createReportCardResource(p, resource.Resource.(metav1.Object), resource.Resource.GetObjectKind().GroupVersionKind().Kind)
	for _, s := range p.SolutionDetails {
		cr.Issue.Solutions = append(cr.Issue.Solutions, s)
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

func createReportCardResource(p Problem, v metav1.Object, kind string) *ReportCardResource {
	issue := ReportCardIssue{
		Name:        p.Name,
		Description: p.Description,
		Tags:        p.Tags,
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
		// TODO log
	}
	m := make(map[string]interface{})
	err = json.Unmarshal(b, &m)
	if err != nil {
		// TODO
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
