package bcc

import (
	"context"
	"fmt"

	"github.com/baidubce/bce-sdk-go/services/vpc"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/uuid"
)

type stepConfigSubnet struct {
	UseDefaultNetwork bool
	SubnetId          string
	SubnetCidrBlock   string
	SubnetName        string
	ZoneName          string
	Description       string
	isCreate          bool
}

func (s *stepConfigSubnet) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	if s.UseDefaultNetwork {
		return multistep.ActionContinue
	}

	client := state.Get("vpc_client").(*vpc.Client)
	ui := state.Get("ui").(packersdk.Ui)

	if len(s.SubnetId) != 0 {
		ui.Say(fmt.Sprintf("Trying to use existing subnet(%s)...", s.SubnetId))
		subnetDetail, err := client.GetSubnetDetail(s.SubnetId)
		if err != nil {
			return halt(state, err, fmt.Sprintf("Failed to get subnet(%s), it may not exist: %s", s.SubnetId, err))
		}
		state.Put("subnet_id", subnetDetail.Subnet.SubnetId)
		return multistep.ActionContinue
	}

	// create new subnet
	ui.Say("Starting to create new subnet...")

	createResult, err := client.CreateSubnet(s.getCreateSubnetArgs(state))
	if err != nil {
		return halt(state, err, "Failed to create subnet")
	}

	s.isCreate = true
	s.SubnetId = createResult.SubnetId
	state.Put("subnet_id", s.SubnetId)
	ui.Message(fmt.Sprintf("Success to create subnet: %s", s.SubnetId))

	return multistep.ActionContinue
}

func (s *stepConfigSubnet) Cleanup(state multistep.StateBag) {
	if !s.isCreate || s.UseDefaultNetwork {
		return
	}

	client := state.Get("vpc_client").(*vpc.Client)
	ui := state.Get("ui").(packersdk.Ui)

	cleanUpMessage(state, "subnet")

	err := client.DeleteSubnet(s.SubnetId, uuid.TimeOrderedUUID())
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to delete subnet(%s), please clean it manully: %s", s.SubnetId, err))
	}
}

func (s *stepConfigSubnet) getCreateSubnetArgs(state multistep.StateBag) *vpc.CreateSubnetArgs {
	vpcId := state.Get("vpc_id").(string)

	return &vpc.CreateSubnetArgs{
		ClientToken: uuid.TimeOrderedUUID(),
		SubnetType:  vpc.SUBNET_TYPE_BCC,
		Name:        s.SubnetName,
		Cidr:        s.SubnetCidrBlock,
		Description: s.Description,
		VpcId:       vpcId,
		ZoneName:    s.ZoneName,
	}
}
