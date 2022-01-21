/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package datadog

import (
	"fmt"
	"net/url"
	"time"
)

const (
	DataDogLogPath         = "/logs?query="
	DataDogLogClusterQuery = "cluster:%s "
	DataDogLogNsPoQuery    = "kube_namespace:%s pod_name:%s"
	DataDogLogParam        = "&from_ts=%d&to_ts=%d&live=false"
	EventQueryPrefix       = "sources:kubernetes tags:cluster:%s,kube_namespace:%s,"
	EventQuerySuffix       = ":%s status:all priority:all"
)

type DatadogLogDeeplinkRetriever struct {
	DatadogHost string
}

type DatadogEventDeeplinkRetriever struct {
	DatadogHost string
}

func (deeplinkReceiver DatadogEventDeeplinkRetriever) GetEventDeepLink(resourceType string, clusterName string,
	namespace string, resourceName string, fromTime time.Time, endTime time.Time) string {

	sourceQuery := fmt.Sprintf(deeplinkReceiver.GetEventQueryString(resourceType), clusterName, namespace,
		resourceName)
	fromTs := fromTime.Unix() * TimeScale
	toTs := endTime.Unix() * TimeScale
	timeStamp := fmt.Sprintf(EventTimeStampQuery, fromTs, toTs)

	return deeplinkReceiver.DatadogHost + EventOperationPath + url.QueryEscape(sourceQuery) +
		timeStamp + EventOtherQuery
}

func (deeplinkReceiver DatadogEventDeeplinkRetriever) GetEventQueryString(resourceType string) string {
	return EventQueryPrefix + resourceType + EventQuerySuffix
}

// Return deep link of datadog logs, from_ts usually is pod creation time, to_ts is Now().
func (deeplinkReceiver DatadogLogDeeplinkRetriever) GetLogDeepLink(cluster string, namespace string, pod string,
	kubeletLog bool, kubeletLogSearchByNamespace bool, fromTime time.Time, endTime time.Time) string {

	var query string
	if kubeletLog {
		if kubeletLogSearchByNamespace {
			query = fmt.Sprintf(DataDogLogClusterQuery, cluster) + pod + "_" + namespace
		} else {
			query = fmt.Sprintf(DataDogLogClusterQuery, cluster) + pod
		}
	} else {
		query = fmt.Sprintf(DataDogLogClusterQuery+DataDogLogNsPoQuery, cluster, namespace, pod)
	}
	fromTs := fromTime.Unix() * TimeScale
	toTs := endTime.Unix() * TimeScale
	param := fmt.Sprintf(DataDogLogParam, fromTs, toTs)

	return deeplinkReceiver.DatadogHost + DataDogLogPath + url.QueryEscape(query) + param
}
