package sample

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"time"

	"github.com/fidelity/theliv/internal/problem"
)

var detectorName = "Sample Detector"

// compiler to validate if the struct indeed implements the interface
var _ problem.Detector = (*imagePullDetector)(nil)

func New(i *problem.DetectorCreationInput) (problem.Detector, error) {

	return imagePullDetector{
		name: detectorName,
	}, nil
}

type imagePullDetector struct {
	//inputs
	//log retrieval client
	// kube client
	name string
}

func (d imagePullDetector) Name() string {
	return d.name
}

func (d imagePullDetector) Domain() problem.DomainName {
	return problem.DeploymentFailuresDomain
}

func (d imagePullDetector) Detect(ctx context.Context) ([]problem.Problem, error) {

	fmt.Println("Running -> imagePullDetector")
	time.Sleep(time.Duration(rand.Intn(5)) * time.Second)

	return []problem.Problem{
		{
			DomainName:        0,
			Name:              "",
			Description:       "",
			Tags:              []string{},
			Docs:              []*url.URL{},
			AffectedResources: map[string]problem.ResourceDetails{},
		},
	}, nil
}

func RegisterWithProblemDomain(regFunc func(problem.DetectorRegistration, problem.DomainName) error) error {

	// future: Detector can evaluate certain conditions and decide whether to register or not. e.g if
	// metrics driver is not available or not enabled, detector can decide not to register itself.

	err := regFunc(problem.DetectorRegistration{
		Registration: problem.Registration{
			Name:        problem.DetectorName(detectorName),
			Description: "This is to detector so and so problem, blah blah",
			Documentation: `A detailed markdown MD string that details the 
							 documentation for this problem detector`,
			Supports: []problem.SupportedPlatform{problem.EKS_Platform, problem.AKS_Platform},
		},
		CreateFunc: New,
	}, problem.PodFailuresDomain)

	return err
}
