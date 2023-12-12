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

func TestAccDomain(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccDomainDestroy(t),
		Steps: []resource.TestStep{
			// Create a custom domain
			{
				Config: genConfigWithSubdomain(uuid.NewString()),
				Check:  resource.ComposeTestCheckFunc(testAccDomainExists(t, "deno_domain.test")),
			},
			// Update the domain (internally delete and create happen instead of
			// in-place update)
			{
				Config: genConfigWithSubdomain(uuid.NewString()),
				Check:  resource.ComposeTestCheckFunc(testAccDomainExists(t, "deno_domain.test")),
			},
		},
	})
}

func genConfigWithSubdomain(subdomain string) string {
	return fmt.Sprintf(`
		resource "deno_domain" "test" {
			domain = "%s.deno-staging.com"
		}

		resource "terraform_data" "a" {
			input = {
				type    = "A"
				name    = deno_domain.test.dns_record_a.name
				content = deno_domain.test.dns_record_a.content
			}
		}

		resource "terraform_data" "aaaa" {
			input = {
				type    = "AAAA"
				name    = deno_domain.test.dns_record_aaaa.name
				content = deno_domain.test.dns_record_aaaa.content
			}
		}

		resource "terraform_data" "cname" {
			input = {
				type    = "CNAME"
				name    = deno_domain.test.dns_record_cname.name
				content = deno_domain.test.dns_record_cname.content
			}
		}
	`, subdomain)
}

func testAccDomainExists(t *testing.T, resourceName string) resource.TestCheckFunc {
	c := getAPIClient(t)

	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found", resourceName)
		}

		rawDomainID, ok := rs.Primary.Attributes["id"]
		if !ok {
			return fmt.Errorf("deno_domain resource is missing id attribute")
		}
		domainID, err := uuid.Parse(rawDomainID)
		if err != nil {
			return fmt.Errorf("failed to parse domain id: %s", err)
		}

		token, ok := rs.Primary.Attributes["token"]
		if !ok {
			return fmt.Errorf("deno_domain resource is missing token attribute")
		}

		resp, err := c.GetDomainWithResponse(context.Background(), domainID)
		if err != nil {
			return fmt.Errorf("failed to get domain %s: %s", rawDomainID, err)
		}
		if client.RespIsError(resp) {
			return fmt.Errorf("domain %s does not exist: %s", rawDomainID, client.APIErrorDetail(resp.HTTPResponse, resp.Body))
		}
		if resp.JSON200.Token != token {
			return fmt.Errorf("domain %s has token %s, expected %s", rawDomainID, resp.JSON200.Token, token)
		}

		return nil
	}
}

func testAccDomainDestroy(t *testing.T) resource.TestCheckFunc {
	client := getAPIClient(t)

	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "deno_domain" {
				continue
			}
			rawDomainID, ok := rs.Primary.Attributes["id"]
			if !ok {
				return fmt.Errorf("deno_domain resource is missing id attribute")
			}
			domainID, err := uuid.Parse(rawDomainID)
			if err != nil {
				return fmt.Errorf("failed to parse domain id: %s", err)
			}
			resp, err := client.GetDomainWithResponse(context.Background(), domainID)
			if err != nil {
				return fmt.Errorf("failed to get domain %s: %s", rawDomainID, err)
			}
			if resp.JSON404 == nil {
				return fmt.Errorf("domain %s still exists", rawDomainID)
			}
		}

		return nil
	}
}
