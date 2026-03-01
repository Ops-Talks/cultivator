# cultivator:tags=app,cache
terraform {
  source = "git::https://github.com/example/terraform-modules.git//app"
}

include "root" {
  path = find_in_parent_folders()
}

inputs = {
  environment = "dev"
  service_name = "app-17"
}
