package bcc

import (
	"context"
	"fmt"

	"github.com/baidubce/bce-sdk-go/services/bcc"
	"github.com/baidubce/bce-sdk-go/services/bcc/api"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type stepStopInstance struct {
	ForceStop        bool
	DisableStop      bool
	StopWithNoCharge bool
}

func (s *stepStopInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*bcc.Client)
	instanceId := state.Get("instance_id").(string)
	ui := state.Get("ui").(packersdk.Ui)

	pause(state)

	err := client.StopInstanceWithNoCharge(instanceId, s.ForceStop, s.StopWithNoCharge)
	if err != nil {
		return halt(state, err, "Failed to stop instance")
	}
	ui.Say(fmt.Sprintf("Waiting instance(%s) stop", instanceId))
	err = WaitForInstance(ctx, client, instanceId, api.InstanceStatusStopped, 600)
	if err != nil {
		return halt(state, err, "Failed to stop bcc instance")
	}
	return multistep.ActionContinue
}

func (s *stepStopInstance) Cleanup(multistep.StateBag) {}
