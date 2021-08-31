module "magento" {

  #####################################
  # Do not modify the following lines #
  source = "./module-magento"
  env      = var.env
  customer = var.customer
  project  = var.project
  #####################################

  #. vpc_id (required):
  #+ Amazon VPC id on which create each components.
  vpc_id                   = "<vpc-id>"

  #. private_subnets_ids (required, array):
  #+ Amazon subnets IDs on which create each components.
  private_subnets_ids      = ["private-subnets"]

  #. magento_ssl_cert (required):
  #+ ARN of an Amazon certificate from Certificate Manager.
  magento_ssl_cert         = "<ssl-cert-arn>"

  #. bastion_sg_allow (optional):
  #+ Amazon source security group ID which will be allowed to connect on Magento front port 22 (ssh).
  bastion_sg_allow         = "<bastion-sg>"

  #. public_subnets_ids (required, array):
  #+ Public subnet IDs to use for the public ELB load balancer.
  public_subnets_ids       = ["public-subnets"]

  #. rds_password (optional): var.rds_password to get it from the pipeline.
  #+ Password of the RDS database.
  rds_password             = var.rds_password

  #. keypair_name (optional): demo
  #+ SSH keypair name to use to deploy ec2 instances.
  keypair_name = "demo"

  #. extra_tags (optional): {}
  #+ Dict of extra tags to add on aws resources. format { "foo" = "bar" }.

  #. rds_database (optional): magento
  #+ Name of the RDS database.
  rds_database         = "magento"

  #. rds_disk_size (optional): 10
  #+ Size in Go of the RDS database.
  rds_disk_size        = 10

  #. rds_multiaz (optional, bool): false
  #+ Enable multi AZ or not for the RDS database.
  rds_multiaz          = false

  #. rds_subnet_group (optional): Automatically generated from private_subnets_ids
  #+ ID of the private DB subnet group to use for RDS database.
  # rds_subnet_group           = "<rds-subnet-group>"

  #. rds_type (optional): db.t3.small
  #+ AWS Instance type of the RDS database.
  rds_type             = "db.t3.small"

  #. rds_username (optional): magento
  #+ User name of the RDS database.
  rds_username         = "magento"

  #. rds_engine (optional): mysql
  #+ Amazon RDS engine to use.
  # rds_engine           = "mysql"

  #. rds_engine_version (optional): "5.7.16"
  #+ Version of the RDS engine.
  rds_engine_version   = "5.7.16"

  #. rds_backup_retention (optional): 7
  #+ RDS backup retention period in days.
  rds_backup_retention = 7

  #. rds_parameters (optional):
  #+ RDS parameters to assign to the RDS database.
  rds_parameters       = ""


  #. cache_subnet_group (optional): Automatically generated from private_subnets_ids
  #+ AWS elasticache subnet name.
  # cache_subnet_group                     = "cache-subnet-id"

  #. elasticache_type (optional): cache.t2.micro
  #+ AWS elasticache instance type.
  elasticache_type                 = "cache.t2.micro"

  #. elasticache_nodes (optional): 1
  #+ Number of AWS elasticache instances.
  elasticache_nodes                = "1"

  #. elasticache_parameter_group_name (optional): default.redis5.0
  #+ AWS elasticache parameter group name.
  # elasticache_parameter_group_name = "default.redis5.0"

  #. elasticache_engine (optional): redis
  #+ AWS elasticache engine.
  # elasticache_engine               = "redis"

  #. elasticache_engine_version (optional): "5.0.0"
  #+ AWS elasticache engine version.
  elasticache_engine_version               = "5.0.0"

  #. elasticache_engine (optional): 6379
  #+ AWS elasticache binding port.
  # elasticache_port                 = "6379"

  #. front_count (optional): 1
  #+ Number of Aws EC2 frontend server to create.
  front_count           = "1"

  #. front_disk_size (optional): 60
  #+ Disk size in Go of Aws EC2 frontend servers.
  front_disk_size       = 60

  #. front_type (optional): t3.small
  #+ Type of Aws EC2 frontend servers.
  front_type            = "t3.small"

  #. front_ebs_optimized (optional, bool): false
  #+ Whether the Instance is EBS optimized or not, related to the instance type you choose.
  front_ebs_optimized   = "false"
}
