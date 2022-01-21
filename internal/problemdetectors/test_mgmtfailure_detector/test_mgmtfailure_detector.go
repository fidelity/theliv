/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package test_mgmtfailure_detector

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"time"

	"github.com/fidelity/theliv/internal/problem"
)

var detectorName = "Test Detector for Mgmt namespaces failure"

// compiler to validate if the struct indeed implements the interface
var _ problem.Detector = (*testMgmtFailureDetector)(nil)

func New(i *problem.DetectorCreationInput) (problem.Detector, error) {

	return testMgmtFailureDetector{
		name: detectorName,
	}, nil
}

type testMgmtFailureDetector struct {
	//inputs
	//log retrieval client
	// kube client
	name string
}

func (d testMgmtFailureDetector) Name() string {
	return d.name
}

func (d testMgmtFailureDetector) Domain() problem.DomainName {
	return problem.DeploymentFailuresDomain
}

func (d testMgmtFailureDetector) Detect(ctx context.Context) ([]problem.Problem, error) {

	fmt.Println("Running -> testMgmtFailureDetector")
	time.Sleep(time.Duration(rand.Intn(5)) * time.Second)

	return []problem.Problem{
		{
			Name:        "Mgmt namespace Failure detected",
			Description: "Mgmt namespace Failure detected",
			Tags:        []string{},
			Docs:        []*url.URL{},
			DomainName:  problem.MgmtNamespacesFailuresDomain,
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
	}, problem.MgmtNamespacesFailuresDomain)

	return err

}
