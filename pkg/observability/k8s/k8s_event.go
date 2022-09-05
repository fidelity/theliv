/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package k8s

import (
	"context"
	"strings"

	com "github.com/fidelity/theliv/pkg/common"
	"github.com/fidelity/theliv/pkg/kubeclient"
	observability "github.com/fidelity/theliv/pkg/observability"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type K8sEventRetriever struct {
	kubeclient *kubeclient.KubeClient
}

type K8sEventDataRef struct {
	K8sEventRetriever
	observability.EventFilterCriteria
}

// Return the instance of EventDataRef, with k8sClient, and filtering conditions set.
func (eventReceiver K8sEventRetriever) Retrieve(filterCriteria observability.EventFilterCriteria) observability.EventDataRef {
	return K8sEventDataRef{eventReceiver, filterCriteria}
}

/*
This function will call the k8s API to retrieve the events. Use predefined filtering conditions.
In FilterCriteria, if namespace is provided, will get Events only under the specified namespace,
otherwise get events of all the namespaces.
In FilterCriteria, if resource name is provided, will do filtering after retrieving all the events,
if the resource name can be found in event.Message or event.InvolvedObject.Name.
In FilterCriteria, StartTime and EndTime won't be used for filtering.
*/
func (dataRef K8sEventDataRef) GetEvents(ctx context.Context) ([]observability.EventRecord, error) {

	eventRecord := make([]observability.EventRecord, 0)

	var name string
	if data, ok := dataRef.FilterCriteria[com.Name]; ok {
		name = data
	}
	namespace := kubeclient.NamespacedName{}
	if data, ok := dataRef.FilterCriteria[com.Namespace]; ok {
		namespace = kubeclient.NamespacedName{Namespace: data}
	}

	events := &v1.EventList{}
	eventsOptions := metav1.ListOptions{}
	dataRef.kubeclient.List(ctx, events, namespace, eventsOptions)

	for _, event := range events.Items {
		if strings.Contains(event.InvolvedObject.Name, name) || strings.Contains(event.Message, name) {
			eventRecord = append(eventRecord,
				observability.EventRecord{
					EventId:        string(event.ObjectMeta.UID),
					Title:          event.ObjectMeta.Name,
					Message:        event.Message,
					Reason:         event.Reason,
					DateHappened:   event.ObjectMeta.CreationTimestamp.Time,
					InvolvedObject: getInvolvedObject(event.InvolvedObject),
					Source:         getSource(event.Source),
				})
		}
	}
	return eventRecord, nil

}

// Get Info from event.EventSource, returns map[string]string.
func getSource(source v1.EventSource) map[string]string {
	data := initMap()
	data["Component"] = source.Component
	return data
}

// Get Info from event.InvolvedObject, returns map[string]string.
func getInvolvedObject(obj v1.ObjectReference) map[string]string {
	data := initMap()
	data[com.Name] = obj.Name
	data[com.Namespace] = obj.Namespace
	data["Kind"] = obj.Kind
	data["UID"] = string(obj.UID)
	data["APIVersion"] = obj.APIVersion
	return data
}

// Default filter, add k8s resource Name and Namespace.
func (eventReceiver K8sEventRetriever) AddFilters(name string, namespace string) map[string]string {
	return map[string]string{com.Name: name, com.Namespace: namespace}
}

// New for K8sEventRetriever.
func NewK8sEventRetriever(kubeclient *kubeclient.KubeClient) K8sEventRetriever {
	return K8sEventRetriever{kubeclient}
}

func initMap() map[string]string {
	return map[string]string{}
}
