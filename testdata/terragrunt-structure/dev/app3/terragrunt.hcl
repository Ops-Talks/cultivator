# cultivator:tags=app,test
terraform {
  source = "git::https://github.com/example/terraform-modules.git//app"
}

include "root" {
  path = find_in_parent_folders()
}

inputs = {
  environment = "dev"
  app_name    = "app3"
  replicas    = 1
}
