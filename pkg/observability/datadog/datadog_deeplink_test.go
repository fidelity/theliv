package datadog

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var datadogHost = "https://test.com"
var retr = DatadogEventDeeplinkRetriever{
	DatadogHost: datadogHost,
}
var retrLog = DatadogLogDeeplinkRetriever{
	DatadogHost: datadogHost,
}

func TestGetEventDeepLink(t *testing.T) {
	e := retr.GetEventDeepLink("pod_type", "cluster", "namespace", "pod", time.Now(), time.Now())
	assert.NotNil(t, e)
	assert.True(t, strings.HasPrefix(e, datadogHost))
	assert.True(t, strings.Contains(e, "pod_type"))
	assert.False(t, strings.Contains(e, "%s"))

	retr2 := DatadogEventDeeplinkRetriever{}
	e2 := retr2.GetEventDeepLink("pod_type", "cluster", "namespace", "pod", time.Now(), time.Now())
	assert.NotNil(t, e2)
	assert.True(t, strings.Contains(e, "pod_type"))
	assert.False(t, strings.Contains(e, "%s"))
}

func TestGetEventDeepLinkEmpty(t *testing.T) {
	e := retr.GetEventDeepLink("", "", "", "", time.Now(), time.Now())
	assert.NotNil(t, e)
	assert.True(t, strings.Contains(e, "tags"))
	assert.False(t, strings.Contains(e, "%s"))

	retr2 := DatadogEventDeeplinkRetriever{}
	e2 := retr2.GetEventDeepLink("", "", "", "", time.Now(), time.Now())
	assert.NotNil(t, e2)
	assert.True(t, strings.Contains(e2, "tags"))
	assert.False(t, strings.Contains(e2, "%s"))
}

func TestGetLogDeepLink(t *testing.T) {
	e := retrLog.GetLogDeepLink("cluster_name", "namespace_test", "pod", true, true, time.Now(), time.Now())
	assert.NotNil(t, e)
	assert.True(t, strings.HasPrefix(e, datadogHost))
	assert.True(t, strings.Contains(e, "namespace_test"))
	assert.True(t, strings.Contains(e, "cluster_name"))
	assert.False(t, strings.Contains(e, "%s"))

	retrLog2 := DatadogLogDeeplinkRetriever{}
	e2 := retrLog2.GetLogDeepLink("cluster_name", "namespace_test", "pod", true, false, time.Now(), time.Now())
	assert.NotNil(t, e2)
	assert.True(t, strings.Contains(e2, "cluster_name"))
	assert.False(t, strings.Contains(e2, "namespace_test"))
	assert.False(t, strings.Contains(e2, "%s"))
}

func TestGetLogDeepLinkEmpty(t *testing.T) {
	e := retrLog.GetLogDeepLink("", "", "", true, true, time.Now(), time.Now())
	assert.NotNil(t, e)

	retrLog2 := DatadogLogDeeplinkRetriever{}
	e2 := retrLog2.GetLogDeepLink("", "", "", true, false, time.Now(), time.Now())
	assert.NotNil(t, e2)
}
