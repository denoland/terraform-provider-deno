---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "deno_deployment Resource - terraform-provider-deno"
subcategory: ""
description: |-
  A resource for a Deno Deploy deployment.
  A deployment belongs to a project, is an immutable, invokable snapshot of the project's assets, can be assigned a custom domain.
---

# deno_deployment (Resource)

A resource for a Deno Deploy deployment.

A deployment belongs to a project, is an immutable, invokable snapshot of the project's assets, can be assigned a custom domain.

## Example Usage

```terraform
# Assuming the following directory structure (fresh project):
# .
# ├── README.md
# ├── components
# │   └── Button.tsx
# ├── deno.json
# ├── dev.ts
# ├── fresh.config.ts
# ├── fresh.gen.ts
# ├── islands
# │   └── Counter.tsx
# ├── main.ts
# ├── routes
# │   ├── _404.tsx
# │   ├── _app.tsx
# │   ├── api
# │   │   └── joke.ts
# │   ├── greet
# │   │   └── [name].tsx
# │   └── index.tsx
# ├── static
# │   ├── favicon.ico
# │   ├── logo.svg
# │   └── styles.css
# └── terraform
#     └── main.tf

resource "deno_project" "my_project" {}

data "deno_assets" "my_assets" {
  # The path to the directory that terraform will look for assets in.
  path = ".."
  # The glob pattern that terraform will use to retrieve assets.
  pattern = "**/*.{ts,tsx,json,ico,svg,css}"
}

resource "deno_deployment" "example1" {
  # Project ID that the created deployment belongs to.
  project_id = deno_project.myproject.id
  # File path for the deployments' entry point.
  entry_point_url = "main.ts"
  # Compiler options; this can be omitted, in which case the values from 
  # `deno.json` will be used.
  compiler_options = {
    jsx               = "react-jsx"
    jsx_import_source = "preact"
  }
  assets = data.deno_assets.my_assets.output
  env_vars = {
    FOO = "42"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `assets` (Attributes Map) The entities that compose the deployment. A key represents a path to the entity. (see [below for nested schema](#nestedatt--assets))
- `entry_point_url` (String) The path to the file that will be executed when the deployment is invoked.
- `env_vars` (Map of String) The environment variables to be set in the runtime environment of the deployment.
- `project_id` (String) The project ID that this deployment belongs to.

### Optional

- `compiler_options` (Attributes) Compiler options to be used when building the deployment. If this is omitted and a deno config file (`deno.json` or `deno.jsonc`) is found in the assets, the value in the config file will be used. (see [below for nested schema](#nestedatt--compiler_options))
- `import_map_url` (String) The path to the import map file. If this is omitted and a deno config file (`deno.json` or `deno.jsonc`) is found in the assets, the value in the config file will be used.
- `lock_file_url` (String) The path to the lock file. If this is omitted and a deno config file (`deno.json` or `deno.jsonc`) is found in the assets, the value in the config will be used.
- `timeouts` (Attributes) (see [below for nested schema](#nestedatt--timeouts))

### Read-Only

- `created_at` (String) The time the deployment was created, formmatting in [RFC3339](https://datatracker.ietf.org/doc/html/rfc3339).
- `deployment_id` (String) The ID of the deployment.
- `domains` (Set of String) The domain(s) that can be used to access the deployment.
- `status` (String) The status of the deployment, indicating whether the deployment succeeded or not. It can be "failed", "pending", or "success"
- `updated_at` (String) The time the deployment was last updated, formmatting in [RFC3339](https://datatracker.ietf.org/doc/html/rfc3339).
- `uploaded_assets` (Attributes Map) The assets that have been uploaded in previous deployments, keyed with hash of the content. This is inteneded to be used to avoid uploading the same assets multiple times. (see [below for nested schema](#nestedatt--uploaded_assets))

<a id="nestedatt--assets"></a>
### Nested Schema for `assets`

Required:

- `kind` (String) The kind of entity: "file" or "symlink".

Optional:

- `content` (String) The inlined content of the asset. This is valid only for `file` asset. If both `content` and `content_source_path` are specified, it will error out.
- `content_source_path` (String) The file path of the asset in the local filesystem.
- `encoding` (String) The encoding of the inlined content. This takes effect only when `content` is present. Possible values are `utf-8` and `base64`. If omitted, the content will be interpreted as `utf-8`.
- `git_sha1` (String) The git SHA1 of the asset. It is only available for `file` asset.
- `target` (String) The target file path of the symlink in the the runtime virtual filesystem. It is only available for `symlink` asset.


<a id="nestedatt--compiler_options"></a>
### Nested Schema for `compiler_options`

Optional:

- `jsx` (String)
- `jsx_factory` (String)
- `jsx_fragment_factory` (String)
- `jsx_import_source` (String)


<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).


<a id="nestedatt--uploaded_assets"></a>
### Nested Schema for `uploaded_assets`

Read-Only:

- `git_sha1` (String)
- `path` (String)
- `updated_at` (String)
