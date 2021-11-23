output "ip_address" {
  value       = module.vm.vm_public_ip_address
  description = "IP of the VM"
}