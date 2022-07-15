package bcc

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/communicator"
)

func getTestRunConfig() *BaiduCloudRunConfig {
	return &BaiduCloudRunConfig{
		SourceImageId: "m-sDghxZu1",
		InstanceSpec:  "bcc.ic4.c2m2",
		Comm: communicator.Config{
			SSH: communicator.SSH{
				SSHUsername: "baiducloud",
			},
		},
	}
}

func TestRunConfigPrepare(t *testing.T) {
	c := getTestRunConfig()

	if errs := c.Prepare(nil); len(errs) > 0 {
		t.Fatalf("Shouldn't raise error: %s", errs)
	}

	c.InstanceSpec = ""
	if errs := c.Prepare(nil); len(errs) != 1 {
		t.Fatalf("Should raise an error: %s", errs)
	}

	c.InstanceSpec = "bcc.ic4.c2m2"
	c.SourceImageId = ""
	if errs := c.Prepare(nil); len(errs) != 1 {
		t.Fatalf("Should raise an error: %s", errs)
	}
}

func TestRunConfigPrepare_UserData(t *testing.T) {
	c := getTestRunConfig()

	c.UserData = "user data"
	if errs := c.Prepare(nil); len(errs) != 0 {
		t.Fatalf("Shouldn't raise error: %s", errs)
	}

	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("Shouldn't raise error: %s", err)
	}
	defer os.Remove(tf.Name())
	defer tf.Close()

	c.UserData = "packer user data"
	c.UserDataFile = tf.Name()
	if errs := c.Prepare(nil); len(errs) != 1 {
		t.Fatalf("Should raise an error: %s", errs)
	}
}

func TestRunConfigPrepare_UserDataFile(t *testing.T) {
	c := getTestRunConfig()

	c.UserDataFile = "don't exist"
	if errs := c.Prepare(nil); len(errs) != 1 {
		t.Fatalf("Should raise an error: %s", errs)
	}

	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("Shouldn't raise error: %s", err)
	}
	defer os.Remove(tf.Name())
	defer tf.Close()

	c.UserDataFile = tf.Name()
	if errs := c.Prepare(nil); len(errs) != 0 {
		t.Fatalf("Shouldn't raise error: %s", errs)
	}
}

func TestRunConfigPrepare_TemporaryKeyPairName(t *testing.T) {
	c := getTestRunConfig()

	c.Comm.SSHTemporaryKeyPairName = ""
	if errs := c.Prepare(nil); len(errs) != 0 {
		t.Fatalf("Shouldn't raise error: %s", errs)
	}

	if c.Comm.SSHTemporaryKeyPairName == "" {
		t.Fatal("Temporary keypair name shouldn't be empty")
	}

	c.Comm.SSHTemporaryKeyPairName = "packer-keypair"
	if errs := c.Prepare(nil); len(errs) != 0 {
		t.Fatalf("Shouldn't raise error: %s", errs)
	}

	if c.Comm.SSHTemporaryKeyPairName != "packer-keypair" {
		t.Fatalf("Keypair name doesn't match")
	}
}

func TestRunConfigPrepare_UseDefaultNetwork(t *testing.T) {
	c := getTestRunConfig()

	c.UseDefaultNetwork = true
	if errs := c.Prepare(nil); len(errs) != 0 {
		t.Fatalf("Shouldn't raise error: %s", errs)
	}

	c.UseDefaultNetwork = true
	c.VpcId = "vpc-id-test"
	if errs := c.Prepare(nil); len(errs) != 1 {
		t.Fatalf("Should raise an error: %s", errs)
	}

	c.UseDefaultNetwork = true
	c.VpcId = ""
	c.SubnetId = "subnet-id-test"
	if errs := c.Prepare(nil); len(errs) != 1 {
		t.Fatalf("Should raise an error: %s", errs)
	}

	c.UseDefaultNetwork = true
	c.VpcId = ""
	c.SubnetId = ""
	c.SecurityGroupId = "security-group-id-test"
	if errs := c.Prepare(nil); len(errs) != 1 {
		t.Fatalf("Should raise an error: %s", errs)
	}
}

func TestRunConfigPrepare_WithVpcId(t *testing.T) {
	c := getTestRunConfig()

	// have vpc_id
	c.UseDefaultNetwork = false
	c.VpcId = "vpc-id-test"

	c.CidrBlock = "10.0.0.0/16"
	c.SubnetId = "subnet-id-test"
	if errs := c.Prepare(nil); len(errs) != 1 {
		t.Fatalf("Should raise an error: %s", errs)
	}
	c.CidrBlock = ""

	c.VpcName = "vpc-name-test"
	if errs := c.Prepare(nil); len(errs) != 1 {
		t.Fatalf("Should raise an error: %s", errs)
	}
	c.VpcName = ""

	c.SubnetId = ""
	c.SubnetCidrBlock = ""
	if errs := c.Prepare(nil); len(errs) != 1 {
		t.Fatalf("Should raise an error: %s", errs)
	}

	c.SubnetId = "subnet-id-test"
	c.SecurityGroupId = ""
	c.SecurityGroupName = ""
	if errs := c.Prepare(nil); len(errs) != 0 {
		t.Fatalf("Shouldn't raise error: %s", errs)
	}
	if c.SecurityGroupName == "" {
		t.Fatalf("security group name shouldn't be empty")
	}
	c.SecurityGroupName = ""
}

func TestRunConfigPrepare_NoVpcId(t *testing.T) {
	c := getTestRunConfig()

	// no vpc id
	c.UseDefaultNetwork = false
	c.VpcId = ""

	c.SubnetId = "subnet-id-test"
	if errs := c.Prepare(nil); len(errs) != 1 {
		t.Fatalf("Should raise an error: %s", errs)
	}
	c.SubnetId = ""

	c.SecurityGroupId = "security-group-id-test"
	if errs := c.Prepare(nil); len(errs) != 1 {
		t.Fatalf("Should raise an error: %s", errs)
	}
	c.SecurityGroupId = ""

	c.VpcName = ""
	c.CidrBlock = ""
	c.SubnetName = ""
	c.SubnetCidrBlock = ""
	c.SecurityGroupName = ""
	if errs := c.Prepare(nil); len(errs) != 0 {
		t.Fatalf("Shouldn't raise error: %s", errs)
	}
	if c.VpcName == "" || c.CidrBlock == "" || c.SubnetName == "" || c.SubnetCidrBlock == "" ||
		c.SecurityGroupName == "" {
		t.Fatalf("The value of vpcname, cidrblock, subnetname, subnetcidrblock, securitygroupname shouldn't be empty")
	}
}

func TestRunConfigPrepare_Eip(t *testing.T) {
	c := getTestRunConfig()

	c.AssociatePublicIpAddress = true
	c.EipName = ""
	c.InternetChargeType = ""
	c.NetworkCapacityInMbps = 0
	if errs := c.Prepare(nil); len(errs) != 0 {
		t.Fatalf("Shouldn't raise error: %s", errs)
	}
	if c.EipName == "" || c.InternetChargeType == "" || c.NetworkCapacityInMbps == 0 {
		t.Fatal("the value of eipname, internetchargetype, networkcapcityinmbps shouldn't be empty")
	}

	c.AssociatePublicIpAddress = true
	c.EipName = ""
	c.InternetChargeType = "no such charge type"
	c.NetworkCapacityInMbps = 0
	if errs := c.Prepare(nil); len(errs) != 1 {
		t.Fatalf("Should raise an error: %s", errs)
	}

	c.AssociatePublicIpAddress = false
	c.EipName = "eip-name-test"
	c.InternetChargeType = ""
	c.NetworkCapacityInMbps = 0
	if errs := c.Prepare(nil); len(errs) != 1 {
		t.Fatalf("Should raise an error: %s", errs)
	}

}

func TestRunConfigPrepare_DataDisk(t *testing.T) {
	c := getTestRunConfig()

	c.DataDisks = []BaiduCloudDataDisk{
		{
			SnapShotId: "",
		},
	}
	if errs := c.Prepare(nil); len(errs) != 1 {
		t.Fatalf("Should raise an error: %s", errs)
	}
}
