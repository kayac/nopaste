package nopaste

import (
	"os"

	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/sns"
)

func NewSNS() *sns.SNS {
	auth := aws.Auth{
		AccessKey: os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
	}
	region := aws.GetRegion(os.Getenv("AWS_REGION"))
	s, err := sns.New(auth, region)
	if err != nil {
		panic(err)
	}
	return s
}
