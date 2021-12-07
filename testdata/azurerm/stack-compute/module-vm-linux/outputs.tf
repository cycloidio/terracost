output "ip_address" {
  value = azurerm_public_ip.vm_pub_ip.ip_address
}

output "resource_group_name" {
  value       = var.resource_group_name
  description = "The name of the resource group to use for the creation of resources."
}

output "network_security_group_name" {
  value       = local.network_security_group_name
  description = "Specifies the name of the Application Security Group."
}