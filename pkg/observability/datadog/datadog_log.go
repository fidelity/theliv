/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package datadog

import (
	"context"
	"fmt"
	golog "log"
	"net/url"
	"time"

	datadog "github.com/DataDog/datadog-api-client-go/api/v2/datadog"
	observability "github.com/fidelity/theliv/pkg/observability"
)

type DatadogLogDataRef struct {
	apiClient                  datadog.APIClient
	ListLogsOptionalParameters datadog.ListLogsOptionalParameters
	FilterCriteria             observability.LogFilterCriteria
	config                     DatadogConfig
}

// Prepare request content for operations in LogDataRef.
// Reference https://docs.datadoghq.com/api/latest/logs/
// Please note, currently log use datadog v2 endpoint, this is different with event API.
func (logReceiver DatadogLogRetriever) Retrieve(filterCriteria observability.LogFilterCriteria) observability.LogDataRef {
	golog.Println("INFO - get log from DataDog")
	request := datadog.NewLogsListRequest()
	filter := datadog.NewLogsQueryFilter()
	// In Datadog index is sensitive, so here use it from config instead of user provided in query.
	filter.Indexes = &[]string{logReceiver.datadogLogConfig.AppId}
	regularExpression := filterCriteria.RegularExpression
	query := CreateQueryFromMap(filterCriteria.FilterCriteria) + regularExpression
	filter.Query = &query
	filter.SetFrom(filterCriteria.StartTime.Format(time.RFC3339))
	filter.SetTo(filterCriteria.EndTime.Format(time.RFC3339))
	request.Filter = filter
	page := datadog.NewLogsListRequestPage()
	page.SetLimit(logReceiver.datadogLogConfig.MaxRecords)
	request.Page = page

	optionalParams := datadog.ListLogsOptionalParameters{
		Body: request,
	}
	logDataRef := NewDatadogLogRef(optionalParams, logReceiver.datadogLogConfig, filterCriteria)
	return logDataRef
}

// Query Datadog backend.
// https://github.com/DataDog/datadog-api-client-go/blob/master/api/v2/datadog/docs/LogsApi.md#listlogs
func (logDataRef DatadogLogDataRef) GetRecords(ctx context.Context) ([]observability.LogRecord, error) {
	ctx = context.WithValue(
		ctx,
		datadog.ContextAPIKeys,
		map[string]datadog.APIKey{
			"apiKeyAuth": {
				Key: logDataRef.config.ClientApiKey,
			},
			"appKeyAuth": {
				Key: logDataRef.config.ClientAppkey,
			},
		},
	)
	logsListResponse, _, err := logDataRef.apiClient.LogsApi.ListLogs(ctx, logDataRef.ListLogsOptionalParameters)
	logs := make([]observability.LogRecord, 0)
	if err != nil {
		golog.Printf("ERROR - %v\n", err)
		return logs, err
	}
	for _, log := range logsListResponse.GetData() {
		logRecord := observability.LogRecord{}
		logRecord.Message = log.Attributes.GetMessage()
		logRecord.Metadata = AddTagsToMetadata(log.Attributes.GetTags())
		deepLink, err := CreateDeepLink(logDataRef, log)
		if err != nil {
			golog.Printf("ERROR - deeplink creation failed, %s", err)
		} else {
			logRecord.DeepLink = deepLink
		}
		logs = append(logs, logRecord)
	}
	return logs, nil
}

type DatadogLogRetriever struct {
	datadogLogConfig DatadogConfig
}

func NewDatadogLogRetriever(logConfig DatadogConfig) DatadogLogRetriever {
	return DatadogLogRetriever{
		datadogLogConfig: logConfig,
	}
}

// Count() use different request and filter struct which are different with other operations,
// pass FilterCriteria to Count() to create required struct.
func NewDatadogLogRef(listLogsOptionalParameters datadog.ListLogsOptionalParameters,
	datadogLogConfig DatadogConfig, filterCriteria observability.LogFilterCriteria) DatadogLogDataRef {

	configuration := datadog.NewConfiguration()
	configuration.Debug = datadogLogConfig.Debug
	apiClient := datadog.NewAPIClient(configuration)
	return DatadogLogDataRef{
		apiClient:                  *apiClient,
		ListLogsOptionalParameters: listLogsOptionalParameters,
		FilterCriteria:             filterCriteria,
		config:                     datadogLogConfig,
	}
}

// Create log link for reference.
func CreateDeepLink(logDataRef DatadogLogDataRef, log datadog.Log) (url.URL, error) {
	return createDeepLinkHelper(logDataRef.config.DatadogHost, *log.Id, *log.Attributes.Host,
		logDataRef.config.AppId, logDataRef.FilterCriteria.StartTime, logDataRef.FilterCriteria.EndTime)
}

// Follow Datadog UI log syntax to create a link for each log.
func createDeepLinkHelper(datadogHost, eventId, host, index string, fromTime, toTime time.Time) (url.URL, error) {
	fromTimeInt, toTimeInt := fromTime.Unix(), toTime.Unix()

	rawUrl := fmt.Sprintf("%s/logs?cols=core_host%%2Ccore_service&context_event=%s"+
		"&event=&from_ts=%d&index=%s&live=false&messageDisplay=inline&"+
		"query=host%%3A%s+service%%3Aec2messages+filename%%3Amessages&saved_view&"+
		"stream_sort=desc&to_event=%s&to_ts=%d&viz=", datadogHost, eventId, fromTimeInt,
		index, host, eventId, toTimeInt)
	deeplink, err := url.Parse(rawUrl)
	if err != nil {
		return url.URL{}, err
	}
	return *deeplink, nil
}
