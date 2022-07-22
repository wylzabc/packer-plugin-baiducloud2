package bcc

import (
	"context"
	"fmt"

	"github.com/baidubce/bce-sdk-go/services/bcc"
	"github.com/baidubce/bce-sdk-go/services/bcc/api"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// stepDetachKeyPair is used to detach temporary keypair from instance after provison
type stepDetachKeyPair struct {
}

func (s *stepDetachKeyPair) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*bcc.Client)

	if _, ok := state.GetOk("temporary_key_pair_id"); !ok {
		return multistep.ActionContinue
	}

	keyPairId := state.Get("temporary_key_pair_id").(string)
	instanceId := state.Get("instance_id").(string)
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say(fmt.Sprintf("Trying to detach temporary keypair(%s)...", keyPairId))

	err := Retry(ctx, func(ctx context.Context) error {
		return client.DetachKeypair(&api.DetachKeypairArgs{
			KeypairId:   keyPairId,
			InstanceIds: []string{instanceId},
		})
	})
	if err != nil {
		return halt(state, err, "Failed to detach keypair")
	}

	ui.Message(fmt.Sprintf("Success to detach keypair(%s) from instance(%s)", keyPairId, instanceId))

	// Wait instance running
	if err := WaitForInstance(ctx, client, instanceId, api.InstanceStatusRunning, 600); err != nil {
		return halt(state, err, "Failed to wait instance running after detach keypair")
	}

	return multistep.ActionContinue
}

func (s *stepDetachKeyPair) Cleanup(state multistep.StateBag) {}
