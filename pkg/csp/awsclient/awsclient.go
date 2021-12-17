package awsclient

import (
	"context"

	tag "github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
)

func GetTagByArn(ctx context.Context, client *tag.Client, arn *string) (*tag.GetResourcesOutput, error) {
	tagRequestInput := tag.GetResourcesInput{
		ResourceARNList: []string{*arn},
	}
	tagOutput, err := client.GetResources(ctx, &tagRequestInput)
	return tagOutput, err
}
