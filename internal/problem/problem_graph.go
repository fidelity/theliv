package problem

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/rajarajanpsj/terraform/dag"
	"github.com/rajarajanpsj/terraform/tfdiags"
)

type graphExecutor interface {
	Execute(context.Context) (*GraphExecutionResults, error)
}

type dotGetter interface {
	DotRepresentation() string
}

// compiler to validate if the struct indeed implements the interface
var _ graphExecutor = (*DefaultProblemGraph)(nil)
var _ dotGetter = (*DefaultProblemGraph)(nil)

type DefaultProblemGraph struct {
	dag.AcyclicGraph
	problemDomains []Domain
	//aws client TODO
	// kubeclient TODO
	// log retrieval client TODO
}

type Namespace string
type ClusterName string

func NewDefaultProblemGraph(d []Domain, detInput *DetectorCreationInput) (DefaultProblemGraph, error) {

	gc := DefaultProblemGraph{
		problemDomains: d,
	}

	// could not do make(map[Domain][]Detector) because Domain struct cant be used as map key.
	prbDetMap := make(map[DomainName][]Detector)

	for _, domain := range gc.problemDomains {
		for _, detectorReg := range domain.detectors {
			if slicecontains(detectorReg.Registration.Supports, detInput.Platform) {
				det, err := detectorReg.CreateFunc(detInput)
				if err != nil {
					return gc, fmt.Errorf("unable to create Detector instance: %w", err)
				}
				gc.Add(det)
				prbDetMap[domain.Name] = append(prbDetMap[domain.Name], det)
			}
		}
	}

	for _, domain := range gc.problemDomains {
		for _, domainDep := range domain.runDeps {
			// DEBUG purposes : fmt.Printf("source: %s -> target: %s", domain.Name, domainDep.Name)
			gc.connectEdges(edgeSources(prbDetMap[domain.Name]), edgeTargets(prbDetMap[domainDep.Name]))
		}
	}

	return gc, nil

}

type edgeSources []Detector
type edgeTargets []Detector

func (pg DefaultProblemGraph) connectEdges(sources edgeSources, targets edgeTargets) {
	for _, sourceDetector := range sources {
		for _, targetDetector := range targets {
			pg.Connect(dag.BasicEdge(sourceDetector, targetDetector))
		}
	}
}

func (pg DefaultProblemGraph) DotRepresentation() string {

	return string(pg.Dot(&dag.DotOpts{
		DrawCycles: true,
	}))
}

func (pg DefaultProblemGraph) Execute(ctx context.Context) (*GraphExecutionResults, error) {

	results := &GraphExecutionResults{
		ProblemMap: make(map[string][]Problem),
		DotGraph:   pg.DotRepresentation(),
	}

	walkFn := func(v dag.Vertex) (diags tfdiags.Diagnostics) {

		if det, ok := v.(Detector); ok {
			ps, err := det.Detect(ctx)
			if err != nil {
				// add errors to diags. MUST TODO
				results.Errors = append(results.Errors, err)
			}
			results.put(det.Name(), ps)
		} else {
			diags = diags.Append(diags, fmt.Errorf("problem graph only support Detector interface for Vertices, received %s", reflect.TypeOf(v).Name()))
		}
		return
	}

	e := pg.Walk(walkFn)
	return results, e.Err()
}

// TODO: check if this needs to use sync.Map rather
type GraphExecutionResults struct {
	mtx        sync.Mutex
	ProblemMap map[string][]Problem
	Errors     []error
	DotGraph   string
}

// Detector name as the key
func (m *GraphExecutionResults) put(k string, v []Problem) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	m.ProblemMap[k] = v
}

func (m *GraphExecutionResults) Get(k string) ([]Problem, bool) {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	v, ok := m.ProblemMap[k]
	return v, ok
}

func slicecontains(s []SupportedPlatform, e SupportedPlatform) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
