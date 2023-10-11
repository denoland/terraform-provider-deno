package provider_test

import (
	"context"
	"fmt"
	"terraform-provider-deno/client"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccProject(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccProjectDestroy(t),
		Steps: []resource.TestStep{
			{
				Config: `
					resource "deno_project" "test" {
						name = "test-project"
					}
				`,
				Check: resource.ComposeTestCheckFunc(testAccProjectExists(t, "deno_project.test")),
			},
		},
	})
}

func testAccProjectExists(t *testing.T, resourceName string) resource.TestCheckFunc {
	c := getAPIClient(t)

	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		rawProjectID, ok := rs.Primary.Attributes["id"]
		if !ok {
			return fmt.Errorf("deno_project resource is missing id attribute")
		}
		projectID, err := uuid.Parse(rawProjectID)
		if err != nil {
			return fmt.Errorf("failed to parse project id %s: %s", rawProjectID, err)
		}
		resp, err := c.GetProjectWithResponse(context.Background(), projectID)
		if err != nil {
			return fmt.Errorf("failed to get project %s: %s", rawProjectID, err)
		}
		if client.RespIsError(resp) {
			return fmt.Errorf("project %s does not exist: %s", rawProjectID, client.APIErrorDetail(resp.HTTPResponse, resp.Body))
		}

		return nil
	}
}

func testAccProjectDestroy(t *testing.T) func(*terraform.State) error {
	client := getAPIClient(t)

	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "deno_project" {
				continue
			}
			rawProjectID, ok := rs.Primary.Attributes["id"]
			if !ok {
				return fmt.Errorf("deno_project resource is missing id attribute")
			}
			projectID, err := uuid.Parse(rawProjectID)
			if err != nil {
				return fmt.Errorf("failed to parse project id: %s", err)
			}
			resp, err := client.GetProjectWithResponse(context.Background(), projectID)
			if err != nil {
				return fmt.Errorf("failed to get project: %s", err)
			}
			if resp.JSON404 == nil {
				return fmt.Errorf("project still exists: %s", rawProjectID)
			}
		}

		return nil
	}
}
