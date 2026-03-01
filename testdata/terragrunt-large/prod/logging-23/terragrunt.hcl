# cultivator:tags=db,cache
terraform {
  source = "git::https://github.com/example/terraform-modules.git//logging"
}

include "root" {
  path = find_in_parent_folders()
}

inputs = {
  environment = "prod"
  service_name = "logging-23"
}
