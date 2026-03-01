# cultivator:tags=cache,db,infra
terraform {
  source = "git::https://github.com/example/terraform-modules.git//cache"
}

include "root" {
  path = find_in_parent_folders()
}

inputs = {
  environment = "staging"
  service_name = "cache-11"
}
