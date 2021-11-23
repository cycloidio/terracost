###
# Cycloid Requirements
###
variable "env" {
  description = "Cycloid project name."
}

variable "project" {
  description = "Cycloid environment name."
}

variable "customer" {
  description = "Cycloid customer name."
}

###
# FIREWALL 
###

# Firewall - ingress
variable "ingress_firewall_name" {
  description = "Name of the firewall ingress resource."
}

variable "ingress_firewall_description" {
  description = "A brief description of the ingress firewall."
  default = ""
}

variable "ingress_disabled" {
  description = "Denotes whether the firewall ingress rule is disabled."
  default     = false
}

variable "ingress_source_ranges" {
  description = "If specified the firewall will only be applied to the source IP address in these ranges."
  default     = []
}

variable "ingress_source_tags" {
  description = "If source tags are specified, the firewall will apply only to traffic with source IP that belongs to a tag listed in source tags. "
  default     = []
}

variable "ingress_allow_protocol" {
  description = "The IP protocol to which the ingress allow rule applies. "
  default     = "tcp"
}

variable "ingress_allow_ports" {
  description = "An otional list of IP ports to which the ingress allow rule applies."
  default     = ["22"]
}



# Firewall - egress
variable "egress_firewall_name" {
  description = "Name of the firewall egress resource."
}
variable "egress_firewall_description" {
  description = "A brief description of the egress firewall."
  default = ""
}

variable "egress_disabled" {
  description = "Denotes whether the firewall egress rule is disabled"
  default     = true
}

variable "egress_destination_ranges" {
  description = "Lists the destination IP address, as CIDR, to apply egress firewalls rules."
  default     = []
}

variable "egress_allow_protocol" {
  description = "The IP protocol to which the egress allow rule applies. "
  default     = ""
}

variable "egress_allow_ports" {
  description = "An otional list of IP ports to which the egress allow rule applies."
  default     = []
}

###
# Cloud init template
###
variable "file_content" {
  description = "The content of the file to use if cloud init is used."
}

###
# Virtual Machine
###
# VM - general speficiations
variable "instance_name" {
  description = "The unique name for the instance."
}

variable "instance_description" {
  description = "A brief description of the instance."
  default = ""
}

variable "machine_type" {
  description = "The machine type to create."
  default     = "e2-small"
}

variable "allow_stopping_for_update" {
  description = "Allows to stop the instance to update its properties."
  default     = true
}

# VM - tags
variable "instance_tags" {
  description = "A list of network tags to attach to the instance."
}

# VM - storage
variable "boot_disk_auto_delete" {
  description = "Eanbles disk deletion when the instance is deleted."
  default     = true
}

variable "boot_disk_device_name" {
  description = "Name with which attached disk will be accessible, as /dev/disk/by-id/google-{{device_name}}"
  default     = ""
}

variable "boot_disk_image" {
  description = "The image from which to initialize this disk."
  default     = "debian-cloud/debian-10"
}

variable "boot_disk_size" {
  description = "The size of the boot disk image in gigabytes."
}

variable "boot_disk_type" {
  description = "The GCE disk type"
  default     = ""
}

# VM - network 

variable "network" {
  description = "The name or self_link of the network to attach this interface to."
}

variable "network_ip" {
  description = "The private IP address to assign to the instance."
  default     = ""
}

# VM - labels
variable "instance_extra_labels" {
  description = "A map of key/value label pairs to assign to the instance."
  default     = {}
}

locals {
  standard_labels = {
    "cycloid.io" = "true"
    env          = var.env
    project      = var.project
    client       = var.customer
    organization = var.customer
  }
  instance_labels = merge(var.instance_extra_labels, local.standard_labels)

  # if names not set take by default ${var.customer}-${var.project}-${var.env}-object
  ingress_name = var.ingress_firewall_name != "" ? var.ingress_firewall_name : "${var.customer}-${var.project}-${var.env}-ingress"
  egress_name = var.egress_firewall_name != "" ? var.egress_firewall_name : "${var.customer}-${var.project}-${var.env}-egress"
  instance_name = var.instance_name != "" ? var.instance_name : "${var.customer}-${var.project}-${var.env}-vm"
  instance_tags = var.instance_tags != [] ? var.instance_tags : ["${var.customer}-${var.project}-${var.env}-network-tag"]
}
