/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package observability

import (
	"context"
	"time"
)

// EventRecord represents a single event
type EventRecord struct {
	EventId        string
	Title          string
	Message        string
	Reason         string
	InvolvedObject map[string]string
	Source         map[string]string
	DateHappened   time.Time
}

type EventFilterCriteria struct {
	FilterCriteria map[string]string
	StartTime      time.Time
	EndTime        time.Time
}

type EventRetriever interface {
	Retrieve(EventFilterCriteria) EventDataRef
	AddFilters(name string, namespace string) map[string]string
}

type EventDataRef interface {
	GetEvents(ctx context.Context) ([]EventRecord, error)
}
