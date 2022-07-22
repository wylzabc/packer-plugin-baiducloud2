package bcc

import (
	"context"
	"fmt"

	"github.com/baidubce/bce-sdk-go/services/bcc"
	"github.com/baidubce/bce-sdk-go/services/bcc/api"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/uuid"
)

type stepCreateImage struct {
	imageId string
}

func (s *stepCreateImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*bcc.Client)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Starting to create custom image...")

	var createImageResult *api.CreateImageResult
	err := Retry(ctx, func(ctx context.Context) error {
		var e error
		createImageResult, e = client.CreateImage(s.getCreateImageArgs(state))
		return e
	})
	if err != nil {
		return halt(state, err, "Failed to creating image")
	}

	imageId := createImageResult.ImageId
	ui.Say(fmt.Sprintf("Waiting to image(%s) status available...", imageId))
	err = WaitForImage(ctx, client, imageId, api.ImageStatusAvailable, 1800)
	if err != nil {
		return halt(state, err, fmt.Sprintf("Failed to creating image(%s)", imageId))
	}
	state.Put("image_id", imageId)
	s.imageId = imageId

	baiduCloudImage := make(map[string]string)
	baiduCloudImage[config.BaiduCloudRegion] = imageId
	state.Put("baiducloud_images", baiduCloudImage)

	return multistep.ActionContinue
}

func (s *stepCreateImage) Cleanup(state multistep.StateBag) {
	if s.imageId == "" {
		return
	}

	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)

	if !cancelled && !halted {
		return
	}

	cleanUpMessage(state, "image")

	client := state.Get("client").(*bcc.Client)
	ui := state.Get("ui").(packersdk.Ui)

	ctx := context.TODO()
	err := Retry(ctx, func(ctx context.Context) error {
		return client.DeleteImage(s.imageId)
	})
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to clean up image %s: %s", s.imageId, err))
	}
}

func (s *stepCreateImage) getCreateImageArgs(state multistep.StateBag) *api.CreateImageArgs {
	config := state.Get("config").(*Config)
	instanceId := state.Get("instance_id").(string)
	return &api.CreateImageArgs{
		ImageName:  config.ImageName,
		InstanceId: instanceId,
		// IsRelateCds: true,
		ClientToken: uuid.TimeOrderedUUID(),
	}
}
