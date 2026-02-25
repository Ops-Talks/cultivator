# Include root configuration
include "root" {
  path = find_in_parent_folders()
}

# Dependencies
dependency "vpc" {
  config_path = "../vpc"
  
  mock_outputs = {
    vpc_id             = "vpc-mock"
    private_subnet_ids = ["subnet-mock-1", "subnet-mock-2"]
  }
}

dependency "database" {
  config_path = "../database"
  
  mock_outputs = {
    endpoint = "db-mock.rds.amazonaws.com"
    port     = 5432
  }
}

# Reference the app module
terraform {
  source = "../../../modules//app"
}

# Module-specific inputs
inputs = {
  app_name        = "myapp"
  environment     = "dev"
  instance_type   = "t3.small"
  min_size        = 1
  max_size        = 2
  desired_capacity = 1
  
  # Use outputs from dependencies
  vpc_id          = dependency.vpc.outputs.vpc_id
  subnet_ids      = dependency.vpc.outputs.private_subnet_ids
  
  # Database connection
  db_endpoint     = dependency.database.outputs.endpoint
  db_port         = dependency.database.outputs.port
  db_name         = "myapp"
}
