variable "create_bucket" {
  description = "Controls if S3 bucket should be created"
  type        = bool
  default     = true
}

variable "metric_configuration" {
  description = "Map containing bucket metric configuration."
  type        = any
  default     = []
}


locals {
  create_bucket        = var.create_bucket
  metric_configuration = try(jsondecode(var.metric_configuration), var.metric_configuration)
}


resource "aws_s3_bucket" "this" {
  count  = local.create_bucket ? 1 : 0
  bucket = "foobar"
}

resource "aws_s3_bucket_metric" "this" {
  for_each = { for k, v in local.metric_configuration : k => v if local.create_bucket }

  name   = each.value.name
  bucket = aws_s3_bucket.this[0].id
}


