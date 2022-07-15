package bcc

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/baidubce/bce-sdk-go/services/bcc"
	"github.com/baidubce/bce-sdk-go/services/bcc/api"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/retry"
)

func halt(state multistep.StateBag, err error, prefix string) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)

	if prefix != "" {
		err = fmt.Errorf("%s: %s", prefix, err)
	}

	state.Put("error", err)
	ui.Error(err.Error())
	return multistep.ActionHalt
}

func cleanUpMessage(state multistep.StateBag, module string) {
	_, isCancelled := state.GetOk(multistep.StateCancelled)
	_, isHalted := state.GetOk(multistep.StateHalted)

	ui := state.Get("ui").(packersdk.Ui)

	if isCancelled {
		ui.Say(fmt.Sprintf("Deleting %s because of cancellation...", module))
	} else if isHalted {
		ui.Say(fmt.Sprintf("Deleting %s because of error...", module))
	} else {
		ui.Say(fmt.Sprintf("Cleaning up '%s'...", module))
	}
}

func Retry(ctx context.Context, fn func(context.Context) error) error {
	return retry.Config{
		Tries: 60,
		ShouldRetry: func(err error) bool {
			return false
		},
		RetryDelay: (&retry.Backoff{
			InitialBackoff: 1 * time.Second,
			MaxBackoff:     5 * time.Second,
			Multiplier:     2,
		}).Linear,
	}.Run(ctx, fn)
}

// DefaultWaitForInterval is the default wait interval of query instance detail
const DefaultWaitForInterval = 5

// WaitForInstance - waiting for the specific isntance reaching target status
func WaitForInstance(ctx context.Context, client *bcc.Client, instanceId string, targetStatus api.InstanceStatus, timeout int) error {
	for {
		var detailResult *api.GetInstanceDetailResult
		err := Retry(ctx, func(ctx context.Context) error {
			var e error
			detailResult, e = client.GetInstanceDetail(instanceId)
			return e
		})
		if err != nil {
			return err
		}
		if detailResult.Instance.Status == targetStatus {
			break
		}
		time.Sleep(DefaultWaitForInterval * time.Second)
		timeout = timeout - DefaultWaitForInterval
		if timeout <= 0 {
			return fmt.Errorf("wait instance(%s) status(%s) timeout", instanceId, targetStatus)
		}
	}
	return nil
}

func WaitForImage(ctx context.Context, client *bcc.Client, imageId string, targetStatus api.ImageStatus, timeout int) error {
	for {
		var detailResult *api.GetImageDetailResult
		err := Retry(ctx, func(ctx context.Context) error {
			var e error
			detailResult, e = client.GetImageDetail(imageId)
			return e
		})
		if err != nil {
			return err
		}
		if detailResult.Image.Status == targetStatus {
			break
		}
		time.Sleep(DefaultWaitForInterval * time.Second)
		timeout = timeout - DefaultWaitForInterval
		if timeout <= 0 {
			return fmt.Errorf("wait image(%s) status(%s) timeout", imageId, targetStatus)
		}
	}
	return nil
}

func pause(state multistep.StateBag) {
	ui := state.Get("ui").(packersdk.Ui)
	anykey := make([]byte, 1)
	ui.Say("Please enter any key to continue...")
	os.Stdin.Read(anykey)
}
