# This resource is intended to be used with other resources to get the custom domain all set up.
# For full example, see the doc of `deno_domain`.

resource "deno_domain_certificate" "example" {
  # Domain ownership verification must be completed to perform certificate provisioning.
  depends_on = [deno_domain_verification.example]

  # The domain to provision a certificate for.
  domain_id = deno_domain.example.id
}
