package bcc

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"

	"github.com/baidubce/bce-sdk-go/model"
	"github.com/baidubce/bce-sdk-go/services/bcc"
	"github.com/baidubce/bce-sdk-go/services/bcc/api"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/uuid"
)

type stepCreateInstance struct {
	UseDefaultNetwork        bool
	AssociatePublicIpAddress bool
	SourceImageId            string
	InstanceName             string
	InstanceSpec             string
	ZoneName                 string
	InternetChargeType       string
	SecurityGroupId          string
	instanceId               string
	EipName                  string
	NetworkCapacityInMbps    int
	UserData                 string
	UserDataFile             string
	Tags                     map[string]string
}

func (s *stepCreateInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*bcc.Client)
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Creating instance...")
	createInstanceBySpecArgs, err := s.getCreateInstanceBySpecArgs(state)
	if err != nil {
		return halt(state, err, "Failed to get `CreateInstanceBySpecArgs`")
	}
	createResult, err := client.CreateInstanceBySpec(createInstanceBySpecArgs)
	if err != nil {
		return halt(state, err, "Failed to create instance")
	}
	// check the return result of creating instance
	if len(createResult.InstanceIds) == 0 {
		return halt(state, fmt.Errorf("No instance id return"), "Failed to create instance")
	}

	// wait for the finish of instance
	instanceId := createResult.InstanceIds[0]
	ui.Say(fmt.Sprintf("Waiting for instance %s ready...", instanceId))
	err = WaitForInstance(ctx, client, instanceId, api.InstanceStatusRunning, 1800)
	if err != nil {
		return halt(state, err, fmt.Sprintf("Failed to wait for instance(%s) ready", instanceId))
	}

	// get instance detail
	instanceDetail, err := client.GetInstanceDetail(instanceId)
	if err != nil {
		return halt(state, err, fmt.Sprintf("Failed to get instance detail: %s", err))
	}
	instance := instanceDetail.Instance
	ui.Message(fmt.Sprintf("Success to create instance, the instance id is %s, public ip is %s, private ip is %s", instance.InstanceId, instance.PublicIP, instance.InternalIP))
	state.Put("instance_id", instanceId)
	state.Put("instance", &instance)
	s.instanceId = instanceId

	return multistep.ActionContinue
}

func (s *stepCreateInstance) Cleanup(state multistep.StateBag) {
	if len(s.instanceId) == 0 {
		return
	}

	ctx := context.TODO()

	// clean up message
	cleanUpMessage(state, "instance and relate resource")

	client := state.Get("client").(*bcc.Client)
	ui := state.Get("ui").(packersdk.Ui)
	// instance := state.Get("instance").(*api.InstanceModel)

	err := Retry(ctx, func(ctx context.Context) error {
		return client.DeleteInstanceWithRelateResource(s.instanceId, &api.DeleteInstanceWithRelateResourceArgs{
			RelatedReleaseFlag: true,
		})
	})
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to clean up instance %s: %s", s.instanceId, err))
	}

	// if instance.PublicIP != "" {
	// 	// todo clean eip
	// 	cleanUpMessage(state, "eip")
	// 	eipClient := state.Get("eip_client").(*eip.Client)
	// 	// clean eip
	// 	err := Retry(ctx, func(ctx context.Context) error {
	// 		return eipClient.DeleteEip(instance.PublicIP, "")
	// 	})
	// 	if err != nil {
	// 		ui.Error(fmt.Sprintf("Failed to clean up eip: %s", instance.PublicIP))
	// 	}
	// }
}

func (s *stepCreateInstance) getCreateInstanceBySpecArgs(state multistep.StateBag) (*api.CreateInstanceBySpecArgs, error) {
	config := state.Get("config").(*Config)

	keypairId := config.KeypairId
	if len(keypairId) == 0 {
		// Use temporary keypair
		if _, ok := state.GetOk("temporary_key_pair_id"); ok {
			keypairId = state.Get("temporary_key_pair_id").(string)
		}
	}

	password := config.Comm.SSHPassword
	if password == "" && config.Comm.WinRMPassword != "" {
		password = config.Comm.WinRMPassword
	}

	userData, err := s.getUserData(state)
	if err != nil {
		return nil, err
	}

	// var dataDisks []api.CreateCdsModel
	// for _, disk := range config.DataDisks {
	// 	var datadisk api.CreateCdsModel
	// 	datadisk.CdsSizeInGB = disk.CdsSizeInGB
	// 	datadisk.StorageType = api.StorageType(disk.StorageType)
	// 	datadisk.SnapShotId = disk.SnapShotId
	// 	dataDisks = append(dataDisks, datadisk)
	// }

	var tags []model.TagModel
	for k, v := range s.Tags {
		tags = append(tags, model.TagModel{TagKey: k, TagValue: v})
	}

	args := &api.CreateInstanceBySpecArgs{
		ImageId: s.SourceImageId,
		Billing: api.Billing{
			PaymentTiming: api.PaymentTimingPostPaid,
		},
		Spec:                s.InstanceSpec,
		RootDiskSizeInGb:    config.RootDiskSizeInGb,
		RootDiskStorageType: api.StorageType(config.RootDiskStorageType),
		ZoneName:            s.ZoneName,
		PurchaseCount:       1,
		Name:                s.InstanceName,
		ClientToken:         uuid.TimeOrderedUUID(),
		// CreateCdsList:       dataDisks,
		UserData: userData,
	}

	if password != "" {
		args.AdminPass = password
	}

	if keypairId != "" {
		args.KeypairId = keypairId
	}

	if !s.UseDefaultNetwork {
		subnetId := state.Get("subnet_id").(string)
		securityGroupId := state.Get("security_group_id").(string)

		args.SubnetId = subnetId
		args.SecurityGroupId = securityGroupId
	}

	if s.AssociatePublicIpAddress {
		args.EipName = s.EipName
		if s.NetworkCapacityInMbps < 1 {
			s.NetworkCapacityInMbps = 1
		}
		args.NetWorkCapacityInMbps = s.NetworkCapacityInMbps
		args.InternetChargeType = s.InternetChargeType
	}

	if len(tags) > 0 {
		args.Tags = tags
		args.RelationTag = true
	}

	return args, nil
}

func (s *stepCreateInstance) getUserData(state multistep.StateBag) (string, error) {
	userData := s.UserData

	if len(s.UserDataFile) != 0 {
		data, err := ioutil.ReadFile(s.UserDataFile)
		if err != nil {
			return "", err
		}

		userData = string(data)
	}

	if len(userData) != 0 {
		userData = base64.StdEncoding.EncodeToString([]byte(userData))
	}

	return userData, nil
}
