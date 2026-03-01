# cultivator:tags=db,cache
terraform {
  source = "git::https://github.com/example/terraform-modules.git//database"
}

include "root" {
  path = find_in_parent_folders()
}

inputs = {
  environment = "prod"
  db_engine   = "postgres"
  backup_retention = 30
}
