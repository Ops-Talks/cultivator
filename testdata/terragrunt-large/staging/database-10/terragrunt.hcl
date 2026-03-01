# cultivator:tags=infra
terraform {
  source = "git::https://github.com/example/terraform-modules.git//database"
}

include "root" {
  path = find_in_parent_folders()
}

inputs = {
  environment = "staging"
  service_name = "database-10"
}
