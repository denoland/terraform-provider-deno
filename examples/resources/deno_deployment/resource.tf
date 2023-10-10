# Assuming the following directory structure (fresh project):
# .
# ├── README.md
# ├── components
# │   └── Button.tsx
# ├── deno.json
# ├── dev.ts
# ├── fresh.config.ts
# ├── fresh.gen.ts
# ├── islands
# │   └── Counter.tsx
# ├── main.ts
# ├── routes
# │   ├── _404.tsx
# │   ├── _app.tsx
# │   ├── api
# │   │   └── joke.ts
# │   ├── greet
# │   │   └── [name].tsx
# │   └── index.tsx
# ├── static
# │   ├── favicon.ico
# │   ├── logo.svg
# │   └── styles.css
# └── terraform
#     └── main.tf

resource "deno_project" "my_project" {}

data "deno_assets" "my_assets" {
  assets_glob = "../**/*.{ts,tsx,json,ico,svg,css}"
}

resource "deno_deployment" "example" {
  # Project ID that the created deployment belongs to.
  project_id = deno_project.myproject.id
  # File path for the deployments' entry point.
  entry_point_url = "../main.ts"
  compiler_options = {
    jsx               = "react-jsx"
    jsx_import_source = "preact"
  }
  assets = data.deno_assets.my_assets.assets_metadata
  env_vars = {
    FOO = "42"
  }

  ###############################################
  # Custom domain association
  ###############################################
  #
  # A custom domain can be associated with the deployment (optional).
  # Note the domain must be verified for its ownership and certificates must be ready.
  # See the doc of deno_domain resource for the full example of the entire process of domain setup.

  # `depends_on` may be useful to ensure the domain is ready.
  depends_on = [deno_certificate_provisioning.example]
  domain_id  = deno_domain.example.id
}
