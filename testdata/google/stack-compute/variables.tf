# Cycloid requirements
variable "project" {
  description = "Cycloid project name."
}

variable "env" {
  description = "Cycloid environment name."
}

variable "customer" {
  description = "Cycloid customer name."
}

# GCP
variable "gcp_project" {
  description = "GCP project to launch vm."
}

variable "gcp_region" {
  description = "GCP region to launch vm."
  default     = "europe-west1"
}

variable "gcp_zone" {
  description = "GCP zone to launch vm."
  default     = "europe-west1-c"
}