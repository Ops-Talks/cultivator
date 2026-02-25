# Include root configuration
include "root" {
  path = find_in_parent_folders()
}

# Dependency on VPC
dependency "vpc" {
  config_path = "../vpc"
  
  mock_outputs = {
    vpc_id             = "vpc-mock"
    private_subnet_ids = ["subnet-mock-1", "subnet-mock-2"]
  }
}

# Reference the database module
terraform {
  source = "../../../modules//database"
}

# Module-specific inputs
inputs = {
  identifier           = "dev-db"
  engine               = "postgres"
  engine_version       = "15.4"
  instance_class       = "db.t3.micro"
  allocated_storage    = 20
  db_name              = "myapp"
  username             = "dbadmin"
  
  # Use outputs from VPC dependency
  vpc_id               = dependency.vpc.outputs.vpc_id
  subnet_ids           = dependency.vpc.outputs.private_subnet_ids
  
  backup_retention_period = 7
  skip_final_snapshot     = true
}
