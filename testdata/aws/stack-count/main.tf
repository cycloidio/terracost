module "example" {
  source = "./ec2_instances"

  count = length(var.instances)
  #instances = [
  #{
  #ami           = "bar"
  #instance_type = "t3.small"
  #}
  #]
  instances = var.instances[count.index]

  #instances = {}
  #instances = null

  instances_count = 3
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
  default = [null]
}
