package bcc

import (
	"context"
	"fmt"

	"github.com/baidubce/bce-sdk-go/services/bcc"
	"github.com/baidubce/bce-sdk-go/services/bcc/api"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type stepPreValidate struct {
	SourceImageId       string
	CustomImageName     string
	SkipImageValidation bool
}

func (s *stepPreValidate) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	if s.SkipImageValidation {
		return multistep.ActionContinue
	}

	client := state.Get("client").(*bcc.Client)
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Trying to check source image id...")
	_, err := client.GetImageDetail(s.SourceImageId)
	if err != nil {
		return halt(state, err, fmt.Sprintf("The source image(id:%s) doesn't exist", s.SourceImageId))
	}

	ui.Say("Trying to check custom image name...")

	imageListResult, err := client.ListImage(&api.ListImageArgs{
		ImageName: s.CustomImageName,
		ImageType: string(api.ImageTypeCustom),
	})
	if err != nil {
		return halt(state, err, "Failed to get images info")
	}

	if len(imageListResult.Images) >= 1 {
		return halt(state, fmt.Errorf("Image name %s has exists: %+v", s.CustomImageName, imageListResult.Images), "")
	}

	return multistep.ActionContinue
}

func (s *stepPreValidate) Cleanup(multistep.StateBag) {}
