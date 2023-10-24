package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAssets_SingleFile_WithTarget(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "deno_assets" "test" {
						path = "./testdata/single-file"
						pattern = "*.ts"
						target = "foo"
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.deno_assets.test", "path", "./testdata/single-file"),
					resource.TestCheckResourceAttr("data.deno_assets.test", "pattern", "*.ts"),
					resource.TestCheckResourceAttr("data.deno_assets.test", "target", "foo"),
					resource.TestCheckResourceAttr("data.deno_assets.test", "output.%", "1"),
					resource.TestCheckResourceAttr("data.deno_assets.test", "output.foo/main.ts.kind", "file"),
					resource.TestCheckResourceAttr("data.deno_assets.test", "output.foo/main.ts.local_file_path", "testdata/single-file/main.ts"),
				),
			},
		},
	})
}

func TestAccAssets_SingleFile_WithoutTarget(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "deno_assets" "test" {
						path = "./testdata/single-file"
						pattern = "*.ts"
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.deno_assets.test", "path", "./testdata/single-file"),
					resource.TestCheckResourceAttr("data.deno_assets.test", "pattern", "*.ts"),
					resource.TestCheckNoResourceAttr("data.deno_assets.test", "target"),
					resource.TestCheckResourceAttr("data.deno_assets.test", "output.%", "1"),
					resource.TestCheckResourceAttr("data.deno_assets.test", "output.main.ts.kind", "file"),
					resource.TestCheckResourceAttr("data.deno_assets.test", "output.main.ts.local_file_path", "testdata/single-file/main.ts"),
				),
			},
		},
	})
}

func TestAccAssets_MultiFile_WithTarget(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "deno_assets" "test" {
						path = "./testdata/multi-file"
						pattern = "**/*"
						target = "foo"
					}
					`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.deno_assets.test", "path", "./testdata/multi-file"),
					resource.TestCheckResourceAttr("data.deno_assets.test", "pattern", "**/*"),
					resource.TestCheckResourceAttr("data.deno_assets.test", "target", "foo"),
					resource.TestCheckResourceAttr("data.deno_assets.test", "output.%", "3"),
					resource.TestCheckResourceAttr("data.deno_assets.test", "output.foo/main.ts.kind", "file"),
					resource.TestCheckResourceAttr("data.deno_assets.test", "output.foo/main.ts.local_file_path", "testdata/multi-file/main.ts"),
					resource.TestCheckResourceAttr("data.deno_assets.test", "output.foo/operands.json.kind", "file"),
					resource.TestCheckResourceAttr("data.deno_assets.test", "output.foo/operands.json.local_file_path", "testdata/multi-file/operands.json"),
					resource.TestCheckResourceAttr("data.deno_assets.test", "output.foo/util/calc.ts.kind", "file"),
					resource.TestCheckResourceAttr("data.deno_assets.test", "output.foo/util/calc.ts.local_file_path", "testdata/multi-file/util/calc.ts"),
				),
			},
		},
	})
}

func TestAccAssets_MultiFile_WithoutTarget(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "deno_assets" "test" {
						path = "./testdata/multi-file"
						pattern = "**/*"
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.deno_assets.test", "path", "./testdata/multi-file"),
					resource.TestCheckResourceAttr("data.deno_assets.test", "pattern", "**/*"),
					resource.TestCheckNoResourceAttr("data.deno_assets.test", "target"),
					resource.TestCheckResourceAttr("data.deno_assets.test", "output.%", "3"),
					resource.TestCheckResourceAttr("data.deno_assets.test", "output.main.ts.kind", "file"),
					resource.TestCheckResourceAttr("data.deno_assets.test", "output.main.ts.local_file_path", "testdata/multi-file/main.ts"),
					resource.TestCheckResourceAttr("data.deno_assets.test", "output.operands.json.kind", "file"),
					resource.TestCheckResourceAttr("data.deno_assets.test", "output.operands.json.local_file_path", "testdata/multi-file/operands.json"),
					resource.TestCheckResourceAttr("data.deno_assets.test", "output.util/calc.ts.kind", "file"),
					resource.TestCheckResourceAttr("data.deno_assets.test", "output.util/calc.ts.local_file_path", "testdata/multi-file/util/calc.ts"),
				),
			},
		},
	})
}

func TestAccAssets_Symlink_WithTarget(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "deno_assets" "test" {
						path = "./testdata/symlink"
						pattern = "**/*.{js,ts}"
						target = "foo"
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.deno_assets.test", "path", "./testdata/symlink"),
					resource.TestCheckResourceAttr("data.deno_assets.test", "pattern", "**/*.{js,ts}"),
					resource.TestCheckResourceAttr("data.deno_assets.test", "target", "foo"),
					resource.TestCheckResourceAttr("data.deno_assets.test", "output.%", "3"),
					resource.TestCheckResourceAttr("data.deno_assets.test", "output.foo/main.ts.kind", "file"),
					resource.TestCheckResourceAttr("data.deno_assets.test", "output.foo/main.ts.local_file_path", "testdata/symlink/main.ts"),
					resource.TestCheckResourceAttr("data.deno_assets.test", "output.foo/calc.js.kind", "file"),
					resource.TestCheckResourceAttr("data.deno_assets.test", "output.foo/calc.js.local_file_path", "testdata/symlink/calc.js"),
					resource.TestCheckResourceAttr("data.deno_assets.test", "output.foo/symlink.js.kind", "symlink"),
					resource.TestCheckResourceAttr("data.deno_assets.test", "output.foo/symlink.js.local_file_path", "testdata/symlink/symlink.js"),
					resource.TestCheckResourceAttr("data.deno_assets.test", "output.foo/symlink.js.runtime_target_path", "foo/calc.js"),
				),
			},
		},
	})
}

func TestAccAssets_Symlink_WithoutTarget(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					data "deno_assets" "test" {
						path = "./testdata/symlink"
						pattern = "**/*.{js,ts}"
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.deno_assets.test", "path", "./testdata/symlink"),
					resource.TestCheckResourceAttr("data.deno_assets.test", "pattern", "**/*.{js,ts}"),
					resource.TestCheckNoResourceAttr("data.deno_assets.test", "target"),
					resource.TestCheckResourceAttr("data.deno_assets.test", "output.%", "3"),
					resource.TestCheckResourceAttr("data.deno_assets.test", "output.main.ts.kind", "file"),
					resource.TestCheckResourceAttr("data.deno_assets.test", "output.main.ts.local_file_path", "testdata/symlink/main.ts"),
					resource.TestCheckResourceAttr("data.deno_assets.test", "output.calc.js.kind", "file"),
					resource.TestCheckResourceAttr("data.deno_assets.test", "output.calc.js.local_file_path", "testdata/symlink/calc.js"),
					resource.TestCheckResourceAttr("data.deno_assets.test", "output.symlink.js.kind", "symlink"),
					resource.TestCheckResourceAttr("data.deno_assets.test", "output.symlink.js.local_file_path", "testdata/symlink/symlink.js"),
					resource.TestCheckResourceAttr("data.deno_assets.test", "output.symlink.js.runtime_target_path", "calc.js"),
				),
			},
		},
	})
}
