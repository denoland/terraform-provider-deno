terraform {
  required_providers {
    deno = {
      source = "denoland/deno"
    }
  }
}

provider "deno" {
  # Put your token here.
  # If omitted, the token will be read from the environment variable `DENO_DEPLOY_TOKEN`.
  token = var.deno_deploy_api_token

  # Organization ID that this provider will interact with.
  # If omitted, the organization ID will be read from the environment variable `DENO_DEPLOY_ORGANIZATION_ID`.
  organization_id = "your_organization_id"
}
