# cultivator:tags=infra,cache
terraform {
  source = "git::https://github.com/example/terraform-modules.git//app"
}

include "root" {
  path = find_in_parent_folders()
}

inputs = {
  environment = "prod"
  service_name = "app-1"
}
