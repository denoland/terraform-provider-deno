# This example demonstrates the whole flow of setting up a custom domain.
# Several terraform resources need to interact with each other to aciheve this.

terraform {
  required_providers {
    deno = {
      source = "denoland/deno"
    }

    # As a demonstration, we use the cloudflare provider to add DNS records.
    # You can use any provider that can add DNS records.
    cloudflare = {
      source = "cloudflare/cloudflare"
    }
  }
}

# Add a new domain to your organization.
resource "deno_domain" "example" {
  domain = "foo.example.com"
}

# Add DNS records to the nameserver.
resource "cloudflare_record" "my_record" {
  for_each = deno_domain.example.dns_records_map

  zone_id = "<put your zone ID>"
  name    = each.value.name
  type    = upper(each.key)
  value   = each.value.content
  proxied = false
  ttl     = 120
}

# Added custom domain needs to be verified for ownership.
resource "deno_domain_verification" "example" {
  depends_on = [cloudflare_record.my_record_0, cloudflare_record.my_record_1, cloudflare_record.my_record_2]

  domain_id = deno_domain.example.id

  timeouts = {
    create = "15m"
  }
}

# Provision a certificate for the domain.
# The certificate will be managed by Deno Deploy.
resource "deno_domain_certificate" "example" {
  depends_on = [deno_domain_verification.example]

  domain_id = deno_domain.example.id
}
