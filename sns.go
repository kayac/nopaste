package nopaste

import (
	"context"
	"errors"
	"strings"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

func NewSNS(region string) *sns.Client {
	cfg, err := awsConfig.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic(err)
	}
	return sns.NewFromConfig(cfg)
}

func getRegionFromARN(arn string) (string, error) {
	p := strings.Split(arn, ":")
	if len(p) < 4 {
		return "", errors.New("invalid arn " + arn)
	}
	return p[3], nil
}
