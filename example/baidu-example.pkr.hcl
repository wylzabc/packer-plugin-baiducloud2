variable "access_key" {
  type    = string
  default = "${env("BAIDUCLOUD_ACCESS_KEY")}"
}

variable "secret_key" {
  type    = string
  default = "${env("BAIDUCLOUD_SECRET_KEY")}"
}

source "baiducloud-bcc" "example-1" {
  access_key      = "${var.access_key}"
  secret_key      = "${var.secret_key}"
  region          = "fwh"
  zone            = "cn-fwh-a"
  source_image_id = "m-sDghxZu1"
  instance_spec   = "bcc.ic4.c2m2"
  image_name      = "your-packer-custom-image"
  associate_public_ip_address = true 
  use_default_network = true 
  ssh_username    = "root"
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
