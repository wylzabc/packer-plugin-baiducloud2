package bcc

import (
	"os"
	"testing"
)

func getTestBaiduCloudAccessConfig() *BaiduCloudAccessConfig {
	return &BaiduCloudAccessConfig{
		BaiduCloudAccessKey: "ak",
		BaiduCloudSecretKey: "sk",
	}
}

func TestBaiduCloudAccessConfigPrepare_Region(t *testing.T) {
	c := getTestBaiduCloudAccessConfig()

	if v := os.Getenv("BAIDUCLOUD_REGION"); v != "" {
		os.Unsetenv("BAIDUCLOUD_REGION")
		defer os.Setenv("BAIDUCLOUD_REGION", v)
	}

	c.BaiduCloudRegion = ""
	if errs := c.Prepare(nil); errs == nil {
		t.Fatal("Should raise error: region must be set")
	}

	c.BaiduCloudRegion = "fwh"
	if errs := c.Prepare(nil); errs != nil {
		t.Fatalf("Shouldn't raise error: %v", errs)
	}

	c.BaiduCloudRegion = "unknown"
	if errs := c.Prepare(nil); errs == nil {
		t.Fatal("Should raise error: unknown region")
	}

	c.SkipValidation = true
	if errs := c.Prepare(nil); errs != nil {
		t.Fatalf("Shouldn't raise error: %s", errs)
	}

	os.Setenv("BAIDUCLOUD_REGION", "fwh")
	c.BaiduCloudRegion = ""
	if errs := c.Prepare(nil); errs != nil {
		t.Fatalf("Shouldn't raise error: %s", errs)
	}

}
