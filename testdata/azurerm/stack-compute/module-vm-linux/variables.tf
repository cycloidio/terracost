###
# Cycloid requirements
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
# Network interface
###

variable "public_ip_name" {
  description = "Specifies the name of the Public IP resource."
  default     = ""
}

variable "network_interface_name" {
  description = "The name of the Network Interface."
  default     = ""
}

variable "ip_config_name" {
  description = "A name used for the IP Configuration in the network interface."
  default     = ""
}

variable "subnet_id" {
  description = "The ID of the Subnet where this Network Interface should be located in."
}

variable "network_extra_tags" {
  description = "A map of tags to assign to the network resources."
  default     = {}
}

###
# Network Security Group
###
variable "network_security_group_name" {
  description = "Specifies the name of the Application Security Group."
  default     = ""
}

variable "security_rule_name" {
  description = "The name of the default security rule."
  default     = "SSH"
}

variable "security_rule_description" {
  description = "A description of the default rule."
  default     = "Enable SSH inbound traffic."
}

variable "security_rule_priority" {
  description = "Specifies the priority of the default rule."
  default     = 1001
}

variable "security_rule_direction" {
  description = " Specifies if default rule will be evaluated on incoming or outgoing traffic."
  default     = "Inbound"
}

variable "security_rule_access" {
  description = "Specifies whether network traffic is allowed or denied by default rule."
  default     = "Allow"
}

variable "security_rule_protocol" {
  description = "Network protocol that default rule applies to."
  default     = "Tcp"
}

variable "security_rule_source_port_range" {
  description = "Default rule source port or range."
  default     = "*"
}

variable "security_rule_destination_port_range" {
  description = "Default rule destination port or range."
  default     = "22"
}

variable "security_rule_source_address_prefix" {
  description = "Default rule, CIDR or source IP range or * to match any IP."
  default     = "*"
}

variable "security_rule_destination_address_prefix" {
  description = "Lists of destination address prefixes to match in the default rule."
  default     = "*"
}

variable "sg_extra_tags" {
  description = "A map of tags to assign to the security_group resources."
  default     = {}
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

# VM- General configuration
variable "instance_name" {
  description = "Specifies the name of the Virtual Machine."
  default     = ""
}

variable "azure_location" {
  description = "Specifies the supported Azure location where the resources exist."
}

variable "resource_group_name" {
  description = "The name of the resource group to use for the creation of resources."
}

variable "vm_size" {
  description = "Specifies the size of the Virtual Machine."
  default     = "Standard_DS1_v2"
}

variable "delete_os_disk_on_termination" {
  description = "Enables deleting the OS disk automatically when deleting the VM."
  default     = "true"
}

variable "os_computer_name" {
  description = "Specifies the name of the Virtual Machine."
  default     = "cycloid"
}

variable "os_admin_username" {
  description = "Specifies the name of the local admin account."
  default     = "admin"
}

variable "os_admin_password" {
  description = "The password associated with the local admin account. Must be [6-72] and contain uppercase + lowercase + number + special caracter"
}

variable "disable_linux_password_authentification" {
  description = "Specifies whether password authentication should be disabled."
  default     = "false"
}

variable "instance_extra_tags" {
  description = "A map of tags to assign to the resource."
  default     = {}
}

# VM - image
variable "image_publisher" {
  description = "Specifies the publisher of the image used to create the virtual machine."
  default     = "debian"
}

variable "image_offer" {
  description = "Specifies the offer of the image used to create the virtual machine."
  default     = "debian-10"
}

variable "image_sku" {
  description = "Specifies the SKU of the image used to create the virtual machine."
  default     = "10-cloudinit-gen2"
}

variable "image_version" {
  description = "Specifies the version of the image used to create the virtual machine."
  default     = "latest"
}

variable "image_id" {
  description = "Specifies the ID of the Custom Image which the Virtual Machine should be created from"
  default     = ""
}

# VM - storage
variable "disk_name" {
  description = "Specifies the name of the OS Disk."
  default     = ""
}

variable "disk_managed_type" {
  description = "Specifies the type of Managed Disk which should be created."
  default     = "Standard_LRS"
}

variable "disk_size" {
  description = "Specifies the name of the OS Disk size in gigabytes."
}

###
# TAGS + names
###
locals {
  standard_tags = {
    "cycloid.io" = "true"
    env          = var.env
    project      = var.project
    client       = var.customer
    organization = var.customer
  }
  network_tags  = merge(var.network_extra_tags, local.standard_tags)
  sg_tags       = merge(var.sg_extra_tags, local.standard_tags)
  instance_tags = merge(var.instance_extra_tags, local.standard_tags)

  # if names not set take by default ${var.customer}-${var.project}-${var.env}-object
  public_ip_name              = var.public_ip_name != "" ? var.public_ip_name : "${var.customer}-${var.project}-${var.env}-public_ip"
  network_interface_name      = var.network_interface_name != "" ? var.network_interface_name : "${var.customer}-${var.project}-${var.env}-nic"
  ip_config_name              = var.ip_config_name != "" ? var.ip_config_name : "${var.customer}-${var.project}-${var.env}-ip_config"
  network_security_group_name = var.network_security_group_name != "" ? var.network_security_group_name : "${var.customer}-${var.project}-${var.env}-sg"
  instance_name               = var.instance_name != "" ? var.instance_name : "${var.customer}-${var.project}-${var.env}-vm"
  disk_name                   = var.disk_name != "" ? var.disk_name : "${var.customer}-${var.project}-${var.env}-disk"
}