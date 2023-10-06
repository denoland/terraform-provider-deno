terraform {
  required_providers {
    deno = {
      source = "github.com/denoland/deploy"
    }
  }
}

variable "deno_deploy_api_host" {
  type = string
}

variable "deno_deploy_api_token" {
  type      = string
  sensitive = true
}

variable "deno_deploy_organization_id" {
  type = string
}

provider "deno" {
  host            = var.deno_deploy_api_host
  token           = var.deno_deploy_api_token
  organization_id = var.deno_deploy_organization_id
}

data "deno_assets" "my_assets" {
  assets_glob = "../../**/*.md"
}

output "assets" {
  value = data.deno_assets.my_assets.assets_metadata
}
