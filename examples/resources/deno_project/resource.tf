resource "deno_project" "myproject" {
  # The name of the project.
  # It must follow some rules; refer to the documentation for details.
  name = "my-project"
}

# Name is optional. If omitted, a random name will be generated.
resource "deno_project" "random_name" {}
