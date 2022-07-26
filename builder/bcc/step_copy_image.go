package bcc

import (
	"context"
	"fmt"
	"strings"

	"github.com/baidubce/bce-sdk-go/services/bcc"
	"github.com/baidubce/bce-sdk-go/services/bcc/api"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type stepRemoteCopyImage struct {
	DestinationRegions []string
	SourceRegion       string
}

func (s *stepRemoteCopyImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	if len(s.DestinationRegions) == 0 || (len(s.DestinationRegions) == 1 && s.DestinationRegions[0] == s.SourceRegion) {
		return multistep.ActionContinue
	}

	//config := state.Get("config").(*Config)
	client := state.Get("client").(*bcc.Client)
	imageId := state.Get("image_id").(string)
	ui := state.Get("ui").(packersdk.Ui)

	// copy image to remote region
	remoteCopyImageArgs := s.getRemoteCopyImageArgs(state)

	ui.Say(fmt.Sprintf("Trying to copy image(%s) to region: %s", imageId, strings.Join(remoteCopyImageArgs.DestRegion, ",")))

	var remoteCopyResult *api.RemoteCopyImageResult
	err := Retry(ctx, func(ctx context.Context) error {
		var e error
		remoteCopyResult, e = client.RemoteCopyImageReturnImageIds(imageId, remoteCopyImageArgs)
		return e
	})
	if err != nil {
		return halt(state, err, "Failed to copy image")
	}

	// waiting image available after remote image copy
	ui.Say("Waiting for image ready...")
	if err := WaitForImage(ctx, client, imageId, api.ImageStatusAvailable, 1800); err != nil {
		panic(err)
	}
	// time.Sleep(10 * time.Second)

	// record the mapping of region to image id of custom image made by the build process
	baiduCloudImages := state.Get("baiducloud_images").(map[string]string)
	for _, image := range remoteCopyResult.RemoteCopyImages {
		baiduCloudImages[image.Region] = image.ImageId
	}

	state.Put("baiducloud_images", baiduCloudImages)

	return multistep.ActionContinue
}

func (s *stepRemoteCopyImage) Cleanup(state multistep.StateBag) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)

	if !cancelled && !halted {
		return
	}

	config := state.Get("config").(*Config)
	baiduCloudImages := state.Get("baiducloud_images").(map[string]string)
	ui := state.Get("ui").(packersdk.Ui)

	ctx := context.TODO()

	ui.Say("Trying to cancel remote copy or deleting images had been copied...")

	for region, imageId := range baiduCloudImages {
		if region == config.BaiduCloudRegion {
			continue
		}

		client, err := config.ClientWithRegion(region)
		if err != nil {
			ui.Error(fmt.Sprintf("Failed to get bcc client of region(%s)", region))
			continue
		}

		imageDetail, err := client.GetImageDetail(imageId)
		if err != nil {
			ui.Error(fmt.Sprintf("Failed to get image detail(%s)", imageId))
			continue
		}

		// cancel remote copy
		if imageDetail.Image.Status != api.ImageStatusAvailable {
			// cancel remote copy image
			if err := client.CancelRemoteCopyImage(imageId); err == nil {
				ui.Message(fmt.Sprintf("Success to cancel remote copy image(%s) to region(%s)", imageId, region))
				continue
			}
		}
		// delete remote image
		err = Retry(ctx, func(ctx context.Context) error {
			return client.DeleteImage(imageId)
		})
		if err != nil {
			ui.Error(fmt.Sprintf("Failed to delete image(%s) of region(%s)", imageId, region))
			continue
		}
		ui.Message(fmt.Sprintf("Success to delete image(%s) of region(%s)", imageId, region))
	}
}

func (s *stepRemoteCopyImage) getRemoteCopyImageArgs(state multistep.StateBag) *api.RemoteCopyImageArgs {
	config := state.Get("config").(*Config)

	var destRegion []string
	for _, region := range s.DestinationRegions {
		if region != s.SourceRegion {
			destRegion = append(destRegion, region)
		}
	}
	return &api.RemoteCopyImageArgs{
		Name:       config.ImageName,
		DestRegion: destRegion,
	}
}
