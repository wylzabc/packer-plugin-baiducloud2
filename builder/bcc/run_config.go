//go:generate packer-sdc struct-markdown

package bcc

import (
	"errors"
	"fmt"
	"os"

	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/hashicorp/packer-plugin-sdk/uuid"
)

// The data disk to amount on instance if you use this struct,
// for security, you will need to deal with
// the volume by yourself
type BaiduCloudDataDisk struct {
	CdsSizeInGB int    `mapstructure:"cds_size_in_gb"`
	StorageType string `mapstructure:"storage_type"`
	// The snapshot id for cds data disk. The field must be specified
	// or the cds data disk will not be created
	SnapShotId string `mapstructure:"snapshot_id"`
}

type BaiduCloudRunConfig struct {
	// Whether allocate public ip(eip) to your instance.
	// Default value is false. If you set this field `true`,
	// please set the `network_capacity_in_mbps` field to allocate
	// public network bandwith
	AssociatePublicIpAddress bool `mapstructure:"associate_public_ip_address" required:"false"`
	// Use default vpc, subnet, securitygroup, if the field is set `true`.
	// if the field is set `false` or not set, the build process will use
	// a vpc, subnet, securitygroup that you offer or create a temporary vpc,
	// subnet, securitygroup, which will be deleted after building.
	UseDefaultNetwork bool `mapstructure:"use_default_network" required:"false"`
	// Type of the instance. For values, see the section [区域机型以及可选配置] of
	// website https://cloud.baidu.com/doc/BCC/s/6jwvyo0q2
	InstanceSpec string `mapstructure:"instance_spec" required:"true"`
	// The name of instance which will be launch. Usually it will be deleted
	// after build, cancellation or error
	InstanceName string `mapstructure:"instance_name" required:"false"`
	// The description of instance
	Description string `mapstructure:"description"`
	// The base image id of Image you want to create
	// your customized image from
	SourceImageId string `mapstructure:"source_image_id" required:"true"`
	// ID of the security group to which a newly
	// created instance belongs. Mutual access is allowed between instances in one
	// security group. If not specified, the newly created instance will be added
	// to the default security group. If the default group doesn’t exist, or the
	// number of instances in it has reached the maximum limit, a new security
	// group will be created automatically.
	SecurityGroupId string `mapstructure:"security_group_id" required:"false"`
	// The security group name
	SecurityGroupName string `mapstructure:"security_group_name" required:"false"`
	// Internet charge type
	InternetChargeType string `mapstructure:"internet_charge_type" required:"false"`
	// Eip name
	EipName string `mapstructure:"eip_name" required:"false"`
	// Eip network bandwith
	NetworkCapacityInMbps int `mapstructure:"network_capacity_in_mbps" required:"false"`
	// RootDiskSizeInGb
	RootDiskSizeInGb int `mapstructure:"root_disk_size_in_gb" required:"false"`
	// RootDiskStorageType
	RootDiskStorageType string `mapstructure:"root_disk_storage_type" required:"false"`
	// The VPC id for your build process will perform in. If the field
	// `use_default_network` is used. This field will be invalid
	VpcId string `mapstructure:"vpc_id" require:"false"`
	// The VPC name. If the field `use_default_network` and `vpc_id`
	// both are unset, the build process will create a temporary
	// vpc network with the name of `vpc_name`
	VpcName string `mapstructure:"vpc_name" require:"false"`
	// Cidr block for VPC. If the temporary vpc network is created, the
	// `vpc_cidr_block` will be used to set cidr for vpc
	CidrBlock string `mapstructure:"vpc_cidr_block" required:"false"`
	// The id of subnet where the bcc instance will be launched on
	SubnetId string `mapstructure:"subnet_id" required:"false"`
	// The cidr block of subnet which will be created
	// if subnet id is not specified
	SubnetCidrBlock string `mapstructure:"subnet_cidr_block" required:"false"`
	// The subnet name which will be created if subnet id is not specified
	SubnetName string `mapstructure:"subnet_name" required:"false"`
	// Keypair id for ssh. baiducloud use keypair id as unique identification.
	// And maybe there more than one keypairs have the same name. So please
	// use `keypair_id` instead of `ssh_keypair_name`
	KeypairId string `mapstructure:"keypair_id" required:"false"`
	// Add one or more data disks to the instance before creating the image.
	// The data disks allow for the following argument:
	// -  `storage_type` - Type of the data disk.
	// -  `cds_size_in_gb` - Size of the data disk.
	// -  `snapshot_id` - Id of the snapshot for a data disk.
	DataDisks []BaiduCloudDataDisk `mapstructure:"data_disks"`
	// Key/value pair tags to apply to the instance that is *launched*
	// to create the image.
	RunTags map[string]string `mapstructure:"run_tags" required:"false"`
	// User data to apply when launching the instance.
	// It is often more convenient to use user_data_file, instead.
	// Packer will not automatically wait for a user script to finish before
	// shutting down the instance this must be handled in a provisioner.
	UserData string `mapstructure:"user_data" required:"false"`
	// Path to a file that will be used for the user
	// data when launching the instance.
	UserDataFile string `mapstructure:"user_data_file" required:"false"`

	// Communicator settings
	Comm communicator.Config `mapstructure:",squash"`
	// If this value is true, packer will connect to
	// the BCC instance created through private ip instead of allocating  an
	// EIP. The default value is false.
	// SSHPrivateIp bool `mapstructure:"ssh_private_ip" required:"false"`
}

func (c *BaiduCloudRunConfig) Prepare(ctx *interpolate.Context) []error {
	packerId := fmt.Sprintf("packer_%s", uuid.TimeOrderedUUID()[:8])

	if c.KeypairId == "" && c.Comm.SSHKeyPairName == "" && c.Comm.SSHTemporaryKeyPairName == "" &&
		c.Comm.SSHPrivateKeyFile == "" && c.Comm.SSHPassword == "" && c.Comm.WinRMPassword == "" {
		c.Comm.SSHTemporaryKeyPairName = packerId
	}

	if c.RunTags == nil {
		c.RunTags = make(map[string]string)
	}

	errs := c.Comm.Prepare(ctx)

	if c.SourceImageId == "" {
		errs = append(errs, errors.New("'source_image_id' must be specified"))
	}

	if c.InstanceSpec == "" {
		errs = append(errs, errors.New("'instance_spec' must be specified"))
	}

	if c.UserData != "" && c.UserDataFile != "" {
		errs = append(errs, errors.New("only one of 'user_data' or 'user_data_file' can be specified"))
	} else if c.UserDataFile != "" {
		if _, err := os.Stat(c.UserDataFile); err != nil {
			errs = append(errs, errors.New("the file path of 'user_data_file' doesn't exist"))
		}
	}

	if c.UseDefaultNetwork {
		// If using default network, there is no need to provide network info
		if c.VpcId != "" || c.VpcName != "" || c.CidrBlock != "" || c.SubnetId != "" ||
			c.SubnetCidrBlock != "" || c.SecurityGroupId != "" || c.SecurityGroupName != "" {
			errs = append(errs, errors.New("there is no need to provide vpc info, subnet info or "+
				"security group info, since the field 'use_default_network' has been set true"))
		}
	} else {
		// If use_default_network field is not provided or is set false.
		// you should provide a existing network id or the build
		// process will create a temporary network
		if c.VpcId != "" && (c.CidrBlock != "" || c.VpcName != "") {
			// If vpc_id is provided, there is no need to provide cidr_block or vpc_name
			errs = append(errs, errors.New("there is no need to provide 'cidr_block' or 'vpc_name',"+
				" since the 'vpc_id' has been set"))
		} else if c.VpcId == "" {
			// Provide a default vpc_name or cidr_block to create a temporary vpc,
			// if vpc_id is set null, and the vpc_name or cidr_block is not provided.
			if c.VpcName == "" {
				c.VpcName = packerId
			}
			if c.CidrBlock == "" {
				c.CidrBlock = "192.168.0.0/16"
			}

			// If vpc_id is not set, the build process will create a temporary subnet,
			// along with temporary vpc
			if c.SubnetId != "" {
				errs = append(errs, errors.New("can't set 'subnet_id' without set 'vpc_id'"))
			}
			if c.SubnetName == "" {
				c.SubnetName = packerId
			}
			if c.SubnetCidrBlock == "" {
				c.SubnetCidrBlock = "192.168.8.0/24"
			}

			// If vpc_id is not set, the build process will create a temporary
			// security group along with temporary vpc
			if c.SecurityGroupId != "" {
				errs = append(errs, errors.New("can't set 'security_group_id without set 'vpc_id'"))
			}
			if c.SecurityGroupName == "" {
				c.SecurityGroupName = packerId
			}
		}

		// If vpc_id is provided and subnet_id is not provided, the build process
		// will create a temporary subnet by the subnet_cidr_block
		if c.VpcId != "" && c.SubnetId == "" {
			if c.SubnetCidrBlock == "" {
				errs = append(errs, errors.New("'subnet_cidr_block' must be provide, if 'vpc_id' is "+
					"provided and 'subnet_id' is null"))
			}
			if c.SubnetName == "" {
				c.SubnetName = packerId
			}
		}

		// If vpc_id is provided and security_group_id is not provided, the build process
		// will create a temporary security group
		if c.VpcId != "" && c.SecurityGroupId == "" {
			if c.SecurityGroupName == "" {
				c.SecurityGroupName = packerId
			}
		}
	}

	if c.AssociatePublicIpAddress {
		if c.EipName == "" {
			c.EipName = packerId
		}
		if c.NetworkCapacityInMbps < 1 {
			c.NetworkCapacityInMbps = 1
		}
		if c.InternetChargeType == "" {
			c.InternetChargeType = "BANDWIDTH_POSTPAID_BY_HOUR"
		}
		if c.InternetChargeType != "BANDWIDTH_POSTPAID_BY_HOUR" && c.InternetChargeType != "TRAFFIC_POSTPAID_BY_HOUR" {
			errs = append(errs, errors.New("there is a error at internet_charge_type field, please check it"))
		}
	} else {
		if c.EipName != "" || c.NetworkCapacityInMbps != 0 || c.InternetChargeType != "" {
			errs = append(errs, errors.New("no need to set fields 'eip_name', 'network_capacity_in_mbps', or 'internet_charge_type', "+
				"since the 'associate_public_ip_address' field is false"))
		}
	}

	if len(c.DataDisks) > 0 {
		for _, disk := range c.DataDisks {
			if disk.SnapShotId == "" {
				errs = append(errs, errors.New("the 'snapshot_id' in 'data_disks' should be provided"))
			}
		}
	}

	return errs
}
