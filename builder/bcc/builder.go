//go:generate packer-sdc mapstructure-to-hcl2 -type Config

package bcc

import (
	"context"
	"fmt"

	"github.com/baidubce/bce-sdk-go/services/bcc/api"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

// The unique ID for this builder component
const BuilderId = "baidu.baiducloud"

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	BaiduCloudAccessConfig `mapstructure:",squash"`
	BaiduCloudImageConfig  `mapstructure:",squash"`
	BaiduCloudRunConfig    `mapstructure:",squash"`

	ctx interpolate.Context
}

type Builder struct {
	config Config
	runner multistep.Runner
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) ([]string, []string, error) {
	err := config.Decode(&b.config, &config.DecodeOpts{
		PluginType:         BuilderId,
		Interpolate:        true,
		InterpolateContext: &b.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"run_command",
			},
		},
	}, raws...)
	b.config.ctx.EnableEnv = true
	if err != nil {
		return nil, nil, err
	}

	// Accumulate any errors
	var errs *packersdk.MultiError
	errs = packersdk.MultiErrorAppend(errs, b.config.BaiduCloudAccessConfig.Prepare(&b.config.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, b.config.BaiduCloudImageConfig.Prepare(&b.config.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, b.config.BaiduCloudRunConfig.Prepare(&b.config.ctx)...)

	if errs != nil && len(errs.Errors) > 0 {
		return nil, nil, errs
	}

	packersdk.LogSecretFilter.Set(b.config.BaiduCloudAccessKey, b.config.BaiduCloudSecretKey)
	return nil, nil, nil
}

func (b *Builder) Run(ctx context.Context, ui packersdk.Ui, hook packersdk.Hook) (packersdk.Artifact, error) {
	ui.Say("Start to run baiducloud builder")
	client, err := b.config.Client()
	if err != nil {
		return nil, err
	}
	vpcClient, err := b.config.VpcClient()
	if err != nil {
		return nil, err
	}
	eipClient, err := b.config.EipClient()
	if err != nil {
		return nil, err
	}

	ui.Say("connect to client:" + client.Config.Endpoint)
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("client", client)
	state.Put("vpc_client", vpcClient)
	state.Put("eip_client", eipClient)
	state.Put("hook", hook)
	state.Put("ui", ui)

	var steps []multistep.Step

	// Build the steps
	steps = []multistep.Step{
		&stepPreValidate{
			SourceImageId:       b.config.SourceImageId,
			CustomImageName:     b.config.ImageName,
			SkipImageValidation: b.config.SkipImageValidation,
		},
		&stepConfigKeyPair{
			Debug:        b.config.PackerDebug,
			Comm:         &b.config.Comm,
			KeyPairId:    b.config.KeypairId,
			DebugKeyPath: fmt.Sprintf("bcc_%s.pem", b.config.PackerBuildName),
			Description:  "keypair for packer",
		},
		&stepConfigVPC{
			UseDefaultNetwork: b.config.UseDefaultNetwork,
			VpcId:             b.config.VpcId,
			VpcName:           b.config.VpcName,
			CidrBlock:         b.config.CidrBlock,
			Description:       "vpc for packer",
		},
		&stepConfigSubnet{
			UseDefaultNetwork: b.config.UseDefaultNetwork,
			SubnetId:          b.config.SubnetId,
			SubnetName:        b.config.SubnetName,
			SubnetCidrBlock:   b.config.SubnetCidrBlock,
			ZoneName:          b.config.Zone,
			Description:       "subnet for packer",
		},
		&stepConfigSecurityGroup{
			UseDefaultNetwork: b.config.UseDefaultNetwork,
			SecurityGroupId:   b.config.SecurityGroupId,
			SecurityGroupName: b.config.SecurityGroupName,
			Description:       "security group for packer",
		},
		&stepCreateInstance{
			UseDefaultNetwork:        b.config.UseDefaultNetwork,
			SourceImageId:            b.config.SourceImageId,
			InstanceName:             b.config.InstanceName,
			InstanceSpec:             b.config.InstanceSpec,
			ZoneName:                 b.config.Zone,
			AssociatePublicIpAddress: b.config.AssociatePublicIpAddress,
			EipName:                  b.config.EipName,
			NetworkCapacityInMbps:    b.config.NetworkCapacityInMbps,
			InternetChargeType:       b.config.InternetChargeType,
			UserData:                 b.config.UserData,
			UserDataFile:             b.config.UserDataFile,
			Tags:                     b.config.RunTags,
		},
		&communicator.StepConnect{
			Config:    &b.config.BaiduCloudRunConfig.Comm,
			SSHConfig: b.config.BaiduCloudRunConfig.Comm.SSHConfigFunc(),
			Host:      getSSHHost(b.config.AssociatePublicIpAddress),
		},
		&commonsteps.StepProvision{},
		&commonsteps.StepCleanupTempKeys{
			Comm: &b.config.Comm,
		},
		&stepDetachKeyPair{},
		// &stepStopInstance{},
		&stepCreateImage{},
		&stepRemoteCopyImage{
			DestinationRegions: b.config.DestinationRegions,
			SourceRegion:       b.config.BaiduCloudRegion,
		},
		&stepShareImage{
			shareAccouts:    b.config.ImageShareAccounts,
			shareAccountIds: b.config.ImageShareAccountIds,
		},
	}

	b.runner = commonsteps.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)

	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	// build the artifact and return it
	artifact := &Artifact{
		BaiduCloudImages: state.Get("baiducloud_images").(map[string]string),
		BuilderIdValue:   BuilderId,
		Client:           client,
	}
	return artifact, nil
}

func getSSHHost(publicIp bool) func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {
		instance := state.Get("instance").(*api.InstanceModel)
		if publicIp {
			return instance.PublicIP, nil
		} else {
			return instance.InternalIP, nil
		}
	}
}
