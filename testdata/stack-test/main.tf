provider "aws" {
  region = "eu-west-1"
}

provider "aws" {
  alias = "paris"
  region = "eu-west-3"
}

module "ec2" {
  source = "./module-ec2"
  disk_size = 123
}

module "rds" {
  source = "./module-rds"
  providers = {
    aws = aws.paris
  }
  multi_az = true
}

resource "aws_instance" "example" {
  provider = aws.paris
  ami = "some-ami"
  instance_type = "t2.micro"
}
