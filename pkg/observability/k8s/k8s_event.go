package k8s

import (
	"context"
	"strings"
	"time"

	"github.com/fidelity/theliv/pkg/kubeclient"
	observability "github.com/fidelity/theliv/pkg/observability"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	Name      = "name"
	Namespace = "namespace"
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
	if data, ok := dataRef.FilterCriteria[Name]; ok {
		name = data
	}
	namespace := kubeclient.NamespacedName{}
	if data, ok := dataRef.FilterCriteria[Namespace]; ok {
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
					DateHappend:    event.ObjectMeta.CreationTimestamp.Time,
					Metadata:       getMetadata(event),
					InvolvedObject: getInvolvedObject(event.InvolvedObject),
					Source:         getSource(event.Source),
				})
		}
	}
	return eventRecord, nil

}

// Get Info from evnet.EventSource, returns map[string]string.
func getSource(source v1.EventSource) map[string]string {
	data := initMap()
	data["Component"] = source.Component
	data["Host"] = source.Host
	return data
}

// Get Info from event.InvolvedObject, returns map[string]string.
func getInvolvedObject(obj v1.ObjectReference) map[string]string {
	data := initMap()
	data["Name"] = obj.Name
	data["Namespace"] = obj.Namespace
	data["Kind"] = obj.Kind
	data["UID"] = string(obj.UID)
	data["APIVersion"] = obj.APIVersion
	return data
}

// Get Info from event.ObjectMeta, returns map[string]string.
func getMetadata(event v1.Event) map[string]string {
	data := initMap()
	data["Name"] = event.ObjectMeta.Name
	data["Namespace"] = event.ObjectMeta.Namespace
	data["SelfLink"] = event.ObjectMeta.SelfLink
	data["UID"] = string(event.ObjectMeta.UID)
	data["Reason"] = event.Reason
	return data
}

// Default filter, add k8s resource Name and Namespace.
func (eventReceiver K8sEventRetriever) AddFilters(name string, namespace string) map[string]string {
	return map[string]string{Name: name, Namespace: namespace}
}

// New for K8sEventRetriever.
func NewK8sEventRetriever(kubeclient *kubeclient.KubeClient) K8sEventRetriever {
	return K8sEventRetriever{kubeclient}
}

func initMap() map[string]string {
	return map[string]string{}
}

func (eventReceiver K8sEventRetriever) GetEventsDeepLink(clusterName string, namespace string,
	podName string, fromTime time.Time, endTime time.Time) string {
	return ""
}

func (eventReceiver K8sEventRetriever) GetResourceEventsDeepLink(resourceQuery string, clusterName string, namespace string,
	resourceName string, fromTime time.Time, endTime time.Time) string {
	return ""
}
