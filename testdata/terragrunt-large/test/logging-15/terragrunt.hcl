# cultivator:tags=infra
terraform {
  source = "git::https://github.com/example/terraform-modules.git//logging"
}

include "root" {
  path = find_in_parent_folders()
}

inputs = {
  environment = "test"
  service_name = "logging-15"
}
