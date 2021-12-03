output "vm_public_ip_address" {
  value = google_compute_instance.vm.network_interface.0.access_config.0.nat_ip
}

output "vm_instance_id" {
  value = google_compute_instance.vm.id
}

output "vm_target_tags" {
  value = var.instance_tags
}
