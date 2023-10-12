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

func TestAccDeployment(t *testing.T) {
	// Single file project
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccDeploymentDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "deno_project" "test" {}

					data "deno_assets" "test" {
						glob = "testdata/single-file/main.ts"
					}

					resource "deno_deployment" "test" {
						project_id = deno_project.test.id
						entry_point_url = "testdata/single-file/main.ts"
						compiler_options = {}
						assets = data.deno_assets.test.output
						env_vars = {}
					}
				`,
				Check: resource.ComposeTestCheckFunc(testAccCheckDeploymentDomains(t, "deno_deployment.test", []byte("Hello world"))),
			},
		},
	})

	// TODO: This isn't working now. Uncomment this test case once it's resolved.
	// Issue: https://github.com/denoland/terraform-provider-deno/issues/18
	// Single file project without compiler_options
	// resource.Test(t, resource.TestCase{
	// 	PreCheck:                 func() { testAccPreCheck(t) },
	// 	ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
	// 	CheckDestroy:             testAccDeploymentDestroy(t),
	// 	Steps: []resource.TestStep{
	// 		{
	// 			Config: `
	// 				resource "deno_project" "test" {}

	// 				data "deno_assets" "test" {
	// 					glob = "testdata/single-file/main.ts"
	// 				}

	// 				resource "deno_deployment" "test" {
	// 					project_id = deno_project.test.id
	// 					entry_point_url = "testdata/single-file/main.ts"
	// 					assets = data.deno_assets.test.output
	// 					env_vars = {}
	// 				}
	// 			`,
	// 			Check: resource.ComposeTestCheckFunc(testAccCheckDeploymentDomains(t, "deno_deployment.test", []byte("Hello world"))),
	// 		},
	// 	},
	// })

	// Mutli-file project
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccDeploymentDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "deno_project" "test" {}

					data "deno_assets" "test" {
						glob = "testdata/multi-file/**/*.{ts,json}"
					}

					resource "deno_deployment" "test" {
						project_id = deno_project.test.id
						entry_point_url = "testdata/multi-file/main.ts"
						compiler_options = {}
						assets = data.deno_assets.test.output
						env_vars = {}
					}
				`,
				Check: resource.ComposeTestCheckFunc(testAccCheckDeploymentDomains(t, "deno_deployment.test", []byte("sum: 42"))),
			},
		},
	})

	// Contains symlink
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccDeploymentDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "deno_project" "test" {}

					data "deno_assets" "test" {
						glob = "testdata/symlink/**/*.{ts,js}"
					}

					resource "deno_deployment" "test" {
						project_id = deno_project.test.id
						entry_point_url = "testdata/symlink/main.ts"
						compiler_options = {}
						assets = data.deno_assets.test.output
						env_vars = {}
					}
				`,
				Check: resource.ComposeTestCheckFunc(testAccCheckDeploymentDomains(t, "deno_deployment.test", []byte("sum: 42"))),
			},
		},
	})

	// Contains binary file (image)
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
						glob = "testdata/binary/**/*.{ts,png}"
					}

					resource "deno_deployment" "test" {
						project_id = deno_project.test.id
						entry_point_url = "testdata/binary/main.ts"
						compiler_options = {}
						assets = data.deno_assets.test.output
						env_vars = {}
					}
				`,
				Check: resource.ComposeTestCheckFunc(testAccCheckDeploymentDomains(t, "deno_deployment.test", expectedBinary)),
			},
		},
	})

	// tsx
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccDeploymentDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "deno_project" "test" {}

					data "deno_assets" "test" {
						glob = "testdata/tsx/main.tsx"
					}

					resource "deno_deployment" "test" {
						project_id = deno_project.test.id
						entry_point_url = "testdata/tsx/main.tsx"
						compiler_options = {}
						assets = data.deno_assets.test.output
						env_vars = {}
					}
				`,
				Check: resource.ComposeTestCheckFunc(testAccCheckDeploymentDomains(t, "deno_deployment.test", []byte("<h1>Hello World!</h1>"))),
			},
		},
	})
}

// nolint:unparam
func testAccCheckDeploymentDomains(t *testing.T, resourceName string, expectedResponse []byte) resource.TestCheckFunc {
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

			resp, err := http.Get(fmt.Sprintf("https://%s", domain))
			if err != nil {
				return fmt.Errorf("failed to get the deployment (domain = %s): %s", domain, err)
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to read the response body (domain = %s): %s", domain, err)
			}

			if !bytes.Equal(body, expectedResponse) {
				var expected string
				if utf8.Valid(expectedResponse) {
					expected = string(expectedResponse)
				} else {
					expected = base64.StdEncoding.EncodeToString(expectedResponse)
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

		return nil
	}
}

// Deployments are immutable resources; destroy check will do nothing.
func testAccDeploymentDestroy(t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		return nil
	}
}
