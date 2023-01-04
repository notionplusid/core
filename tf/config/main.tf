locals {
  ws = try(terraform.workspace, "production")

  ws_list = ["staging", "production"]

  config = yamldecode(file("${path.module}/../config.yaml"))
}

resource "null_resource" "is_ws_name_valid" {
  count = contains(local.ws_list, local.ws) == true ? 0 : file("ERROR: The WORKSPACE name must match the one in the config.yaml file")
}
