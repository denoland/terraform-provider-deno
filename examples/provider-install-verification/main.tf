terraform {
  required_providers {
    deno = {
      source = "github.com/denoland/deploy"
    }

    cloudflare = {
      source = "github.com/cloudflare/cloudflare"
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

variable "cloudflare_api_token" {
  type      = string
  sensitive = true
}

variable "cloudflare_zone_id" {
  type = string
}

provider "deno" {
  host            = var.deno_deploy_api_host
  token           = var.deno_deploy_api_token
  organization_id = var.deno_deploy_organization_id
}

data "deno_organization" "example" {}

resource "deno_project" "example_project" {
  name = "myproject-2"
}

resource "deno_domain" "example_domain" {
  domain = "my-project2.maguro.me"
}

resource "deno_domain_verification" "example_domain_verification" {
  depends_on = [cloudflare_record.my_record_0, cloudflare_record.my_record_1, cloudflare_record.my_record_2]

  domain_id = deno_domain.example_domain.id

  timeouts = {
    create = "15m"
  }
}

resource "deno_certificate_provisioning" "example_certificate_provisioning" {
  depends_on = [deno_domain_verification.example_domain_verification]

  domain_id = deno_domain.example_domain.id
}

output "project_id" {
  value = deno_project.example_project.id
}

///////////////////////////////////////////////
// DNS record
///////////////////////////////////////////////
provider "cloudflare" {
  api_token = var.cloudflare_api_token
}

resource "cloudflare_record" "my_record_0" {
  zone_id = var.cloudflare_zone_id
  name    = deno_domain.example_domain.dns_records[0].name
  type    = upper(deno_domain.example_domain.dns_records[0].type)
  value   = deno_domain.example_domain.dns_records[0].content
  proxied = false
  ttl     = 120
}

resource "cloudflare_record" "my_record_1" {
  zone_id = var.cloudflare_zone_id
  name    = deno_domain.example_domain.dns_records[1].name
  type    = upper(deno_domain.example_domain.dns_records[1].type)
  value   = deno_domain.example_domain.dns_records[1].content
  proxied = false
  ttl     = 120
}

resource "cloudflare_record" "my_record_2" {
  zone_id = var.cloudflare_zone_id
  name    = deno_domain.example_domain.dns_records[2].name
  type    = upper(deno_domain.example_domain.dns_records[2].type)
  value   = deno_domain.example_domain.dns_records[2].content
  proxied = false
  ttl     = 120
}
