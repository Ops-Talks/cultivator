# cultivator:tags=prod-critical,infra
terraform {
  source = "git::https://github.com/example/terraform-modules.git//auth"
}

include "root" {
  path = find_in_parent_folders()
}

inputs = {
  environment = "prod"
  service_name = "auth-13"
}
