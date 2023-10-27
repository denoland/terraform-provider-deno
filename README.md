# Deno Terraform Provider

This provider uses the Deno API to manage Deno projects, custom domains, and
deployments. A simple example is available in examples/. This is a very early
release of the provider. Additional documentation and examples are forthcoming.

There are working examples in `/examples`. To use the examples, you need a Deno
deploy organization ID and an access token. See `docs/index.md` for instructions
on finding or creating those values.

Resource and schema documentation still needs to be written. In the interim, API
documentation is available at https://apidocs.deno.com/. Resources are aligned
with the API.

If you want to build and run the provider locally, you can override where
terraform finds the provider. The override is configured in $HOME/.terraformrc.

```terraform.rc
provider_installation {

  dev_overrides {
      "registry.terraform.io/denoland/deno" = "/Users/<username>/go/bin"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
```

With this set (assuming your go binary install location is $HOME/go/bin), you
can `go install` the provider and use the locally built version vs. the registry
published version. (Note that the override doesn't seem to resolve environment
variables, so `$HOME/go/bin` will not work.)
