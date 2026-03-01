# cultivator:tags=db
terraform {
  source = "git::https://github.com/example/terraform-modules.git//auth"
}

include "root" {
  path = find_in_parent_folders()
}

inputs = {
  environment = "staging"
  service_name = "auth-5"
}
