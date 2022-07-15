variable "access_key" {
  type    = string
  default = "${env("BAIDUCLOUD_ACCESS_KEY")}"
}

variable "secret_key" {
  type    = string
  default = "${env("BAIDUCLOUD_SECRET_KEY")}"
}

source "baiducloud-bcc" "example-1" {
  associate_public_ip_address = true 
<<<<<<< HEAD
//  use_default_network = true 
  source_image_id = "m-sDghxZu1"
  region          = "fwh"
  image_name      = "your-packer-custom-image-00002"
=======
  use_default_network = true 
  source_image_id = "m-sDghxZu1"
  region          = "fwh"
  image_name      = "your-packer-custom-image-00003"
>>>>>>> dev
  instance_spec   = "bcc.ic4.c2m2"
  zone            = "cn-fwh-a"
  access_key      = "${var.access_key}"
  secret_key      = "${var.secret_key}"
  eip_name        = "your-eip-name"
  network_capacity_in_mbps = 1
  root_disk_size_in_gb = 20
  ssh_username    = "root"
  temporary_key_pair_name = "your-packer-test"
  vpc_name            = "your-packer-test"
  vpc_cidr_block      = "192.168.0.0/16"
  subnet_name         = "your-packer-test"
  subnet_cidr_block   = "192.168.128.0/17"
  security_group_name = "your-packer-test"
//  packer_debug        = true
//  data_disks {
//    cds_size_in_gb = 50
//    storage_type   = "ssd"
//    snapshot_id    = ""
//  }
  run_tags            = {
    image_type_tag    = "custom"
    master            = "wangyl"
  }
} 

build {
  sources = ["source.baiducloud-bcc.example-1"]
  
  provisioner "shell" {
    inline = [
      "cd ~ && echo 'this is a file made in custom image' >> custom && cat custom",
      "yum install redis.x86_64 -y",
    ]
  }
}
