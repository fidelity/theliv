/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package problem

import (
	"context"
	"errors"
	"net/url"
	"sync"

	"k8s.io/apimachinery/pkg/runtime"
)

// Custom type for holding the supported platforms. This will be used in many places like building the
// problem execution graph for each platform etc.
type SupportedPlatform int

const (
	EKS_Platform  SupportedPlatform = iota // Amazon kubernetes service
	AKS_Platform                           // Microsoft Azure kubernetes service
	RKS_Platform                           // Rancher Kubernetes service
	GKE_Platform                           // Google Kubernetes engine
	IKS_Platform                           // IBM Kuberenetes engine
	DOKS_Platform                          // DigitalOcean Kubernetes engine
)

// String returns the string version of the supported platforms.
func (d SupportedPlatform) String() string {
	return [...]string{
		"EKS",
		"AKS",
		"RKS",
		"GKE",
		"IKS",
		"DOKS",
	}[d]
}

// Custom type for holding the Problem domains.
type DomainName int

const (
	RootFailuresDomain DomainName = iota
	ControlPlaneFailuresDomain
	NodeFailuresDomain
	MgmtNamespacesFailuresDomain
	PodFailuresDomain
	DeploymentFailuresDomain
	IngressFailuresDomain
	ServiceFailuresDomain
	end
)

// String returns the string version of the problem domains.
func (d DomainName) String() string {
	return [...]string{
		"RootFailuresDomain",
		"ControlPlaneFailuresDomain",
		"NodeFailuresDomain",
		"MgmtNamespacesFailuresDomain",
		"PodFailuresDomain",
		"DeploymentFailuresDomain",
		"IngressFailuresDomain",
		"ServiceFailuresDomain",
	}[d]
}

// DomainRegistry is a registry that will be looked up by anyone who wishes to discover various
// problem domains and the related problem detectors.
type DomainRegistry map[DomainName]*Domain

// ProblemMgr will be created once while theliv server starts, to setup problem domain wiring etc.

type ProblemMgr struct {
	domainReg DomainRegistry
}

func (p ProblemMgr) Domains() []Domain {

	domains := make([]Domain, 0, len(p.domainReg))
	for _, domain := range p.domainReg {
		domains = append(domains, *domain)
	}
	return domains
}

// Domain returns the Domain reference based on DomainName
func (p ProblemMgr) Domain(name DomainName) *Domain {
	return p.domainReg[name]
}

// DetectorRegistrationFunc return value is used by various problem detectors under
// detectors/ folder to register themselves with their problem domain.Refer
// detectors/sample/sample_detector.go for how detectors are expected to be registered.
func (p ProblemMgr) DetectorRegistrationFunc() func(DetectorRegistration, DomainName) error {

	return func(dReg DetectorRegistration, dn DomainName) error {
		registryLock.Lock()
		defer registryLock.Unlock()

		if _, f := p.domainReg[dn]; !f {
			return ErrProblemDomainNotFound
		}
		if _, z := p.domainReg[dn].detectors[dReg.Name]; z {
			return ErrDuplicateProblemDetector
		}
		// check if exists in map otherwise throw an error
		p.domainReg[dn].detectors[dReg.Name] = dReg
		return nil
	}

}

var pmgr *ProblemMgr

func DefaultProblemMgr() ProblemMgr {

	if pmgr != nil {
		return *pmgr
	}

	pmgr = &ProblemMgr{}
	domainReg := make(DomainRegistry)

	for d := DomainName(0); d < end; d++ {
		domainReg[d] = &Domain{
			Name:      d,
			detectors: make(map[DetectorName]DetectorRegistration),
		}
	}
	//statically wire the problem domain dependencies.
	// ** 	RUN DEPENDENCY **
	// These are used to run the relevant detectors in parallel in an execution graph. e.g. detectors
	// belonging to MgmtNamespacesFailures run BEFORE the detectors of PodFailureDomain can run.

	// RootFailuresDomain
	// ControlPlaneFailuresDomain   -> RootFailure
	// NodeFailuresDomain           -> RootFailure
	// MgmtNamespacesFailuresDomain -> ControlPlaneFailures, NodeFailures
	// PodFailuresDomain            -> MgmtNamespacesFailures
	// DeploymentFailuresDomain     -> MgmtNamespacesFailures
	// IngressFailuresDomain        -> MgmtNamespacesFailures

	// **	CORRELATION DEPENDENCY **
	// These are used in the aggregation logic to identify which problems are possible root causes. A
	// problem belonging to PodFailure could be a possible root cause than a problem in IngressFailureDomain

	// (Transitive dependencies automatically calculated. e.g. NodeFailuresDomain depends on BOTH
	// ControlPlaneFailuresDomain as well as RootFailuresDomain)

	// ControlPlaneFailuresDomain      ->  RootFailuresDomain
	// NodeFailuresDomain              ->  ControlPlaneFailuresDomain
	// MgmtNamespacesFailuresDomain    ->  NodeFailuresDomain
	// PodFailuresDomain               ->  MgmtNamespacesFailuresDomain
	// DeploymentFailuresDomain        ->  PodFailuresDomain
	// IngressFailuresDomain           ->  DeploymentFailuresDomain

	cDeps := map[DomainName]DomainName{
		ControlPlaneFailuresDomain:   RootFailuresDomain,
		NodeFailuresDomain:           ControlPlaneFailuresDomain,
		MgmtNamespacesFailuresDomain: NodeFailuresDomain,
		PodFailuresDomain:            MgmtNamespacesFailuresDomain,
		DeploymentFailuresDomain:     PodFailuresDomain,
		ServiceFailuresDomain:        PodFailuresDomain,
		IngressFailuresDomain:        ServiceFailuresDomain,
	}

	var calculateCorrelationDeps func(DomainName, *[]Domain)
	calculateCorrelationDeps = func(d DomainName, dn *[]Domain) {

		if _, ok := cDeps[d]; ok {
			*dn = append(*dn, *domainReg[cDeps[d]])
			// to figure out indirect/transitive dependencies.
			calculateCorrelationDeps(cDeps[d], dn)
		}
	}

	for k, v := range domainReg {
		switch k {
		case ControlPlaneFailuresDomain:
			//updating run dependencies
			v.runDeps = append(v.runDeps, *domainReg[RootFailuresDomain])

		case NodeFailuresDomain:
			//updating run dependencies
			v.runDeps = append(v.runDeps, *domainReg[RootFailuresDomain])

			//updating correlation dependencies
			c := make([]Domain, 0)
			calculateCorrelationDeps(k, &c)
			v.correlationDeps = append(v.correlationDeps, c...)

		case MgmtNamespacesFailuresDomain:
			//updating run dependencies
			v.runDeps = append(v.runDeps, *domainReg[ControlPlaneFailuresDomain])
			v.runDeps = append(v.runDeps, *domainReg[NodeFailuresDomain])

			//updating correlation dependencies
			c := make([]Domain, 0)
			calculateCorrelationDeps(k, &c)
			v.correlationDeps = append(v.correlationDeps, c...)

		case PodFailuresDomain:
			//updating run dependencies
			v.runDeps = append(v.runDeps, *domainReg[MgmtNamespacesFailuresDomain])

			//updating correlation dependencies
			c := make([]Domain, 0)
			calculateCorrelationDeps(k, &c)
			v.correlationDeps = append(v.correlationDeps, c...)

		case DeploymentFailuresDomain:
			//updating run dependencies
			v.runDeps = append(v.runDeps, *domainReg[MgmtNamespacesFailuresDomain])

			//updating correlation dependencies
			c := make([]Domain, 0)
			calculateCorrelationDeps(k, &c)
			v.correlationDeps = append(v.correlationDeps, c...)

		case IngressFailuresDomain:
			//updating run dependencies
			v.runDeps = append(v.runDeps, *domainReg[MgmtNamespacesFailuresDomain])

			//updating correlation dependencies
			c := make([]Domain, 0)
			calculateCorrelationDeps(k, &c)
			v.correlationDeps = append(v.correlationDeps, c...)

		case ServiceFailuresDomain:
			//updating run dependencies
			v.runDeps = append(v.runDeps, *domainReg[MgmtNamespacesFailuresDomain])

			//updating correlation dependencies
			c := make([]Domain, 0)
			calculateCorrelationDeps(k, &c)
			v.correlationDeps = append(v.correlationDeps, c...)

		}
	}

	pmgr.domainReg = domainReg
	return *pmgr
}

// Domain represents a problem domain. Each problem domain has a collection of problem detectors
// associated with it.
type Domain struct {
	Name            DomainName
	runDeps         []Domain
	correlationDeps []Domain
	detectors       map[DetectorName]DetectorRegistration
}

// RunAfter establishes the execution dependency relationship between problem domains.
//This dependency info will be used to decide which set of detectors needs to run first before others.
func (d Domain) RunAfter() []Domain {
	return d.runDeps
}

// CorrelationDependencies establishes the correlation dependency relationship between the problem
// domains. This will be used while trying to correlate the related issues and trying to find which
// of those issues is the root cause.
func (d Domain) CorrelationDependencies() []Domain {
	return d.correlationDeps
}

// Generic registration struct which holds basic registration info. Used in registering the
// various detectors with the registry.
type Registration struct {
	Name          DetectorName
	Description   string
	Documentation string
	Supports      []SupportedPlatform
}

// DetectorRegistration is an extension to Registration which will be used by problem detectors
// while registering with the registry. CreateFunc is the constructor function that will be
//called later whenever a detector needs to be created.
type DetectorRegistration struct {
	Registration
	CreateFunc func(input *DetectorCreationInput) (Detector, error)
}

var (
	registryLock                sync.Mutex
	ErrDuplicateProblemDetector = errors.New("problem detector already registered with the same name")
	ErrProblemDomainNotFound    = errors.New("the problem domain you are trying to register with, does not exist")
)

type DetectorName string
type DeeplinkType string

const (
	DeeplinkEvent      DeeplinkType = "event"
	DeeplinkAppLog     DeeplinkType = "applog"
	DeeplinkKubeletLog DeeplinkType = "kubeletlog"
)

// Detector represents the interface that needs to be satisfied by the detectors.
type Detector interface {
	Name() string
	Domain() DomainName
	Detect(context.Context) ([]Problem, error)
}

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
	Level             ProblemLevel
	CauseLevel        int
	SolutionDetails   []*string          // output field after detetor. It contains solutions details to show in UI.
	AffectedResources NewResourceDetails // output field after detetor. It contains the resources affected by this problem that to show in UI.
}

// To differentiate with previous ResourceDetails struct, it is named as NewResourceDetails.
// TODO: This is a temporary name, and will be renamed after old code cleanup.
type NewResourceDetails struct {
	ResourceKind string
	ResourceName string
	Resource     runtime.Object
}
