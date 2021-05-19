resource "aws_instance" "front" {
  ami = data.aws_ami.debian.id

  count = var.instance_count
  instance_type = var.instance_type

  root_block_device {
    volume_size = var.disk_size
    volume_type = var.disk_type
    delete_on_termination = true
  }
}

resource "aws_elb" "front" {
  listener {
    lb_port = 80
    lb_protocol = "tcp"
    instance_port = 80
    instance_protocol = "tcp"
  }

  instances = [aws_instance.front[0].id]
}

module "ebs" {
  source = "./module-ebs"
  availability_zone = aws_instance.front[0].availability_zone
}

data "aws_ami" "debian" {
  most_recent = true

  filter {
    name   = "name"
    values = ["debian-stretch-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  filter {
    name   = "architecture"
    values = ["x86_64"]
  }

  filter {
    name   = "root-device-type"
    values = ["ebs"]
  }

  owners = ["379101102735"] # Debian
}
