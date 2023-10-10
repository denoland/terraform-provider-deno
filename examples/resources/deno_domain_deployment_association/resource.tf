# Assumes that the domain and deployment already exist.
#
# Note that the domain must be verified for its ownership and certificates must be ready.
# For the full example of the entire process of domain setup, see the doc of deno_domain resource.
resource "deno_domain_deployment_association" "example" {
  domain_id     = deno_domain.example_domain.id
  deployment_id = deno_deployment.example_deployment.deployment_id
}
