package test_controlplanefailure_detector

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"time"

	"github.com/fidelity/theliv/internal/problem"
)

var detectorName = "Test Detector for Control Plane"

// compiler to validate if the struct indeed implements the interface
var _ problem.Detector = (*testControlPlaneFailureDetector)(nil)

func New(i *problem.DetectorCreationInput) (problem.Detector, error) {

	return testControlPlaneFailureDetector{
		name: detectorName,
	}, nil
}

type testControlPlaneFailureDetector struct {
	//inputs
	//log retrieval client
	// kube client
	name string
}

func (d testControlPlaneFailureDetector) Name() string {
	return d.name
}

func (d testControlPlaneFailureDetector) Domain() problem.DomainName {
	return problem.DeploymentFailuresDomain
}

func (d testControlPlaneFailureDetector) Detect(ctx context.Context) ([]problem.Problem, error) {

	fmt.Println("Running -> testControlPlaneFailureDetector")
	time.Sleep(time.Duration(rand.Intn(5)) * time.Second)

	return []problem.Problem{
		{
			Name:        "Control Plane Failure detected",
			Description: "Control Plane Failure detected",
			Tags:        []string{},
			Docs:        []*url.URL{},
			DomainName:  problem.ControlPlaneFailuresDomain,
			//AffectedResources: map[problem.ResourceType]func() (ResourceDetails, ResourceTags),
		},
	}, nil
}

func RegisterWithProblemDomain(regFunc func(problem.DetectorRegistration, problem.DomainName) error) error {

	err := regFunc(problem.DetectorRegistration{
		Registration: problem.Registration{
			Name:        problem.DetectorName(detectorName),
			Description: "This is to detector so and so problem, blah blah",
			Documentation: `A detailed markdown MD string that details the 
							 documentation for this problem detector`,
			Supports: []problem.SupportedPlatform{problem.EKS_Platform, problem.AKS_Platform},
		},
		CreateFunc: New,
	}, problem.ControlPlaneFailuresDomain)

	return err
}
