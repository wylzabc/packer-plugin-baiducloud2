package bcc

import (
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

func testBuilderConfig() map[string]interface{} {
	return map[string]interface{}{
		"access_key":      "ak-test",
		"secret_key":      "sk-test",
		"source_image_id": "source-image-id-test",
		"instance_spec":   "bcc.ic4.c2m2",
		"image_name":      "image-name-test",
		"region":          "fwh",
		"zone":            "cn-fwh-a",
		"ssh_username":    "root",
	}
}

func TestBuilder_ImplementsBuilder(t *testing.T) {
	var raw interface{}
	raw = &Builder{}
	if _, ok := raw.(packersdk.Builder); !ok {
		t.Fatalf("Builder should be a builder")
	}
}

func TestBuilderPrepare_BadType(t *testing.T) {
	b := &Builder{}
	c := map[string]interface{}{
		"access_key": []string{},
	}

	_, warnings, err := b.Prepare(c)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatal("Prepare should fail")
	}
}

func TestBuilderPrepare_ImageName(t *testing.T) {
	var b Builder
	config := testBuilderConfig()

	// Test good case
	config["image_name"] = "foo"
	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("Shouldn't raise error: %s", err)
	}

	// Test bad case
	config["image_name"] = "foo bar"
	b = Builder{}
	_, warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatal("Should raise error")
	}

	// Test bad case
	delete(config, "image_name")
	b = Builder{}
	_, warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatal("Should raise error")
	}
}

func TestBuidlerPrepare_InvalidKey(t *testing.T) {
	var b Builder
	config := testBuilderConfig()

	// Add a random key
	config["i_should_not_be_valid"] = true
	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatalf("Should raise error")
	}
}
