# cultivator:tags=prod-critical
terraform {
  source = "git::https://github.com/example/terraform-modules.git//api"
}

include "root" {
  path = find_in_parent_folders()
}

inputs = {
  environment = "dev"
  service_name = "api-0"
}
