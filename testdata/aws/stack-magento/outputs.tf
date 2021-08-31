output "elb_front_dns_name" {
  value       = module.magento.elb_front_dns_name
  description = "DNS name of the front elb."
}

output "elb_front_zone_id" {
  value       = module.magento.elb_front_zone_id
  description = "Zone ID of the front ELB."
}

output "front_private_ips" {
  value       = module.magento.front_private_ips
  description = "Private IPs of the front EC2 server."
}

output "cache_address" {
  value       = module.magento.cache_address
  description = "Address of the ElastiCache."
}

output "cache_cluster_id" {
  value       = module.magento.cache_cluster_id
  description = "Cluster Id of the ElastiCache."
}

output "rds_address" {
  value       = module.magento.rds_address
  description = "Address of the RDS database."
}

output "rds_port" {
  value       = module.magento.rds_port
  description = "Port of the RDS database."
}

output "rds_username" {
  value       = module.magento.rds_username
  description = "Username of the RDS database."
}

output "rds_password" {
  value     = var.rds_password
  sensitive = true
}

output "rds_database" {
  value       = module.magento.rds_database
  description = "Database name of the RDS database."
}
