package k8s

import "time"

type K8sLogDeeplinkRetriever struct{}

type K8sEventDeeplinkRetriever struct{}

func (deeplinkReceiver K8sEventDeeplinkRetriever) GetEventDeepLink(resourceType string, clusterName string,
	namespace string, resourceName string, fromTime time.Time, endTime time.Time) string {
	return ""
}

func (deeplinkReceiver K8sEventDeeplinkRetriever) GetEventQueryString(resourceType string) string {
	return ""
}

func (deeplinkReceiver K8sLogDeeplinkRetriever) GetLogDeepLink(cluster string, namespace string, pod string,
	kubeletLog bool, kubeletLogSearchByNamespace bool, fromTime time.Time, endTime time.Time) string {
	return ""
}
