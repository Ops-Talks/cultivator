# cultivator:tags=db,app
terraform {
  source = "git::https://github.com/example/terraform-modules.git//logging"
}

include "root" {
  path = find_in_parent_folders()
}

inputs = {
  environment = "staging"
  service_name = "logging-23"
}
