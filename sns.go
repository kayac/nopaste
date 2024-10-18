package nopaste

import (
	"context"
	"errors"
	"fmt"
	"strings"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

func NewSNS(ctx context.Context, region string) (*sns.Client, error) {
	cfg, err := awsConfig.LoadDefaultConfig(ctx, awsConfig.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load config, %v", err)
	}
	return sns.NewFromConfig(cfg), nil
}

func getRegionFromARN(arn string) (string, error) {
	p := strings.Split(arn, ":")
	if len(p) < 4 {
		return "", errors.New("invalid arn " + arn)
	}
	return p[3], nil
}
