package datadog

import (
	"context"
	"fmt"
	golog "log"
	"net/url"
	"strconv"
	"time"

	datadog "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	observability "github.com/fidelity/theliv/pkg/observability"
)

const (
	EventParameterPriority         string = "priority"
	EventParameterSources          string = "sources"
	EventParameterTags             string = "tags"
	EventParameterUnaggregated     string = "unaggregated"
	EventParameterExcludeaggregate string = "excludeAggregate"
	EventParameterPage             string = "page"
	EventOperationPath             string = "/event/stream?query="
	EventTimeStampQuery            string = "&from_ts=%d&to_ts=%d&aggregate_up=true"
	EventOtherQuery                string = "&live=false"
)

type DatadogEventDataRef struct {
	apiClient                    datadog.APIClient
	ListEventsOptionalParameters datadog.ListEventsOptionalParameters
	FilterCriteria               observability.EventFilterCriteria
	config                       DatadogConfig
}

// Prepare request content for operations in LogDataRef.
// Reference https://docs.datadoghq.com/api/latest/events/
// Please note, currently event use datadog v1 endpoint, this is different with log API.
func (eventReceiver DatadogEventRetriever) Retrieve(filterCriteria observability.EventFilterCriteria) observability.EventDataRef {
	golog.Println("INFO - Get event from DataDog")

	priority := datadog.EventPriority(filterCriteria.FilterCriteria[EventParameterPriority])
	sources := filterCriteria.FilterCriteria[EventParameterSources]
	tags := filterCriteria.FilterCriteria[EventParameterTags]
	unaggregated, _ := strconv.ParseBool(filterCriteria.FilterCriteria[EventParameterUnaggregated])
	excludeAggregate, _ := strconv.ParseBool(filterCriteria.FilterCriteria[EventParameterExcludeaggregate])
	page, _ := strconv.Atoi(filterCriteria.FilterCriteria[EventParameterPage])
	pageInt32 := int32(page)

	optionalParams := datadog.ListEventsOptionalParameters{
		Unaggregated:     &unaggregated,
		ExcludeAggregate: &excludeAggregate,
	}

	if pageInt32 > 0 {
		optionalParams.Page = &pageInt32
	}
	if len(priority) > 0 {
		optionalParams.Priority = &priority
	}
	if len(sources) > 0 {
		optionalParams.Sources = &sources
	}
	if len(tags) > 0 {
		optionalParams.Tags = &tags
	}

	return NewDatadogEventRef(optionalParams, eventReceiver.datadogConfig, filterCriteria)
}

var DefaultEventFilterCriteria = map[string]string{
	EventParameterSources:      "kubernetes",
	EventParameterUnaggregated: "true",
}

func (eventReceiver DatadogEventRetriever) AddFilters(name string, namespace string) map[string]string {
	consolidatedFilterCriteria := make(map[string]string)
	filterCriteria := map[string]string{EventParameterTags: "name:" + name}
	MergeMap(DefaultEventFilterCriteria, consolidatedFilterCriteria)
	MergeMap(filterCriteria, consolidatedFilterCriteria)
	return consolidatedFilterCriteria
}

// Add key and value from source map into target map, overwrite exising key and value.
func MergeMap(source map[string]string, target map[string]string) {
	for k, v := range source {
		target[k] = v
	}
}

// Query Datadog backend.
// https://github.com/DataDog/datadog-api-client-go/blob/master/api/v2/datadog/docs/LogsApi.md#listlogs
func (dataRef DatadogEventDataRef) GetEvents(ctx context.Context) ([]observability.EventRecord, error) {
	ctx = context.WithValue(
		ctx,
		datadog.ContextAPIKeys,
		map[string]datadog.APIKey{
			"apiKeyAuth": {
				Key: dataRef.config.ClientApiKey,
			},
			"appKeyAuth": {
				Key: dataRef.config.ClientAppkey,
			},
		},
	)
	start := dataRef.FilterCriteria.StartTime.Unix()
	end := dataRef.FilterCriteria.EndTime.Unix()
	eventsListResponse, _, err := dataRef.apiClient.EventsApi.ListEvents(ctx, start, end,
		dataRef.ListEventsOptionalParameters)
	events := make([]observability.EventRecord, 0)
	if err != nil {
		golog.Printf("ERROR - %v\n", err)
		return events, err
	}
	for _, event := range *eventsListResponse.Events {
		eventRecord := observability.EventRecord{}
		eventRecord.Title = *event.Title
		eventRecord.Message = *event.Text
		eventRecord.Metadata = AddTagsToMetadata(*event.Tags)
		eventRecord.EventId = fmt.Sprint(*event.Id)
		eventRecord.DateHappend = time.Unix(*event.DateHappened, 0)
		if url, err := url.Parse(dataRef.config.DatadogHost + *event.Url); err != nil {
			golog.Printf("WARN - Event url generation failed %s", err)
		} else {
			eventRecord.DeepLink = *url
		}
		events = append(events, eventRecord)
	}
	return events, nil
}

type DatadogEventRetriever struct {
	datadogConfig DatadogConfig
}

func NewDatadogEventRetriever(config DatadogConfig) DatadogEventRetriever {
	return DatadogEventRetriever{
		datadogConfig: config,
	}
}

func NewDatadogEventRef(listEventsOptionalParameters datadog.ListEventsOptionalParameters,
	datadogConfig DatadogConfig, filterCriteria observability.EventFilterCriteria) DatadogEventDataRef {
	configuration := datadog.NewConfiguration()
	configuration.Debug = datadogConfig.Debug
	apiClient := datadog.NewAPIClient(configuration)
	return DatadogEventDataRef{
		apiClient:                    *apiClient,
		ListEventsOptionalParameters: listEventsOptionalParameters,
		FilterCriteria:               filterCriteria,
		config:                       datadogConfig,
	}
}
