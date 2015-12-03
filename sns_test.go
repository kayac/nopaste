package nopaste

import "testing"

func TestGetRegionFromARN(t *testing.T) {
	arn1 := "arn:aws:sns:us-east-1:999999999:example"
	r1, _ := getRegionFromARN(arn1)
	if r1 != "us-east-1" {
		t.Errorf("invalid region %s from %s", r1, arn1)
	}

	arn2 := "arn:aws:sns"
	r2, err := getRegionFromARN(arn2)
	if r2 != "" || err == nil {
		t.Errorf("must be failed %s from %s", r2, arn2)
	}
}
