package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUpdatedElement(t *testing.T) {
	oldE1 := []string{"e1", "e2"}
	newE1 := []string{"e2", "e3"}
	updated1 := getUpdatedElement(oldE1, newE1)
	assert.EqualValues(t, "e1,e2,e3", strings.Join(updated1, SEPARATOR))

	oldE2 := []string{"e1", "e2"}
	newE2 := []string{"e2"}
	updated2 := getUpdatedElement(oldE2, newE2)
	assert.EqualValues(t, "e1,e2", strings.Join(updated2, SEPARATOR))
}
