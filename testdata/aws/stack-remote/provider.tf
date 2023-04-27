provider "aws" {
  #version    = "1.40.0"
  access_key = var.access_key
  secret_key = var.secret_key
  region     = var.aws_region
}

variable "customer" {
}

variable "project" {
}

variable "env" {
}

variable "access_key" {
}

variable "secret_key" {
}

variable "rds_password" {
  default = "ChangeMePls"
}

variable "aws_region" {
  description = "AWS region to launch servers."
  default     = "eu-west-1"
}
