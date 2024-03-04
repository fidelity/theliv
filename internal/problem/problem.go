/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package problem

import (
	"github.com/fidelity/theliv/pkg/common"
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

type ProblemLevel int

const (
	Cluster ProblemLevel = iota
	ManagedNamespace
	UserNamespace
)

// Problem struct is for Prometheus alerts feature. It is the input and output struct for detectors.
type Problem struct {
	Name              string
	Description       string
	Tags              map[string]string
	Level             ProblemLevel
	CauseLevel        int
	SolutionDetails   *common.LockedSlice // output field after detetor. It contains solutions details to show in UI.
	UsefulCommands    *common.LockedSlice // output field after detetor. It contains solutions details to show in UI.
	AiSuggestions     *common.LockedSlice
	AiKnowledge       *common.LockedSlice
	AffectedResources ResourceDetails // output field after detetor. It contains the resources affected by this problem that to show in UI.
}

type ResourceDetails struct {
	ResourceKind string
	ResourceName string
	Resource     runtime.Object
}
