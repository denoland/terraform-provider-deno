# This data source is intended to be used with `deno_deployment` resource.
# For full example, see the doc of `deno_deployment`.

data "deno_assets" "my_assets" {
  path    = "src"
  pattern = "**/*.{ts,txt,png}"
}
