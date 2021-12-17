package datadog

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateDeepLink(t *testing.T) {
	datadogHost, eventId, host, index := "https://fake.datadog.com", "event-1", "i-fake-instance",
		"ap-test"
	to := time.Now()
	from := to.Add(time.Hour * -8)
	val, err := createDeepLinkHelper(datadogHost, eventId, host, index, from, to)
	assert.Nil(t, err)
	assert.EqualValues(t, "https://fake.datadog.com/logs?cols=core_host%2Ccore_service&context_event="+
		"event-1&event=&from_ts="+fmt.Sprint(from.Unix())+"&index=ap-test&live=false&messageDisplay=inline"+
		"&query=host%3Ai-fake-instance+service%3Aec2messages+filename%3Amessages&saved_view&"+
		"stream_sort=desc&to_event=event-1&to_ts="+fmt.Sprint(to.Unix())+"&viz=", val.String())

}

func TestMergeMap(t *testing.T) {
	defaultMap, customMap := map[string]string{"a": "1", "b": "2"}, map[string]string{"b": "3", "c": "4"}
	allMap := map[string]string{}
	expectedMap := map[string]string{"a": "1", "b": "3", "c": "4"}
	MergeMap(defaultMap, allMap)
	assert.EqualValues(t, 2, len(allMap))
	assert.EqualValues(t, true, reflect.DeepEqual(allMap, defaultMap))
	MergeMap(customMap, allMap)
	assert.EqualValues(t, 3, len(allMap))
	assert.EqualValues(t, true, reflect.DeepEqual(allMap, expectedMap))
}
