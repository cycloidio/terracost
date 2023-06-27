resource "aws_launch_template" "foo" {
  name = "foo"

  block_device_mappings {
    device_name = "/dev/sdf"

    ebs {
      volume_size = 200
    }
  }

  ebs_optimized = true

  instance_type = "t2.micro"
}

resource "aws_eks_node_group" "example" {
  cluster_name    = aws_eks_cluster.example.name
  node_group_name = "example"
  node_role_arn   = "arn:aws:iam::123456789012:user/johndoe"
  subnet_ids      = ["1", "2"]

  scaling_config {
    desired_size = 1
    max_size     = 2
    min_size     = 1
  }

  instance_types = ["t3.large"]

  launch_template {
    id      = aws_launch_template.foo.id
    version = 1
  }
  update_config {
    max_unavailable = 1
  }
}



