# cultivator:tags=app,db
terraform {
  source = "git::https://github.com/example/terraform-modules.git//storage"
}

include "root" {
  path = find_in_parent_folders()
}

inputs = {
  environment = "prod"
  service_name = "storage-22"
}
