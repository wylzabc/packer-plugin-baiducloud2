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

type stepConfigSecurityGroup struct {
	UseDefaultNetwork bool
	SecurityGroupId   string
	SecurityGroupName string
	Description       string
	VpcId             string
	isCreate          bool
}

func (s *stepConfigSecurityGroup) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	if s.UseDefaultNetwork {
		return multistep.ActionContinue
	}

	client := state.Get("client").(*bcc.Client)
	ui := state.Get("ui").(packersdk.Ui)
	s.VpcId = state.Get("vpc_id").(string)

	if len(s.SecurityGroupId) != 0 {
		// todo check security group
		ui.Say(fmt.Sprintf("Trying to check security group(%s)...", s.SecurityGroupId))
		listResult, err := client.ListSecurityGroup(s.getListSecurityGroupArgs())
		if err != nil {
			return halt(state, err, "Failed to list security group")
		}

		for _, securityGroupItem := range listResult.SecurityGroups {
			if securityGroupItem.Id == s.SecurityGroupId {
				state.Put("security_group_id", s.SecurityGroupId)
				s.isCreate = false
				return multistep.ActionContinue
			}
		}

		s.isCreate = false
		return halt(state, fmt.Errorf("The specified security group(%s) doesn't exist", s.SecurityGroupId), "")
	}

	// create security group
	ui.Say("Starting to create security group...")

	var createSecurityGroupResult *api.CreateSecurityGroupResult
	err := Retry(ctx, func(ctx context.Context) error {
		var e error
		createSecurityGroupResult, e = client.CreateSecurityGroup(s.getCreateSecurityGroupArgs())
		return e
	})
	if err != nil {
		return halt(state, err, "Failed to create security group")
	}

	securityGroupId := createSecurityGroupResult.SecurityGroupId

	ui.Message(fmt.Sprintf("Success to create security group: %s", securityGroupId))
	state.Put("security_group_id", securityGroupId)
	s.isCreate = true
	s.SecurityGroupId = securityGroupId

	return multistep.ActionContinue
}

func (s *stepConfigSecurityGroup) Cleanup(state multistep.StateBag) {
	if !s.isCreate || s.UseDefaultNetwork {
		return
	}

	cleanUpMessage(state, "security group")

	client := state.Get("client").(*bcc.Client)
	ui := state.Get("ui").(packersdk.Ui)
	ctx := context.TODO()

	err := Retry(ctx, func(ctx context.Context) error {
		return client.DeleteSecurityGroup(s.SecurityGroupId)
	})
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to delete security group(%s), you can delete it manually: %s", s.SecurityGroupId, err))
	}
}

func (s *stepConfigSecurityGroup) getListSecurityGroupArgs() *api.ListSecurityGroupArgs {
	return &api.ListSecurityGroupArgs{
		VpcId: s.VpcId,
	}
}

func (s *stepConfigSecurityGroup) getCreateSecurityGroupArgs() *api.CreateSecurityGroupArgs {
	return &api.CreateSecurityGroupArgs{
		ClientToken: uuid.TimeOrderedUUID(),
		VpcId:       s.VpcId,
		Name:        s.SecurityGroupName,
		Desc:        s.Description,
		Rules: []api.SecurityGroupRuleModel{
			{
				Remark:    "packer test ingress",
				Direction: "ingress",
			},
			{
				Remark:    "packer test egress",
				Direction: "egress",
			},
		},
	}
}
