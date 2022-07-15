package bcc

import "testing"

func getTestImageConfig() *BaiduCloudImageConfig {
	return &BaiduCloudImageConfig{
		ImageName: "image-name-test",
	}
}

func TestImageConfigPrepare_ImageName(t *testing.T) {
	c := getTestImageConfig()

	if errs := c.Prepare(nil); len(errs) != 0 {
		t.Fatalf("Shouldn't raise error: %s", errs)
	}

	c.ImageName = ""
	if errs := c.Prepare(nil); len(errs) != 1 {
		t.Fatalf("Should raise an error: %s", errs)
	}

}

func TestImageConfigPrepare_Regions(t *testing.T) {
	c := getTestImageConfig()

	c.DestinationRegions = []string{"bj", "gz"}
	if errs := c.Prepare(nil); len(errs) != 0 {
		t.Fatalf("Shouldn't raise error: %s", errs)
	}

	c.DestinationRegions = []string{"unknown", "bj"}
	if errs := c.Prepare(nil); len(errs) != 1 {
		t.Fatalf("Should raise an error: %s", errs)
	}

	c.DestinationRegions = []string{"bj", "bj", "su"}
	if errs := c.Prepare(nil); len(errs) != 0 {
		t.Fatalf("Shouldn't raise an error: %s", errs)
	}
	if len(c.DestinationRegions) != 2 {
		t.Fatalf("Should have two destination")
	}

	c.SkipValidation = true
	c.DestinationRegions = []string{"unknown"}
	if errs := c.Prepare(nil); len(errs) != 0 {
		t.Fatalf("Shouldn't raise error: %s", errs)
	}
}
