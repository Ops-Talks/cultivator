# cultivator:tags=db,cache
terraform {
  source = "git::https://github.com/example/terraform-modules.git//storage"
}

include "root" {
  path = find_in_parent_folders()
}

inputs = {
  environment = "staging"
  service_name = "storage-6"
}
