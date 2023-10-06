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

resource "deno_project" "sample_project" {
  name = "yusuket-3"
}

data "deno_assets" "my_assets" {
  assets_glob = "src/**/*.{ts,txt,png}"
}

resource "deno_deployment" "sample_deployment2" {
  project_id      = deno_project.sample_project.id
  entry_point_url = "src/main.ts"
  compiler_options = {
    jsx               = "react-jsx"
    jsx_import_source = "preact"
  }
  assets = data.deno_assets.my_assets.assets_metadata
  env_vars = {
    FOO = "9"
  }
}

output "deployment_id" {
  value = deno_deployment.sample_deployment2.deployment_id
}

output "deployment_urls" {
  value = deno_deployment.sample_deployment2.domains
}
