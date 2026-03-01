# cultivator:tags=db,app,prod-critical
terraform {
  source = "git::https://github.com/example/terraform-modules.git//api"
}

include "root" {
  path = find_in_parent_folders()
}

inputs = {
  environment = "staging"
  service_name = "api-0"
}
