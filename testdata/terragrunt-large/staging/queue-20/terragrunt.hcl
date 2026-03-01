# cultivator:tags=db,prod-critical,infra
terraform {
  source = "git::https://github.com/example/terraform-modules.git//queue"
}

include "root" {
  path = find_in_parent_folders()
}

inputs = {
  environment = "staging"
  service_name = "queue-20"
}
