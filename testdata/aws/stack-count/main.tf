module "example" {
  source = "./ec2_instances"

  count = length(var.instances)
  instances = [
    {
      ami           = "bar"
      instance_type = "t3.small"
    },
    {
      ami           = "bar2"
      instance_type = "t2.small"
    }
  ]

  secrets = {
    "user_passwords_${var.chamber_name}" = {
      description             = "Secret for multiple users with random passwords"
      recovery_window_in_days = 0
      instances_count         = length(var.instances)
    }
  }

  instances_count = 3
}

variable "chamber_name" {
  type    = string
  default = "test"
}

variable "secrets" {
  description = "Map of secrets to keep in AWS Secrets Manager"
  type        = any
  default     = {}
}


variable "instances" {
  description = "instance inputs"
  type = list(list(object({
    ami             = string
    instance_type   = string
    disk_gb         = string
    name            = string
    user_data       = string
    imdsv1_disabled = bool
    application     = optional(string, "")
  })))
  default = []
}
