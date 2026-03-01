# cultivator:tags=cache,infra,prod-critical
terraform {
  source = "git::https://github.com/example/terraform-modules.git//queue"
}

include "root" {
  path = find_in_parent_folders()
}

inputs = {
  environment = "dev"
  service_name = "queue-4"
}
