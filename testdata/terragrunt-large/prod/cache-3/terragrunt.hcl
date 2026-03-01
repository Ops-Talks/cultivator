# cultivator:tags=db,infra,cache
terraform {
  source = "git::https://github.com/example/terraform-modules.git//cache"
}

include "root" {
  path = find_in_parent_folders()
}

inputs = {
  environment = "prod"
  service_name = "cache-3"
}
