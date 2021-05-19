resource "aws_db_instance" "db" {
  instance_class = var.instance_class
  allocated_storage = var.storage_size
  storage_type = var.storage_type
  multi_az = var.multi_az
  engine = var.rds_engine
}
