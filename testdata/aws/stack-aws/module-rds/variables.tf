variable "storage_size" {
  default = 10
}

variable "storage_type" {
  default = "gp2"
}

variable "instance_class" {
  default = "db.t3.small"
}

variable "multi_az" {
  default = false
}

variable "rds_engine" {
  default = "mysql"
}
