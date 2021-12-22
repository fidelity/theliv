package observability

import (
	"time"
)

// Generate log deep link, it is independent of log retriever. User can have different implementations
// of deep link and log. The configurations can be found at configs/theliv.yaml logDeeplinkDriver and
// eventDeeplinkDriver.
type LogDeeplinkRetriever interface {
	GetLogDeepLink(cluster string, namespace string, pod string, kubeletLog bool,
		kubeletLogSearchByNamespace bool, fromTime time.Time, endTime time.Time) string
}

// Generate event deep link. Similar functions as LogDeeplinkRetriever.GetLogDeepLink.
type EventDeeplinkRetriever interface {
	GetEventDeepLink(resourceType string, clusterName string, namespace string,
		podName string, fromTime time.Time, toTime time.Time) string

	// Create the query parameter for event deep link, different kubernetes resource have different
	// query parameter.
	GetEventQueryString(resourceType string) string
}
