package bcc

import (
	"github.com/baidubce/bce-sdk-go/services/bcc"
	"github.com/baidubce/bce-sdk-go/services/eip"
	"github.com/baidubce/bce-sdk-go/services/vpc"
)

func newBccClient(ak string, sk string, endpoint string) (*bcc.Client, error) {
	return bcc.NewClient(ak, sk, endpoint)
}

func newVpcClient(ak string, sk string, endpoint string) (*vpc.Client, error) {
	return vpc.NewClient(ak, sk, endpoint)
}

func newEipClient(ak string, sk string, endpoint string) (*eip.Client, error) {
	return eip.NewClient(ak, sk, endpoint)
}
