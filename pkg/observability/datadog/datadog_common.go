package datadog

import (
	"bytes"
	"fmt"
	"strings"
)

const (
	DATADOG_TAG_SEPARATOR string = ":"
	TimeScale             int64  = 1000
)

type DatadogConfig struct {
	ClientApiKey string
	ClientAppkey string
	AppId        string
	MaxRecords   int32
	Debug        bool
	DatadogHost  string
}

// Datadog related configurations.
func NewDatadogLogConfigWithDebug(clientApikey string, clientAppKey string, appId string,
	maxRecords int32, debug bool) DatadogConfig {
	return DatadogConfig{
		ClientApiKey: clientApikey,
		ClientAppkey: clientAppKey,
		AppId:        appId,
		MaxRecords:   maxRecords,
		Debug:        debug,
	}
}

func NewDatadogLogConfig(clientApikey string, clientAppKey string, appId string, from string, to string,
	maxRecords int32) DatadogConfig {
	return NewDatadogLogConfigWithDebug(clientApikey, clientAppKey, appId, maxRecords, false)
}

// Datadog stores tag with single string, convert it to map.
func AddTagsToMetadata(tags []string) map[string]string {
	result := make(map[string]string)

	for _, tag := range tags {
		if strings.Contains(tag, DATADOG_TAG_SEPARATOR) {
			data := strings.Split(tag, DATADOG_TAG_SEPARATOR)
			length := len(data)
			if length == 2 {
				result[data[0]] = data[1]
			} else if length < 2 {
				result[tag] = ""
			} else {
				result[data[0]] = strings.Join(data[1:], DATADOG_TAG_SEPARATOR)
			}
		} else {
			result[tag] = ""
		}
	}
	return result
}

// Convert the map in FilterCriteria to query string, separated by ":", follow Datadog search syntax.
func CreateQueryFromMap(m map[string]string) string {
	if len(m) == 0 {
		return ""
	}
	query := new(bytes.Buffer)
	for key, value := range m {
		if len(value) == 0 {
			fmt.Fprint(query, key, " ")
		} else {
			fmt.Fprintf(query, "%s:%s ", key, value)
		}
	}
	return query.String()
}
