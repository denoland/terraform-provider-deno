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
  glob = "../**/*.{ts,tsx,json,ico,svg,css}"
}

resource "deno_deployment" "example1" {
  # Project ID that the created deployment belongs to.
  project_id = deno_project.myproject.id
  # File path for the deployments' entry point.
  entry_point_url = "../main.ts"
  compiler_options = {
    jsx               = "react-jsx"
    jsx_import_source = "preact"
  }
  assets = data.deno_assets.my_assets.output
  env_vars = {
    FOO = "42"
  }
}
