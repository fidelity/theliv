package kubernetes

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateEventFilterCriteria(t *testing.T) {
	filterCriteria := map[string]string{"tags": "name:pod_name", "sources": "kubernetes", "unaggregated": "false"}
	expectedFilter := map[string]string{"tags": "name:pod_name", "sources": "kubernetes", "unaggregated": "false"}
	filter := CreateEventFilterCriteria(DefaultTimespan, filterCriteria)
	assert.EqualValues(t, 3, len(filter.FilterCriteria))
	assert.EqualValues(t, true, reflect.DeepEqual(filter.FilterCriteria, expectedFilter))

}

func TestCheckPossibleErrorMessage(t *testing.T) {
	event := "Failed to pull image \"a.b.com/d-123/noimage:0.0.1\": rpc error: code = Unknown desc = Error response" +
		" from daemon: manifest for a.b.com/d-123/noimage:0.0.1 not found: manifest unknown: The named manifest is not found in registry."
	input := ProblemInput{
		PossibleErrorMessages: PossibleErrorMessages,
	}
	var msg string = UnknownManifestMsg
	assert.EqualValues(t, true, checkPossibleErrorMessage(&event, &msg, &input))

	event = "Failed to pull image \"a.b.com/d-123/noimage:0.0.1\": rpc error: code = Unknown desc = Error response" +
		" from daemon: manifest for a.b.com/d-123/noimage:0.0.1 not found: Manifest not available."
	assert.EqualValues(t, true, checkPossibleErrorMessage(&event, &msg, &input))

	event = "Failed to pull image \"a.b.com/d-123/noimage:0.0.1\": rpc error: code = Unknown desc = Error response"
	assert.EqualValues(t, false, checkPossibleErrorMessage(&event, &msg, &input))
}

func TestMaskString(t *testing.T) {
	assert.EqualValues(t, "*****", MaskString(""))
	assert.EqualValues(t, "*****", MaskString("", 5))
	assert.EqualValues(t, "*****", MaskString("test", 4))
	assert.EqualValues(t, "*****tab", MaskString("testab", 3))
	assert.EqualValues(t, "*****9", MaskString("123456789", 1))
	assert.EqualValues(t, "*****56789", MaskString("123456789"))
}
