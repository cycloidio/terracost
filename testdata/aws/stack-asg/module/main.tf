
variable "front_ebs_optimized" {
  default = false
}
variable "env" {
}

variable "enable_mon" {
}

resource "aws_launch_template" "foobar" {
  name_prefix   = "rds-${var.env}"
  image_id      = "ami-1a2b3c"
  instance_type = "t3.large"
  ebs_optimized = true
  placement {
    availability_zone = "eu-west-1c"
    tenancy           = "dedicated"
  }

  credit_specification {
    cpu_credits = "unlimited"
  }
  monitoring {
    enabled = var.enable_mon
  }

  block_device_mappings {
    device_name = "/dev/sda1"

    ebs {
      volume_size = 20
    }
  }
}

resource "aws_autoscaling_group" "lt" {
  availability_zones = ["eu-west-1a"]
  desired_capacity   = 2
  max_size           = 10
  min_size           = 1


  launch_template {
    id      = aws_launch_template.foobar.id
    version = "$Latest"
  }
}


resource "aws_launch_configuration" "lc" {
  #image_id      = data.aws_ami.ubuntu.id
  image_id          = "ami-123456789"
  instance_type     = "m4.large"
  placement_tenancy = "dedicated"
  enable_monitoring = true
}

resource "aws_autoscaling_group" "lc" {
  availability_zones   = ["eu-west-1a"]
  desired_capacity     = 3
  max_size             = 5
  min_size             = 1
  launch_configuration = aws_launch_configuration.lc.name
}


resource "aws_autoscaling_group" "mixed" {
  desired_capacity = 3
  max_size         = 15
  min_size         = 2

  mixed_instances_policy {
    instances_distribution {
      on_demand_base_capacity = 2
    }

    launch_template {
      launch_template_specification {
        launch_template_name = aws_launch_template.foobar.name
      }

      override {
        instance_type     = "c4.large"
        weighted_capacity = "3"
      }
    }
  }
}
