# Include root configuration
include "root" {
  path = find_in_parent_folders()
}

# Reference the VPC module
terraform {
  source = "../../../modules//vpc"
}

# Module-specific inputs
inputs = {
  vpc_name            = "dev-vpc"
  vpc_cidr            = "10.0.0.0/16"
  availability_zones  = ["us-east-1a", "us-east-1b"]
  public_subnets      = ["10.0.1.0/24", "10.0.2.0/24"]
  private_subnets     = ["10.0.10.0/24", "10.0.20.0/24"]
  enable_nat_gateway  = true
  single_nat_gateway  = true
}
