# Root Terragrunt configuration
# This file is included by all child terragrunt.hcl files

locals {
  # Common variables
  aws_region = "us-east-1"
  
  # Parse environment and module from path
  path_parts = split("/", path_relative_to_include())
  environment = length(local.path_parts) > 1 ? local.path_parts[1] : "unknown"
  module_name = length(local.path_parts) > 2 ? local.path_parts[2] : "unknown"
  
  # Common tags
  common_tags = {
    ManagedBy   = "Terragrunt"
    Environment = local.environment
    Module      = local.module_name
  }
}

# Remote state configuration
remote_state {
  backend = "s3"
  config = {
    bucket         = "my-terraform-state-${local.environment}"
    key            = "${path_relative_to_include()}/terraform.tfstate"
    region         = local.aws_region
    encrypt        = true
    dynamodb_table = "terraform-locks"
  }
}

# Generate provider configuration
generate "provider" {
  path      = "provider.tf"
  if_exists = "overwrite_terragrunt"
  contents  = <<EOF
provider "aws" {
  region = "${local.aws_region}"
  
  default_tags {
    tags = ${jsonencode(local.common_tags)}
  }
}
EOF
}

# Terragrunt configuration
terraform {
  extra_arguments "common_vars" {
    commands = get_terraform_commands_that_need_vars()
  }
}
