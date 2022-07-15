package bcc

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/baidubce/bce-sdk-go/services/bcc"
	"github.com/baidubce/bce-sdk-go/services/bcc/api"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/uuid"
)

type stepConfigKeyPair struct {
	Debug        bool
	Comm         *communicator.Config
	DebugKeyPath string
	KeyPairId    string
	Description  string
	isCreate     bool
}

func (s *stepConfigKeyPair) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)

	if s.Comm.SSHPrivateKeyFile != "" {
		ui.Say("Try to use existing SSH private key")
		privateKeyBytes, err := s.Comm.ReadSSHPrivateKeyFile()
		if err != nil {
			return halt(state, err, "Failed to load configured private key")
		}
		s.Comm.SSHPrivateKey = privateKeyBytes
		ui.Say(fmt.Sprintf("Loaded %d bytes private key data", len(s.Comm.SSHPrivateKey)))
		return multistep.ActionContinue
	}

	if s.Comm.SSHAgentAuth && len(s.KeyPairId) == 0 {
		ui.Say("Using SSH Agent with key pair in source image")
		return multistep.ActionContinue
	}

	if s.Comm.SSHAgentAuth && len(s.KeyPairId) != 0 {
		ui.Say(fmt.Sprintf("Using SSH Agent with existing key pair(%s)", s.Comm.SSHKeyPairName))
		return multistep.ActionContinue
	}

	if s.Comm.SSHTemporaryKeyPairName == "" {
		ui.Say("Not using temporary keypair")
		s.Comm.SSHKeyPairName = ""
		return multistep.ActionContinue
	}

	client := state.Get("client").(*bcc.Client)
	// create temporary keypair
	ui.Say(fmt.Sprintf("Creating temporary keypair: %s", s.Comm.SSHTemporaryKeyPairName))

	createResult, err := client.CreateKeypair(s.getCreateKeypairArgs(state))
	if err != nil {
		return halt(state, err, "Failed to create temporary keypair")
	}

	// set the keypair id for delete it later
	s.KeyPairId = createResult.Keypair.KeypairId
	s.isCreate = true

	// set some state data for use in future steps
	s.Comm.SSHPrivateKey = []byte(createResult.Keypair.PrivateKey)
	state.Put("temporary_key_pair_id", s.KeyPairId)

	ui.Message(fmt.Sprintf("Success to create temporary keypair, the id is: %s", s.KeyPairId))

	// Set debug key path to use debug mode
	if s.Debug {
		ui.Message(fmt.Sprintf("Saving private key for debug at path: %s", s.DebugKeyPath))
		file, err := os.Create(s.DebugKeyPath)
		if err != nil {
			state.Put("error", fmt.Errorf("Failed to create debug key path: %s", err))
			return multistep.ActionHalt
		}
		defer file.Close()

		// Write private key to debug file
		if _, err := file.Write([]byte(s.Comm.SSHPrivateKey)); err != nil {
			state.Put("error", fmt.Errorf("Failed to save debug key: %s", err))
			return multistep.ActionHalt
		}

		// Chmod debug key path for ssh
		if runtime.GOOS != "windows" {
			if err := file.Chmod(0600); err != nil {
				state.Put("error", fmt.Errorf("Failed to set permissions of debug key file: %s", err))
				return multistep.ActionHalt
			}
		}
	}
	return multistep.ActionContinue
}

func (s *stepConfigKeyPair) Cleanup(state multistep.StateBag) {
	if !s.isCreate {
		return
	}

	client := state.Get("client").(*bcc.Client)
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Cleaning up 'keypair'...")

	err := client.DeleteKeypair(&api.DeleteKeypairArgs{KeypairId: s.KeyPairId})
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to cleanup keypair(%s), please delete it manually: %s", s.KeyPairId, err))
	}

	if s.Debug {
		if err := os.Remove(s.DebugKeyPath); err != nil {
			ui.Error(fmt.Sprintf("Failed to remove debug key file(%s), please delete the file manually: %s", s.DebugKeyPath, err))
		}
	}
}

func (s *stepConfigKeyPair) getCreateKeypairArgs(state multistep.StateBag) *api.CreateKeypairArgs {

	return &api.CreateKeypairArgs{
		ClientToken: uuid.TimeOrderedUUID(),
		Name:        s.Comm.SSHTemporaryKeyPairName,
		Description: s.Description,
	}
}
