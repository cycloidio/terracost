variable "instances" {
  description = "instance inputs"
  type = list(object({
    ami           = string
    instance_type = string
  }))
}

variable "instances_count" {
  default = 2
}


resource "aws_instance" "instances" {

  for_each               = { for index, instance in var.instances : index => instance }
  ami                    = each.value.ami
  instance_type          = each.value.instance_type
  vpc_security_group_ids = ["foo"]
  subnet_id              = "foo"
  root_block_device {
    delete_on_termination = true
    encrypted             = true
    volume_type           = "gp3"
    volume_size           = 10
  }
  disable_api_termination = false

}


resource "aws_instance" "instancescount" {

  count                  = var.instances_count
  ami                    = "foo"
  instance_type          = "t3.medium"
  vpc_security_group_ids = ["foo"]
  subnet_id              = "foo"
  root_block_device {
    delete_on_termination = true
    encrypted             = true
    volume_type           = "gp3"
    volume_size           = 10
  }
  disable_api_termination = false
}
