/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package k8s

import (
	"context"
	"time"

	"github.com/fidelity/theliv/pkg/kubeclient"
	observability "github.com/fidelity/theliv/pkg/observability"
)

func (logReceiver K8sLogRetriever) Retrieve(filterCriteria observability.LogFilterCriteria) observability.LogDataRef {
	return nil
}

func (logDataRef K8sLogDataRef) GetRecords(ctx context.Context) ([]observability.LogRecord, error) {
	return nil, nil
}

type K8sLogRetriever struct {
	kubeclient *kubeclient.KubeClient
}

type K8sLogDataRef struct{}

func NewK8sLogRetriever(kubeclient *kubeclient.KubeClient) K8sLogRetriever {
	return K8sLogRetriever{kubeclient}
}

func (logReceiver K8sLogRetriever) GetLogDeepLink(cluster string, namespace string, pod string, kubeletLog bool,
	kubeletLogSearchByNamespace bool, fromTime time.Time, endTime time.Time) string {
	return ""
}
