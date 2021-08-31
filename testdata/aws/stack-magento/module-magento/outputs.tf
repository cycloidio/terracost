# output "redis_magento_dns" {
#   value = "${aws_route53_record.redis.fqdn}"
# }
#
# output "rds_magento_dns" {
#   value = "${aws_route53_record.rds.fqdn}"
# }

#ELB

output "elb_front_dns_name" {
  value = aws_elb.front.dns_name
}

output "elb_front_zone_id" {
  value = aws_elb.front.zone_id
}

output "front_private_ips" {
  value = join(",", aws_instance.front.*.private_ip)
}

output "rds_address" {
  value = aws_db_instance.magento.address
}

output "rds_port" {
  value = aws_db_instance.magento.port
}

output "rds_database" {
  value = aws_db_instance.magento.name
}

output "rds_username" {
  value = aws_db_instance.magento.username
}

output "cache_address" {
  value = aws_elasticache_cluster.redis.cache_nodes[0].address
}

output "cache_cluster_id" {
  value = local.elasticache_cluster_id
}
