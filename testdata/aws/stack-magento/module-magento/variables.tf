data "aws_region" "current" {}

variable "bastion_sg_allow" {
}

variable "env" {
}

variable "customer" {
}

variable "zones" {
  type    = list(string)
  default = []
}

data "aws_availability_zones" "available" {
  state = "available"
}

locals {
  aws_availability_zones = length(var.zones) > 0 ? var.zones : data.aws_availability_zones.available.names
}

variable "keypair_name" {
  default = "demo"
}

variable "extra_tags" {
  default = {}
}

locals {
  standard_tags = {
    "cycloid.io" = "true"
    env          = var.env
    project      = var.project
    client       = var.customer
  }
  merged_tags = merge(local.standard_tags, var.extra_tags)
}

variable "private_subnets_ids" {
  type    = list(string)
  default = [""]
}

variable "public_subnets_ids" {
  type = list(string)
}

# variable "private_zone_id" {}

variable "cache_subnet_group" {
  default = ""
}

variable "vpc_id" {
  default = ""
}

variable "project" {
  default = "magento"
}

variable "rds_database" {
  default = "magento"
}

variable "rds_disk_size" {
  default = 10
}

variable "rds_multiaz" {
  default = false
}

variable "rds_password" {
  default = "ChangeMePls"
}

variable "rds_type" {
  default = "db.t3.small"
}

variable "rds_username" {
  default = "magento"
}

variable "rds_engine" {
  default = "mysql"
}

variable "rds_engine_version" {
  default = "5.7.16"
}

variable "rds_backup_retention" {
  default = 7
}

variable "rds_parameters" {
  default = "default.mysql5.7"
}

variable "rds_subnet_group" {
  default = ""
}

variable "rds_storage_type" {
  default = "gp2"
}

variable "rds_skip_final_snapshot" {
  default = false
}

###

# front

###

variable "magento_ssl_cert" {
}

variable "front_disk_size" {
  default = 60
}

variable "front_disk_type" {
  default = "gp2"
}

variable "front_type" {
  default = "t3.small"
}

variable "front_ebs_optimized" {
  default = false
}

variable "front_count" {
  default = 1
}

variable "debian_ami_name" {
  default = "debian-stretch-*"
}

###

# ElastiCache

###

variable "elasticache_type" {
  default = "cache.t2.micro"
}

variable "elasticache_nodes" {
  default = 1
}

variable "elasticache_engine" {
  default = "redis"
}

variable "elasticache_parameter_group_name" {
  default = "default.redis5.0"
}

variable "elasticache_engine_version" {
  default = "5.0.0"
}

variable "elasticache_port" {
  default = "6379"
}

variable "elasticache_cluster_id" {
  default = ""
}

resource "random_string" "elasticache_cluster_id" {
  length  = 18
  upper   = false
  special = false
}

locals {
  elasticache_cluster_id = var.elasticache_cluster_id != "" ? var.elasticache_cluster_id : "cy${random_string.elasticache_cluster_id.result}"
}
