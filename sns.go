package nopaste

import (
	"errors"
	"os"
	"strings"

	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/sns"
)

func NewSNS(r string) *sns.SNS {
	auth := aws.Auth{
		AccessKey: os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
	}
	if r == "" {
		r = os.Getenv("AWS_REGION")
	}
	region := aws.GetRegion(r)
	s, err := sns.New(auth, region)
	if err != nil {
		panic(err)
	}
	return s
}

func getRegionFromARN(arn string) (string, error) {
	p := strings.Split(arn, ":")
	if len(p) < 4 {
		return "", errors.New("invalid arn " + arn)
	}
	return p[3], nil
}
