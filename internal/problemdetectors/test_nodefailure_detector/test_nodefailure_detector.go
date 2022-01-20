/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package test_nodefailure_detector

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"time"

	"github.com/fidelity/theliv/internal/problem"
)

var detectorName = "Test Detector for Node failures"

// compiler to validate if the struct indeed implements the interface
var _ problem.Detector = (*testNodeFailureDetector)(nil)

func New(i *problem.DetectorCreationInput) (problem.Detector, error) {

	return testNodeFailureDetector{
		name: detectorName,
	}, nil
}

type testNodeFailureDetector struct {
	//inputs
	//log retrieval client
	// kube client
	name string
}

func (d testNodeFailureDetector) Name() string {
	return d.name
}

func (d testNodeFailureDetector) Domain() problem.DomainName {
	return problem.DeploymentFailuresDomain
}

func (d testNodeFailureDetector) Detect(ctx context.Context) ([]problem.Problem, error) {

	fmt.Println("Running -> testNodeFailureDetector")
	time.Sleep(time.Duration(rand.Intn(5)) * time.Second)

	return []problem.Problem{
		{
			Name:        "Node Failure detected",
			Description: "Node Failure detected",
			Tags:        []string{},
			Docs:        []*url.URL{},
			DomainName:  problem.NodeFailuresDomain,
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
	}, problem.NodeFailuresDomain)

	return err
}
