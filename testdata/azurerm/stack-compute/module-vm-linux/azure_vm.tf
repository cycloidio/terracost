###
# Network interface
###

resource "azurerm_public_ip" "vm_pub_ip" {
  name                = local.public_ip_name
  location            = var.azure_location
  resource_group_name = var.resource_group_name
  allocation_method   = "Dynamic"
  tags                = local.network_tags
}


resource "azurerm_network_interface" "vm_net_interface" {
  name                = local.network_interface_name
  location            = var.azure_location
  resource_group_name = var.resource_group_name

  ip_configuration {
    name                          = local.ip_config_name
    subnet_id                     = var.subnet_id
    private_ip_address_allocation = "Dynamic"
    public_ip_address_id          = azurerm_public_ip.vm_pub_ip.id
    primary                       = true
  }
  tags = local.network_tags

}

###
# Network Security Group
###

# Create Network Security Group and rule
resource "azurerm_network_security_group" "vm_sg" {
  name                = local.network_security_group_name
  location            = var.azure_location
  resource_group_name = var.resource_group_name

  security_rule {
    name                       = var.security_rule_name
    description                = var.security_rule_description
    priority                   = var.security_rule_priority
    direction                  = var.security_rule_direction
    access                     = var.security_rule_access
    protocol                   = var.security_rule_protocol
    source_port_range          = var.security_rule_source_port_range
    destination_port_range     = var.security_rule_destination_port_range
    source_address_prefix      = var.security_rule_source_address_prefix
    destination_address_prefix = var.security_rule_destination_address_prefix
  }

  tags = local.sg_tags
}

# Connect the security group to the network interface
resource "azurerm_network_interface_security_group_association" "vm_sg_assocation" {
  network_interface_id      = azurerm_network_interface.vm_net_interface.id
  network_security_group_id = azurerm_network_security_group.vm_sg.id
}

###
# Cloud init template
###

data "template_file" "user_data" {
  template = file("${path.module}/cloud-init.sh.tpl")
  vars = {
    file_content  = var.file_content
  }
}


###
# Virtual Machine
###

resource "azurerm_virtual_machine" "main" {
  // general configuration
  name                          = local.instance_name
  location                      = var.azure_location
  resource_group_name           = var.resource_group_name
  network_interface_ids         = [azurerm_network_interface.vm_net_interface.id]
  vm_size                       = var.vm_size
  delete_os_disk_on_termination = var.delete_os_disk_on_termination

  storage_image_reference {
    publisher = var.image_publisher
    offer     = var.image_offer
    sku       = var.image_sku
    version   = var.image_version
    // required for costume_image
    id = var.image_id
  }
  storage_os_disk { //Required
    name              = local.disk_name
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = var.disk_managed_type
    disk_size_gb      = var.disk_size
  }

  os_profile {
    computer_name  = var.os_computer_name
    admin_username = var.os_admin_username
    admin_password = var.os_admin_password
    custom_data    = data.template_file.user_data.rendered
  }

  // Required, when a Linux machine
  os_profile_linux_config {
    disable_password_authentication = var.disable_linux_password_authentification
  }

  tags = local.instance_tags
}