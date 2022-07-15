//go:generate packer-sdc struct-markdown

package bcc

import (
	"fmt"
	"os"

	"github.com/baidubce/bce-sdk-go/services/bcc"
	"github.com/baidubce/bce-sdk-go/services/eip"
	"github.com/baidubce/bce-sdk-go/services/vpc"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

type Region string

const (
	RegionBeijing   = Region("bj")
	RegionGuangzhou = Region("gz")
	RegionSuzhou    = Region("su")
	RegionXiangGang = Region("hkg")
	RegionWuhan     = Region("fwh")
	RegionBaoding   = Region("bd")
	RegionSingapore = Region("sin")
	RegionShanghai  = Region("fsh")
)

var ValidRegions = []Region{
	RegionBeijing, RegionGuangzhou, RegionSuzhou, RegionXiangGang,
	RegionWuhan, RegionBaoding, RegionSingapore, RegionShanghai,
}

type BaiduCloudAccessConfig struct {
	// Baiducloud access key must be provided, unless the environment
	// variable `BAIDUCLOUD_ACCESS_KEY` is set
	BaiduCloudAccessKey string `mapstructure:"access_key" required:"true"`
	// Baiducloud serect key must be provided, unless the environment
	// variable `BAIDUCLOUD_SECRET_KEY` is set
	BaiduCloudSecretKey string `mapstructure:"secret_key" required:"true"`
	// Baiducloud region must be provided, unless the environment variable
	// `BAIDUCLOUD_REGION` is set.
	BaiduCloudRegion string `mapstructure:"region" required:"true"`
	// The zone where your bcc instance will be launched. It must be set,
	// unless  the field `use_default_network` is set true,
	// which means you use the default network
	Zone string `mapstructure:"zone" required:"true"`
	// Do not check region and zone when validate
	SkipValidation bool `mapstructure:"skip_region_validation" required:"false"`
}

// Client - create a client of baiducloud bcc
func (c *BaiduCloudAccessConfig) Client() (*bcc.Client, error) {
	return newBccClient(c.BaiduCloudAccessKey, c.BaiduCloudSecretKey, c.GetBccEndpoint())
}

// VpcClient - create a client of baiducloud vpc
func (c *BaiduCloudAccessConfig) VpcClient() (*vpc.Client, error) {
	return newVpcClient(c.BaiduCloudAccessKey, c.BaiduCloudSecretKey, c.GetVpcEndpoint())
}

// EipClient - create a client of baiducloud eip
func (c *BaiduCloudAccessConfig) EipClient() (*eip.Client, error) {
	return newEipClient(c.BaiduCloudAccessKey, c.BaiduCloudSecretKey, c.GetEipEndpoint())
}

// ClientWithRegion - create a bcc client for specified region
func (c *BaiduCloudAccessConfig) ClientWithRegion(region string) (*bcc.Client, error) {
	return newBccClient(c.BaiduCloudAccessKey, c.BaiduCloudSecretKey, c.GetBccEndpointWithRegion(region))
}

func (c *BaiduCloudAccessConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error
	if err := c.Config(); err != nil {
		errs = append(errs, err)
	}

	if c.BaiduCloudRegion == "" {
		c.BaiduCloudRegion = os.Getenv("BAIDUCLOUD_REGION")
	}

	if c.BaiduCloudRegion == "" {
		errs = append(errs, fmt.Errorf("region option or BAIDUCLOUD_REGION must be provided in template file or environment variables"))
	} else if !c.SkipValidation {
		if err := c.validateRegion(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

func (c *BaiduCloudAccessConfig) Config() error {
	if c.BaiduCloudAccessKey == "" {
		c.BaiduCloudAccessKey = os.Getenv("BAIDUCLOUD_ACCESS_KEY")
	}
	if c.BaiduCloudSecretKey == "" {
		c.BaiduCloudSecretKey = os.Getenv("BAIDUCLOUD_SECRET_KEY")
	}
	if c.BaiduCloudAccessKey == "" || c.BaiduCloudSecretKey == "" {
		return fmt.Errorf("parameter access_key and secret_key must be provided in template file or environment variables")
	}
	return nil
}

// GetEndpoint - get the endpoint of baidu cloud server of specific region
func (c *BaiduCloudAccessConfig) GetBccEndpoint() string {
	return fmt.Sprintf("https://bcc.%s.baidubce.com", c.BaiduCloudRegion)
}

func (c *BaiduCloudAccessConfig) GetEipEndpoint() string {
	return fmt.Sprintf("https://eip.%s.baidubce.com", c.BaiduCloudRegion)
}

func (c *BaiduCloudAccessConfig) GetVpcEndpoint() string {
	return fmt.Sprintf("https://bcc.%s.baidubce.com", c.BaiduCloudRegion)
}

func (c *BaiduCloudAccessConfig) GetBccEndpointWithRegion(region string) string {
	return fmt.Sprintf("https://bcc.%s.baidubce.com", region)
}

func (c *BaiduCloudAccessConfig) validateRegion() error {
	return validRegion(c.BaiduCloudRegion)
}

func validRegion(region string) error {
	for _, validRegion := range ValidRegions {
		if Region(region) == validRegion {
			return nil
		}
	}

	return fmt.Errorf("unknown region: %s", region)
}
