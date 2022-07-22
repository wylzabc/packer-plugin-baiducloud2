package bcc

import (
	"context"
	"fmt"

	"github.com/baidubce/bce-sdk-go/services/bcc/api"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type stepShareImage struct {
	shareAccouts    []string
	shareAccountIds []string
}

func (s *stepShareImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	baiduCloudImages := state.Get("baiducloud_images").(map[string]string)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)

	shareAccountArgsList := s.getShareAccountArgsList()
	if len(shareAccountArgsList) == 0 {
		return multistep.ActionContinue
	}
	ui.Say("Starting to share images to specified users...")
	for region, imageId := range baiduCloudImages {
		client, err := config.ClientWithRegion(region)
		if err != nil {
			return halt(state, err, fmt.Sprintf("Failed to get bcc client of region(%s)", region))
		}
		for _, shareAccoutArgs := range shareAccountArgsList {
			err := Retry(ctx, func(ctx context.Context) error {
				return client.ShareImage(imageId, shareAccoutArgs)
			})
			if err != nil {
				return halt(state, err, fmt.Sprintf("Failed to share image(%s) of region(%s) to account(%+v)", imageId, region, shareAccoutArgs))
			} else {
				ui.Message(fmt.Sprintf("Success to share image(%s) of region(%s) to account(%+v)", imageId, region, shareAccoutArgs))
			}
		}
	}

	return multistep.ActionContinue
}

func (s *stepShareImage) Cleanup(state multistep.StateBag) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)

	if !cancelled && !halted {
		return
	}

	shareAccountArgsList := s.getShareAccountArgsList()
	if len(shareAccountArgsList) == 0 {
		return
	}

	ui := state.Get("ui").(packersdk.Ui)
	config := state.Get("config").(*Config)
	baiduCloudImages := state.Get("baiducloud_images").(map[string]string)
	ctx := context.TODO()

	ui.Error("Cancel image share because cancellations or error...")

	for region, imageId := range baiduCloudImages {
		client, err := config.ClientWithRegion(region)
		if err != nil {
			ui.Error(fmt.Sprintf("Failed to get bcc client of region(%s): %s", region, err))
		}
		for _, shareAccoutArgs := range shareAccountArgsList {
			err := Retry(ctx, func(ctx context.Context) error {
				return client.UnShareImage(imageId, shareAccoutArgs)
			})
			if err != nil {
				ui.Error(fmt.Sprintf("Failed to cancel share image(%s) of region(%s) to account(%+v)", imageId, region, shareAccoutArgs))
			} else {
				ui.Message(fmt.Sprintf("Success to cancel share image(%s) of region(%s) to account(%+v)", imageId, region, shareAccoutArgs))
			}
		}
	}
}

func (s *stepShareImage) getShareAccountArgsList() []*api.SharedUser {
	var shareAccountArgsList []*api.SharedUser
	for _, account := range s.shareAccouts {
		shareAccountArgsList = append(shareAccountArgsList, &api.SharedUser{
			Account: account,
		})
	}
	for _, accountId := range s.shareAccountIds {
		shareAccountArgsList = append(shareAccountArgsList, &api.SharedUser{
			AccountId: accountId,
		})
	}
	return shareAccountArgsList
}
