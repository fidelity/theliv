/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package problemdetectors

import (
	"sync"

	"github.com/fidelity/theliv/internal/problem"
	"github.com/fidelity/theliv/internal/problemdetectors/kubernetes"

	"github.com/fidelity/theliv/internal/problemdetectors/kubernetes/ingress"
)

var (
	lock       sync.Mutex
	registered bool
)

func Register(f func(problem.DetectorRegistration, problem.DomainName) error) {
	lock.Lock()
	defer lock.Unlock()
	if registered {
		return
	}
	// errCheck(sample.RegisterWithProblemDomain(f))
	// errCheck(test_controlplanefailure_detector.RegisterWithProblemDomain(f))
	errCheck(kubernetes.RegisterDeploymentFailureWithProblemDomain(f))
	errCheck(kubernetes.RegisterServiceFailureWithProblemDomain(f))
	errCheck(ingress.RegisterIngressEksWithProblemDomain(f))
	// errCheck(test_mgmtfailure_detector.RegisterWithProblemDomain(f))
	errCheck(kubernetes.RegisterNodeFailureWithProblemDomain(f))
	// errCheck(test_rootfailure_detector.RegisterWithProblemDomain(f))
	errCheck(kubernetes.RegisterImagePullBackOffWithProblemDomain(f))
	errCheck(kubernetes.RegisterCrashLoopBackOffWithProblemDomain(f))
	errCheck(kubernetes.RegisterPendingPodsWithProblemDomain(f))
	registered = true
}

func errCheck(e error) {
	if e != nil {
		panic(e)
	}
}
