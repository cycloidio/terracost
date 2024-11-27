provider "aws" {
  region = "eu-west-1"
}

module "ec2" {
  source = "./module-ec2"
}
