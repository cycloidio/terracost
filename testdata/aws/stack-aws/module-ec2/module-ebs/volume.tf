resource "aws_ebs_volume" "volume" {
  availability_zone = var.availability_zone
  size = var.size
  type = var.type
}
