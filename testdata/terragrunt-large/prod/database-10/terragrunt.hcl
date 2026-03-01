# cultivator:tags=infra,prod-critical,app
terraform {
  source = "git::https://github.com/example/terraform-modules.git//database"
}

include "root" {
  path = find_in_parent_folders()
}

inputs = {
  environment = "prod"
  service_name = "database-10"
}
