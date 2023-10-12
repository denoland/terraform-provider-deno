package provider_test

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDeployment(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccDeploymentDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "deno_project" "test" {}

					data "deno_assets" "test" {
						glob = "testdata/TestAccDeployment/main.ts"
					}

					resource "deno_deployment" "test" {
						project_id = deno_project.test.id
						entry_point_url = "testdata/TestAccDeployment/main.ts"
						compiler_options = {}
						assets = data.deno_assets.test.output
						env_vars = {}
					}
				`,
				Check: resource.ComposeTestCheckFunc(testAccCheckDeploymentDomains(t, "deno_deployment.test", "Hello world")),
			},
		},
	})
}

func testAccCheckDeploymentDomains(t *testing.T, resourceName, expectedResponse string) resource.TestCheckFunc {
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

			if string(body) != expectedResponse {
				return fmt.Errorf("the response body is expected %s, but got %s (domain = %s)", expectedResponse, string(body), domain)
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
