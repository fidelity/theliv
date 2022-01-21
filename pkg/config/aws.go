/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package config

import "github.com/aws/aws-sdk-go-v2/aws"

type AwsConfig struct {
	Credentials aws.Credentials
}
