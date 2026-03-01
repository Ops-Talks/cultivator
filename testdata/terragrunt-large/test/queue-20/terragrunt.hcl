# cultivator:tags=cache,prod-critical
terraform {
  source = "git::https://github.com/example/terraform-modules.git//queue"
}

include "root" {
  path = find_in_parent_folders()
}

inputs = {
  environment = "test"
  service_name = "queue-20"
}
