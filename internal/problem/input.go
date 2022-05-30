/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package problem

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"k8s.io/client-go/rest"
)

type TimeSpan struct {
	Timespan     int
	TimespanType time.Duration
}

type DetectorCreationInput struct {
	Kubeconfig    *rest.Config
	Namespace     string
	ClusterName   string
	EventTimespan TimeSpan
	LogTimespan   TimeSpan
	AwsConfig     aws.Config
}
