package nopaste

import (
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
)

func NewSNS(region string) *sns.SNS {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))
	return sns.New(sess)
}

func getRegionFromARN(arn string) (string, error) {
	p := strings.Split(arn, ":")
	if len(p) < 4 {
		return "", errors.New("invalid arn " + arn)
	}
	return p[3], nil
}
