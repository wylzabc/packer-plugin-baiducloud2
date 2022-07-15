//go:generate packer-sdc struct-markdown

package bcc

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

type BaiduCloudImageConfig struct {
	// The name you want to create your customize image,
	// it supports upper and lower case letters, numbers, Chinese
	// and -_/. special characters,
	// which must start with a letter and be 1-65 in length
	ImageName string `mapstructure:"image_name" required:"true"`
	// Copy the custom image created by build steps to destination regions
	DestinationRegions []string `mapstructure:"image_copy_regions" required:"false"`
	// Account names of baiducloud to share images to, which
	// must be a baidu account
	ImageShareAccounts []string `mapstructure:"image_share_accounts" required:"false"`
	// The account ids of baiducloud to share image
	ImageShareAccountIds []string `mapstructure:"image_share_account_ids" required:"false"`
	// Do not check region and zone when validate
	SkipValidation bool `mapstructure:"skip_region_validation" required:"false"`
	// The image validation can be skipped if this value is true, the default
	// value is false.
	SkipImageValidation bool `mapstructure:"skip_image_validation" required:"false"`
}

func (c *BaiduCloudImageConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error

	if c.ImageName == "" {
		errs = append(errs, errors.New("image_name must be specified"))
	} else if len(c.ImageName) > 65 {
		errs = append(errs, fmt.Errorf("image_name must less than or equal to 65 letters"))
	} else {
		matchFormat := `^[a-zA-Z][0-9a-zA-Z-_/.]*$`
		reg, err := regexp.Compile(matchFormat)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to create regular expression: %s", matchFormat))
		}
		if !reg.MatchString(c.ImageName) {
			errs = append(errs, fmt.Errorf("image_name has a format error: %s", c.ImageName))
		}
	}

	// Remove duplicate regions
	if len(c.DestinationRegions) > 0 {
		regionSet := make(map[string]struct{})
		regions := make([]string, 0, len(c.DestinationRegions))

		for _, region := range c.DestinationRegions {
			if _, ok := regionSet[region]; ok {
				continue
			}

			regionSet[region] = struct{}{}

			if !c.SkipValidation {
				if err := validRegion(region); err != nil {
					errs = append(errs, err)
					continue
				}
			}
			regions = append(regions, region)
		}

		c.DestinationRegions = regions
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}
