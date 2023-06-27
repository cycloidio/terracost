resource "aws_eks_cluster" "example" {
  name     = "example"
  role_arn = "arn:aws:iam::123456789012:user/johndoe"

  vpc_config {
    subnet_ids = ["1", "2"]
  }
}
