module "vm" {
  #####################################
  # Do not modify the following lines #
  source = "./module-vm-linux"

  project  = var.project
  env      = var.env
  customer = var.customer

  ##################################### 

  ###
  # Network interface
  ###

  #. network_interface_name (optional, string):
  #+ The name of the Network Interface.
  network_interface_name = "${var.customer}-${var.project}-${var.env}-nic"

  #. public_ip_name (optional, string):
  #+ Specifies the name of the Public IP resource.
  public_ip_name = "${var.customer}-${var.project}-${var.env}-public_ip"

  #. ip_config_name (optional, string):
  #+ A name used for the IP Configuration in the network interface.
  ip_config_name = "${var.customer}-${var.project}-${var.env}-ip_config"

  #. subnet_id (optional, string):
  #+ The ID of the Subnet where this Network Interface should be located in.
  subnet_id = ""

  #. network_extra_tags (optional):
  #+ Map of extra tags to assign to the network resources.
  network_extra_tags = {}

  ###
  # Network Security Group
  ###

  #. network_security_group_name (optional, string):
  #+ Specifies the name of the Application Security Group.
  network_security_group_name = "${var.customer}-${var.project}-${var.env}-sg"

  #. security_rule_name (required, string):
  #+ The name of the default security rule.
  security_rule_name = "SSH"

  #. security_rule_description (optional, string):
  #+ A description of the default rule.
  security_rule_description = "Enable SSH inbound traffic."

  #. security_rule_priority (required, integer): 
  #+ Specifies the priority of the default rule.
  security_rule_priority = 1001

  #. security_rule_direction (required, string):
  #+ Specifies if default rule will be evaluated on incoming or outgoing traffic.
  security_rule_direction = "Inbound"

  #. security_rule_access (required, string):
  #+ Specifies whether network traffic is allowed or denied by default rule.
  security_rule_access = "Allow"

  #. security_rule_protocol (required, string):
  #+ Network protocol that default rule applies to.
  security_rule_protocol = "Tcp"

  #. security_rule_source_port_range (required, string):
  #+ Default rule source port or range.
  security_rule_source_port_range = "*"

  #. security_rule_destination_port_range (required, string):
  #+ "Default rule destination port or range."
  security_rule_destination_port_range = "22"

  #. security_rule_source_address_prefix (required, string):
  #+ "Default rule, CIDR or source IP range or * to match any IP."
  security_rule_source_address_prefix = "*"

  #. security_rule_destination_address_prefix (required, string):
  #+ "Lists of destination address prefixes to match in the default rule."
  security_rule_destination_address_prefix = "*"

  #. sg_extra_tags (optional):
  #+ Map of extra tags to assign to the security group.
  sg_extra_tags = {}

  ###
  # Cloud init template
  ###
  #. file_content (optional, string):
  #+ The content of the file to use if cloud init is used.
  file_content = ""

  ###
  # Virtual Machine
  ###

  # VM- General configuration

  #. instance_name (optional, string):
  #+ Specifies the name of the Virtual Machine.
  instance_name = "${var.customer}-${var.project}-${var.env}-vm"

  #. azure_location (required, string):
  #+ Specifies the supported Azure location where the resources exist.
  azure_location = var.azure_location

  #. resource_group_name (required, string):
  #+ The name of the resource group to use for the creation of resources.
  resource_group_name = ""

  #. vm_size (required, string):
  #+ Specifies the size of the Virtual Machine.
  vm_size = "Standard_DS1_v2"

  #. delete_os_disk_on_termination (required, boolean):
  #+ Enables deleting the OS disk automatically when deleting the VM.
  delete_os_disk_on_termination = true

  #. os_computer_name (required, string):
  #+ Specifies the name of the Virtual Machine.
  os_computer_name = "cycloid"

  #. os_admin_username (required, string):
  #+ Specifies the name of the local admin account.
  os_admin_username = "admin"

  #. os_admin_password (required, string):
  #+ The password associated with the local admin account. Must be [6-72] and contain uppercase + lowercase + number + special caracter
  os_admin_password = "MyPassWordToChange_2021"

  #. disable_linux_password_authentification (required, boolean):
  #+ Specifies whether password authentication should be disabled.
  disable_linux_password_authentification = false

  #. instance_extra_tags (optional):
  #+ A map of tags to assign to the resource.
  instance_extra_tags = {}

  # VM - image

  #. image_publisher (required, string):
  #+ Specifies the publisher of the image used to create the virtual machine.
  image_publisher = "debian"

  #. image_offer (required, string):
  #+ Specifies the offer of the image used to create the virtual machine.
  image_offer = "debian-10"

  #. image_sku (required, string):
  #+ Specifies the SKU of the image used to create the virtual machine.
  image_sku = "10-cloudinit-gen2"

  #. image_version (optional, string):
  #+ Specifies the version of the image used to create the virtual machine.
  image_version = "latest"

  #. image_id (optional, string):
  #+ Specifies the ID of the Custom Image which the Virtual Machine should be created from
  image_id = ""


  # VM - storage

  #. disk_name (optional, string):
  #+ Specifies the name of the OS Disk.
  disk_name = "${var.customer}-${var.project}-${var.env}-disk"

  #. disk_managed_type (optional, string):
  #+ Specifies the type of Managed Disk which should be created.
  disk_managed_type = "Standard_LRS"

  #. disk_size (optional, string):
  #+ Specifies the name of the OS Disk size in gigabytes.
  disk_size = 5
}
