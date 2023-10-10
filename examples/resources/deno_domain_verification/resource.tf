# This resource is intended to be used with other resources to get the custom domain all set up.
# For full example, see the doc of `deno_domain`.

resource "deno_domain_verification" "example" {
  # You may want to wait for DNS records to be propagated.
  depends_on = [cloudflare_record.record_x, cloudflare_record.record_y, cloudflare_record.record_z]

  domain_id = deno_domain.example.id

  # DNS propagation may take a while; timeout period can be specified.
  timeouts = {
    create = "15m"
  }
}
