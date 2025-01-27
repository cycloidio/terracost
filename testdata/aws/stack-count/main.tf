module "example" {
  source = "./ec2_instances"

  instances = [
    {
      ami           = "bar"
      instance_type = "t3.small"
    }
  ]

  instances_count = 3
}

variable "instances" {
  description = "instance inputs"
  type = list(object({
    ami           = string
    instance_type = string
  }))
  default = [
    {
      ami           = "bar"
      instance_type = "t3.small"
    }
  ]
}
