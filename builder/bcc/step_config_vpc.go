package bcc

import (
	"context"
	"fmt"

	"github.com/baidubce/bce-sdk-go/services/vpc"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/uuid"
)

type stepConfigVPC struct {
	UseDefaultNetwork bool
	VpcId             string
	CidrBlock         string
	VpcName           string
	Description       string
	isCreate          bool
}

func (s *stepConfigVPC) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	if s.UseDefaultNetwork {
		return multistep.ActionContinue
	}

	client := state.Get("vpc_client").(*vpc.Client)
	ui := state.Get("ui").(packersdk.Ui)

	if len(s.VpcId) != 0 {
		// check whether the specified vpi id exists
		ui.Say(fmt.Sprintf("Trying to check existing VPC(%s)...", s.VpcId))
		vpcDetail, err := client.GetVPCDetail(s.VpcId)
		if err != nil {
			return halt(state, err, fmt.Sprintf("Failed to get specified vpc(%s)", s.VpcId))
		}

		state.Put("vpc_id", vpcDetail.VPC.VPCId)

		return multistep.ActionContinue
	}

	// create vpc
	ui.Say("Starting to create vpc...")

	var createResult *vpc.CreateVPCResult
	err := Retry(ctx, func(ctx context.Context) error {
		var e error
		createResult, e = client.CreateVPC(s.getCreateVpcArgs())
		return e
	})
	if err != nil {
		return halt(state, err, "Failed to create vpc")
	}

	vpcId := createResult.VPCID

	ui.Message(fmt.Sprintf("Success to create vpc: %s", vpcId))
	state.Put("vpc_id", vpcId)
	s.isCreate = true
	s.VpcId = vpcId

	return multistep.ActionContinue
}

func (s *stepConfigVPC) Cleanup(state multistep.StateBag) {
	if !s.isCreate || s.UseDefaultNetwork {
		return
	}

	cleanUpMessage(state, "VPC")

	client := state.Get("vpc_client").(*vpc.Client)
	ui := state.Get("ui").(packersdk.Ui)
	ctx := context.TODO()

	err := Retry(ctx, func(ctx context.Context) error {
		return client.DeleteVPC(s.VpcId, uuid.TimeOrderedUUID())
	})
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to delete vpc(%s), please clean it manually: %s", s.VpcId, err))
	}
}

func (s *stepConfigVPC) getCreateVpcArgs() *vpc.CreateVPCArgs {
	return &vpc.CreateVPCArgs{
		ClientToken: uuid.TimeOrderedUUID(),
		Name:        s.VpcName,
		Cidr:        s.CidrBlock,
		Description: s.Description,
	}
}
