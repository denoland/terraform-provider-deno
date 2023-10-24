package provider_test

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDeployment_SingleFile(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccDeploymentDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "deno_project" "test" {}

					data "deno_assets" "test" {
						path = "./testdata/single-file"
						pattern = "main.ts"
					}

					resource "deno_deployment" "test" {
						project_id = deno_project.test.id
						entry_point_url = "main.ts"
						compiler_options = {}
						assets = data.deno_assets.test.output
						env_vars = {}
					}
				`,
				Check: resource.ComposeTestCheckFunc(testAccCheckDeploymentDomains(t, "deno_deployment.test", []responseTest{
					{
						path:     "/",
						expected: []byte("Hello world"),
					},
				})),
			},
		},
	})
}

func TestAccDeployment_SingleFileWithoutCompilerOptions(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccDeploymentDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "deno_project" "test" {}

					data "deno_assets" "test" {
						path = "testdata/single-file"
						pattern = "main.ts"
					}

					resource "deno_deployment" "test" {
						project_id = deno_project.test.id
						entry_point_url = "main.ts"
						assets = data.deno_assets.test.output
						env_vars = {}
					}
				`,
				Check: resource.ComposeTestCheckFunc(testAccCheckDeploymentDomains(t, "deno_deployment.test", []responseTest{
					{
						path:     "/",
						expected: []byte("Hello world"),
					},
				})),
			},
		},
	})
}

func TestAccDeployment_MultiFile(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccDeploymentDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "deno_project" "test" {}

					data "deno_assets" "test" {
						path = "testdata/multi-file"
						pattern = "**/*.{ts,json}"
					}

					resource "deno_deployment" "test" {
						project_id = deno_project.test.id
						entry_point_url = "main.ts"
						compiler_options = {}
						assets = data.deno_assets.test.output
						env_vars = {}
					}
				`,
				Check: resource.ComposeTestCheckFunc(testAccCheckDeploymentDomains(t, "deno_deployment.test", []responseTest{
					{
						path:     "/",
						expected: []byte("sum: 42"),
					},
				})),
			},
		},
	})
}

func TestAccDeployment_Symlink(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccDeploymentDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "deno_project" "test" {}

					data "deno_assets" "test" {
						path = "testdata/symlink"
						pattern = "**/*.{ts,js}"
					}

					resource "deno_deployment" "test" {
						project_id = deno_project.test.id
						entry_point_url = "main.ts"
						compiler_options = {}
						assets = data.deno_assets.test.output
						env_vars = {}
					}
				`,
				Check: resource.ComposeTestCheckFunc(testAccCheckDeploymentDomains(t, "deno_deployment.test", []responseTest{
					{
						path:     "/",
						expected: []byte("sum: 42"),
					},
				})),
			},
		},
	})
}

func TestAccDeployment_Binary(t *testing.T) {
	expectedBinary, err := os.ReadFile("testdata/binary/computer_screen_programming.png")
	if err != nil {
		t.Fatal(err)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccDeploymentDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "deno_project" "test" {}

					data "deno_assets" "test" {
						path = "testdata/binary"
						pattern = "**/*.{ts,png}"
					}

					resource "deno_deployment" "test" {
						project_id = deno_project.test.id
						entry_point_url = "main.ts"
						compiler_options = {}
						assets = data.deno_assets.test.output
						env_vars = {}
					}
				`,
				Check: resource.ComposeTestCheckFunc(testAccCheckDeploymentDomains(t, "deno_deployment.test", []responseTest{
					{
						path:     "/",
						expected: expectedBinary,
					},
				})),
			},
		},
	})
}

func TestAccDeployment_TSX(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccDeploymentDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "deno_project" "test" {}

					data "deno_assets" "test" {
						path = "testdata/tsx"
						pattern = "main.tsx"
					}

					resource "deno_deployment" "test" {
						project_id = deno_project.test.id
						entry_point_url = "main.tsx"
						compiler_options = {}
						assets = data.deno_assets.test.output
						env_vars = {}
					}
				`,
				Check: resource.ComposeTestCheckFunc(testAccCheckDeploymentDomains(t, "deno_deployment.test", []responseTest{
					{
						path:     "/",
						expected: []byte("<h1>Hello World!</h1>"),
					},
				})),
			},
		},
	})
}

func TestAccDeployment_ImportMap(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccDeploymentDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "deno_project" "test" {}

					data "deno_assets" "test" {
						path = "testdata/import_map"
						pattern = "**/*.{ts,json}"
					}

					resource "deno_deployment" "test" {
						project_id = deno_project.test.id
						entry_point_url = "main.ts"
						import_map_url = "import_map.json"
						compiler_options = {}
						assets = data.deno_assets.test.output
						env_vars = {}
					}
				`,
				Check: resource.ComposeTestCheckFunc(testAccCheckDeploymentDomains(t, "deno_deployment.test", []responseTest{
					{
						path:     "/",
						expected: []byte("Hello World"),
					},
				})),
			},
		},
	})
}

func TestAccDeployment_LockFile(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccDeploymentDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "deno_project" "test" {}

					data "deno_assets" "test" {
						path = "testdata/lockfile"
						pattern = "*"
					}

					resource "deno_deployment" "test" {
						project_id = deno_project.test.id
						entry_point_url = "main.ts"
						lock_file_url = "deno.lock"
						compiler_options = {}
						assets = data.deno_assets.test.output
						env_vars = {}
					}
				`,
				Check: resource.ComposeTestCheckFunc(testAccCheckDeploymentDomains(t, "deno_deployment.test", []responseTest{
					{
						path: "/",
						expected: []byte(` _______
< Hello >
 -------
        \   ^__^
         \  (oo)\_______
            (__)\       )\/\
                ||----w |
                ||     ||`),
					},
				})),
			},
		},
	})
}

func TestAccDeployment_EnvVars(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccDeploymentDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "deno_project" "test" {}

					data "deno_assets" "test" {
						path = "testdata/env_var"
						pattern = "main.ts"
					}

					resource "deno_deployment" "test" {
						project_id = deno_project.test.id
						entry_point_url = "main.ts"
						compiler_options = {}
						assets = data.deno_assets.test.output
						env_vars = {
							"FOO" = "Deno"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(testAccCheckDeploymentDomains(t, "deno_deployment.test", []responseTest{
					{
						path:     "/",
						expected: []byte("Hello Deno"),
					},
				})),
			},
		},
	})
}

func TestAccDeployment_ConfigAutoDiscovery(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccDeploymentDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "deno_project" "test" {}

					data "deno_assets" "test" {
						path = "testdata/config_auto_discovery"
						pattern = "**/*"
					}

					resource "deno_deployment" "test" {
						project_id = deno_project.test.id
						entry_point_url = "main.tsx"
						compiler_options = {}
						assets = data.deno_assets.test.output
						env_vars = {}
					}
				`,
				Check: resource.ComposeTestCheckFunc(testAccCheckDeploymentDomains(t, "deno_deployment.test", []responseTest{
					{
						path:     "/",
						expected: []byte("<h1>Hello World!</h1>"),
					},
				})),
			},
		},
	})
}

func TestAccDeployment_InlineAsset_Utf8(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccDeploymentDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "deno_project" "test" {}

					resource "deno_deployment" "test" {
						project_id = deno_project.test.id
						entry_point_url = "main.ts"
						compiler_options = {}
						assets = {
							"main.ts" = {
								kind = "file"
								content = "Deno.serve(() => new Response('Hello world'))"
							}
						}
						env_vars = {}
					}
				`,
				Check: resource.ComposeTestCheckFunc(testAccCheckDeploymentDomains(t, "deno_deployment.test", []responseTest{
					{
						path:     "/",
						expected: []byte("Hello world"),
					},
				})),
			},
		},
	})
}

func TestAccDeployment_InlineAsset_Base64(t *testing.T) {
	expectedBinary, err := os.ReadFile("testdata/binary/computer_screen_programming.png")
	if err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccDeploymentDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "deno_project" "test" {}

					resource "deno_deployment" "test" {
						project_id = deno_project.test.id
						entry_point_url = "main.ts"
						compiler_options = {}
						assets = {
							"main.ts" = {
								kind = "file"
								content = <<-EOT
									Deno.serve(async () => {
										try {
											const image = await Deno.readFile('computer_screen_programming.png');
											return new Response(image);
										} catch (error) {
											return new Response(error.message);
										}
									});
								EOT
							}
							"computer_screen_programming.png" = {
								kind = "file"
								content = filebase64("testdata/binary/computer_screen_programming.png")
								encoding = "base64"
							}
						}
						env_vars = {}
					}
				`,
				Check: resource.ComposeTestCheckFunc(testAccCheckDeploymentDomains(t, "deno_deployment.test", []responseTest{
					{
						path:     "/",
						expected: expectedBinary,
					},
				})),
			},
		},
	})
}

func TestAccDeployment_InlineAsset_Merge(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccDeploymentDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "deno_project" "test" {}

					data "deno_assets" "test" {
						path = "testdata/multi-file"
						pattern = "**/*.{ts,json}"
					}

					resource "deno_deployment" "test" {
						project_id = deno_project.test.id
						entry_point_url = "main.ts"
						compiler_options = {}
						# operands.json appears twice; later one takes precedence
						assets = merge(data.deno_assets.test.output, {
							"operands.json" = {
								kind = "file"
								content = "[1, 2]"
							}
						})
						env_vars = {}
					}
				`,
				Check: resource.ComposeTestCheckFunc(testAccCheckDeploymentDomains(t, "deno_deployment.test", []responseTest{
					{
						path:     "/",
						expected: []byte("sum: 3"),
					},
				})),
			},
		},
	})
}

func TestAccDeployment_DataSourceAsset_Merge(t *testing.T) {
	expectedBinary, err := os.ReadFile("testdata/binary/computer_screen_programming.png")
	if err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccDeploymentDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "deno_project" "test" {}

					data "deno_assets" "image" {
						path = "testdata/binary"
						pattern = "*.png"
						target = "image"
					}

					data "deno_assets" "util" {
						path = "testdata/multi-file/util"
						pattern = "*.ts"
						target = "util"
					}

					resource "deno_deployment" "test" {
						project_id = deno_project.test.id
						entry_point_url = "main.ts"
						compiler_options = {}
						assets = merge(
							data.deno_assets.image.output,
							data.deno_assets.util.output,
							{
								"main.ts" = {
									kind = "file"
									content = <<-EOT
										import { add } from "./util/calc.ts";
										Deno.serve(async (req) => {
											const url = new URL(req.url);
											if (url.pathname === "/image") {
												try {
													const image = await Deno.readFile('image/computer_screen_programming.png');
													return new Response(image);
												} catch (error) {
													return new Response(error.message);
												}
											} else {
												return new Response(add(2, 4).toString());
											}
										});
									EOT
								}
							},
						)
						env_vars = {}
					}
				`,
				Check: resource.ComposeTestCheckFunc(testAccCheckDeploymentDomains(t, "deno_deployment.test", []responseTest{
					{
						path:     "/image",
						expected: expectedBinary,
					},
					{
						path:     "/",
						expected: []byte("6"),
					},
				})),
			},
		},
	})
}

type responseTest struct {
	path     string
	expected []byte
}

// nolint:unparam
func testAccCheckDeploymentDomains(t *testing.T, resourceName string, responseTests []responseTest) resource.TestCheckFunc {
	_ = getAPIClient(t)

	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		numDomainsStr, ok := rs.Primary.Attributes["domains.#"]
		if !ok {
			return fmt.Errorf("deno_deployment resource is missing domains attribute")
		}
		numDomains, err := strconv.Atoi(numDomainsStr)
		if err != nil {
			return fmt.Errorf("failed to parse the number of domains: %s", err)
		}

		// Wait for a bit to make sure domain mapping update is propagated
		time.Sleep(3 * time.Second)

		for i := 0; i < numDomains; i++ {
			domain, ok := rs.Primary.Attributes[fmt.Sprintf("domains.%d", i)]
			if !ok {
				return fmt.Errorf("deno_deployment resource is missing domains attribute")
			}

			for _, responseTest := range responseTests {
				resp, err := http.Get(fmt.Sprintf("https://%s%s", domain, responseTest.path))
				if err != nil {
					return fmt.Errorf("failed to get the deployment (domain = %s): %s", domain, err)
				}
				defer resp.Body.Close()
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return fmt.Errorf("failed to read the response body (domain = %s): %s", domain, err)
				}

				if !bytes.Equal(body, responseTest.expected) {
					var expected string
					if utf8.Valid(responseTest.expected) {
						expected = string(responseTest.expected)
					} else {
						expected = base64.StdEncoding.EncodeToString(responseTest.expected)
					}

					var got string
					if utf8.Valid(body) {
						got = string(body)
					} else {
						got = base64.StdEncoding.EncodeToString(body)
					}

					return fmt.Errorf("the response body is expected %s, but got %s (domain = %s)", expected, got, domain)
				}
			}
		}

		return nil
	}
}

// Deployments are immutable resources; destroy check will do nothing.
func testAccDeploymentDestroy(t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return nil
	}
}
