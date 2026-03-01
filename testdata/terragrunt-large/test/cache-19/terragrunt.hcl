# cultivator:tags=cache,infra
terraform {
  source = "git::https://github.com/example/terraform-modules.git//cache"
}

include "root" {
  path = find_in_parent_folders()
}

inputs = {
  environment = "test"
  service_name = "cache-19"
}
