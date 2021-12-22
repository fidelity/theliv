package test_rootfailure_detector

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"time"

	"github.com/fidelity/theliv/internal/problem"
)

var detectorName = "Test Detector for RootFailure"

// compiler to validate if the struct indeed implements the interface
var _ problem.Detector = (*testRootFailureDetector)(nil)

func New(i *problem.DetectorCreationInput) (problem.Detector, error) {

	return testRootFailureDetector{
		name: detectorName,
	}, nil
}

type testRootFailureDetector struct {
	//inputs
	//log retrieval client
	// kube client
	name string
}

func (d testRootFailureDetector) Name() string {
	return d.name
}

func (d testRootFailureDetector) Domain() problem.DomainName {
	return problem.DeploymentFailuresDomain
}

func (d testRootFailureDetector) Detect(ctx context.Context) ([]problem.Problem, error) {

	fmt.Println("Running -> testRootFailureDetector")
	time.Sleep(time.Duration(rand.Intn(5)) * time.Second)
	return []problem.Problem{
		{
			Name:        "Root Failure detected",
			Description: "Root Failure detected",
			Tags:        []string{},
			Docs:        []*url.URL{},
			DomainName:  problem.RootFailuresDomain,
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
	}, problem.RootFailuresDomain)

	return err
}
