/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package problem

import (
	"net/url"

	"k8s.io/apimachinery/pkg/runtime"
)

// Custom type for holding the Problem domains.
type DomainName int

type DetectorName string
type DeeplinkType string

const (
	DeeplinkEvent      DeeplinkType = "event"
	DeeplinkAppLog     DeeplinkType = "applog"
	DeeplinkKubeletLog DeeplinkType = "kubeletlog"
)

type ResourceDetails struct {
	Resource  runtime.Object
	Deeplink  map[DeeplinkType]*url.URL
	NextSteps []string
	Details   map[string]string
}

// Problem is the struct that is returned by various detectors. Various problems returned by all the
// detectors are then aggregated to produce various report cards which is displayed to the user.
type Problem struct {
	DomainName        DomainName
	Name              string
	Description       string
	Tags              []string
	Docs              []*url.URL
	Level             ProblemLevel
	AffectedResources map[string]ResourceDetails
}

type ProblemLevel int

const (
	Cluster ProblemLevel = iota
	ManagedNamespace
	UserNamespace
)

// New Problem struct is for Prometheus alerts feature. It is the input and output struct for detectors.
// To differentiate with previous Problem struct, it is named as NewProblem.
// TODO: This is a temporary name, and will be renamed after old code cleanup.
type NewProblem struct {
	Name              string
	Description       string
	Tags              map[string]string
	Details           []string             // output field after detetor. It contains solutions details to show in UI.
	AffectedResources []NewResourceDetails // output field after detetor. It contains the resources affected by this problem that to show in UI.
}

// To differentiate with previous ResourceDetails struct, it is named as NewResourceDetails.
// TODO: This is a temporary name, and will be renamed after old code cleanup.
type NewResourceDetails struct {
	Object     runtime.Object
	ObjectKind string
	OwnerKind  string
	Owner      runtime.Object
}
