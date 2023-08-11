/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package service

import (
	"context"
	"fmt"

	invest "github.com/fidelity/theliv/internal/investigators"
	com "github.com/fidelity/theliv/pkg/common"
	"github.com/fidelity/theliv/pkg/config"
	errors "github.com/fidelity/theliv/pkg/err"
	"github.com/fidelity/theliv/pkg/kubeclient"
	log "github.com/fidelity/theliv/pkg/log"
	"github.com/fidelity/theliv/pkg/observability/k8s"
)

func GetEvents(ctx context.Context) (interface{}, error) {
	input := GetDetectorInput(ctx)

	contact := fmt.Sprintf(com.Contact, config.GetThelivConfig().TeamName)

	client, err := kubeclient.NewKubeClient(input.Kubeconfig)
	if err != nil {
		return nil, errors.NewCommonError(ctx, 4, com.LoadKubeConfigFailed+contact)
	}
	input.KubeClient = client

	eventRetriever := k8s.NewK8sEventRetriever(client)
	input.EventRetriever = eventRetriever

	filter := invest.CreateEventFilterCriteria(invest.DefaultTimespan,
		input.EventRetriever.AddFilters("", input.Namespace))
	eventDataRef := input.EventRetriever.Retrieve(filter)

	events, err := eventDataRef.GetEvents(ctx)
	if err != nil {
		log.SWithContext(ctx).Error("Got error when calling kubernetes event API, error is %s", err)
		return nil, errors.NewCommonError(ctx, 4, com.LoadEventsFailed+contact)
	}
	return events, nil
}
