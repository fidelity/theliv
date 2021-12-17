package config

import "github.com/aws/aws-sdk-go-v2/aws"

type AwsConfig struct {
	Credentials aws.Credentials
}
