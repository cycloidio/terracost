provider "aws" {
  region = "eu-west-1"
}

variable "front_ebs_optimized" {
  default = true
}

variable "enable_mon" {
  type = bool
}

variable "env" {
  type = string
}

module "magento" {
  source              = "./module"
  enable_mon          = var.enable_mon
  front_ebs_optimized = var.front_ebs_optimized
  env                 = var.env
}
