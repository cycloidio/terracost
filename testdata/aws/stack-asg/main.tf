provider "aws" {
  region = "eu-west-1"
}

variable "front_ebs_optimized" {
  default = true
}

variable "project" {
  type = bool
}

module "magento" {
  source = "./module"
  project = var.project
  front_ebs_optimized = var.front_ebs_optimized
}
