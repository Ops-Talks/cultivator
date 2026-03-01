# cultivator:tags=app,db,prod-critical
terraform {
  source = "git::https://github.com/example/terraform-modules.git//app"
}

include "root" {
  path = find_in_parent_folders()
}

inputs = {
  environment = "prod"
  app_name    = "app1"
  replicas    = 3
}
