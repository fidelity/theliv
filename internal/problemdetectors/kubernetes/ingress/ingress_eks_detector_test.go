package ingress

import (
	"fmt"
	"testing"
	"time"

	"github.com/fidelity/theliv/internal/problem"
	"github.com/stretchr/testify/assert"
	network "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	currentTimeStr     = "2021-10-12T00:00:00.00Z"
	certArn            = "test-cert-123"
	maskedArnWithSpace = " *****t-123"
)

var ingress = network.Ingress{
	ObjectMeta: metav1.ObjectMeta{
		Name: "test-ingress",
	},
}

// currentTimeStr = "2021-10-12T00:00:00.00Z"
func TestCheckCertValidPeriodValid(t *testing.T) {
	affectedResources := make(map[string]problem.ResourceDetails)
	currentTime, _ := time.Parse(time.RFC3339, currentTimeStr)
	notBefore, _ := time.Parse(time.RFC3339, "2020-10-12T00:00:00.00Z")
	notAfter, _ := time.Parse(time.RFC3339, "2022-10-12T00:00:00.00Z")
	checkCertValidPeriod(&notAfter, &notBefore, certArn, affectedResources, ingress, currentTime)
	assert.EqualValues(t, 0, len(affectedResources))
}

// currentTimeStr = "2021-10-12T00:00:00.00Z"
func TestCheckCertValidPeriodExpired(t *testing.T) {
	affectedResources := make(map[string]problem.ResourceDetails)
	currentTime, _ := time.Parse(time.RFC3339, currentTimeStr)
	notBefore, _ := time.Parse(time.RFC3339, "2020-10-12T00:00:00.00Z")
	notAfter, _ := time.Parse(time.RFC3339, "2021-10-11T00:00:00.00Z")
	checkCertValidPeriod(&notAfter, &notBefore, certArn, affectedResources, ingress, currentTime)
	assert.EqualValues(t, 1, len(affectedResources))
	res, ok := affectedResources[ingress.Name]
	assert.EqualValues(t, true, ok)
	assert.EqualValues(t, 1, len(res.Details))
	detail, ok := res.Details["Certificate expired "+certArn]
	assert.EqualValues(t, true, ok)
	assert.EqualValues(t, fmt.Sprintf("Certificate %s expired at %s", certArn, notAfter.String()), detail)
}

// currentTimeStr = "2021-10-12T00:00:00.00Z"
func TestCheckCertValidPeriodNotActive(t *testing.T) {
	affectedResources := make(map[string]problem.ResourceDetails)
	currentTime, _ := time.Parse(time.RFC3339, currentTimeStr)
	notBefore, _ := time.Parse(time.RFC3339, "2022-10-12T00:00:00.00Z")
	notAfter, _ := time.Parse(time.RFC3339, "2023-10-11T00:00:00.00Z")
	checkCertValidPeriod(&notAfter, &notBefore, certArn, affectedResources, ingress, currentTime)
	assert.EqualValues(t, 1, len(affectedResources))
	res, ok := affectedResources[ingress.Name]
	assert.EqualValues(t, true, ok)
	assert.EqualValues(t, 1, len(res.Details))
	detail, ok := res.Details["Certificate inactive "+certArn]
	assert.EqualValues(t, true, ok)
	assert.EqualValues(t, fmt.Sprintf("Certificate %s will be activate after %s", certArn,
		notBefore.String()), detail)
}

func TestValidateIngressHost(t *testing.T) {
	hosts := []string{"bar.com", "*.foo.com", "x.y.com"}
	assert.EqualValues(t, true, validateIngressHost("bar.com", hosts))
	assert.EqualValues(t, true, validateIngressHost("a.foo.com", hosts))
	assert.EqualValues(t, true, validateIngressHost("foo.bar.com", hosts))
	assert.EqualValues(t, true, validateIngressHost("a.foo.bar.com", hosts))
	assert.EqualValues(t, true, validateIngressHost("*.x.y.com", hosts))
	assert.EqualValues(t, false, validateIngressHost("invalid.com", hosts))
	assert.EqualValues(t, false, validateIngressHost("x.com", hosts))
	assert.EqualValues(t, false, validateIngressHost("y.com", hosts))
}

func TestMapToString(t *testing.T) {
	rule1 := map[string]string{"path": "p1", "serviceName": "s1"}
	rule2 := map[string]string{"serviceName": "s1", "path": "p1"}
	res1, res2 := mapToString(rule1), mapToString(rule2)
	assert.EqualValues(t, res1, res2)
	// empty or nil map should not be errored out.
	assert.EqualValues(t, mapToString(map[string]string{}), mapToString(map[string]string{}))
	assert.EqualValues(t, mapToString(nil), mapToString(nil))
}
