// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"terraform-provider-deno/client"
	"terraform-provider-deno/internal/provider"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"deno": providerserver.NewProtocol6WithError(provider.New("test")()),
}

var apiClient client.ClientWithResponsesInterface

func getAPIClient(t *testing.T) client.ClientWithResponsesInterface {
	if apiClient == nil {
		token := os.Getenv("DENO_DEPLOY_TOKEN")
		addAuth := func(ctx context.Context, req *http.Request) error {
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
			return nil
		}
		c, err := client.NewClientWithResponses(os.Getenv("DEPLOY_API_HOST"), client.WithRequestEditorFn(addAuth))
		if err != nil {
			t.Fatalf("failed to create Deno Deploy API client: %s", err)
		}
		apiClient = c
	}

	return apiClient
}

func testAccPreCheck(t *testing.T) {
	ensureEnvVarExist(t, "DENO_DEPLOY_TOKEN")
	ensureEnvVarExist(t, "DEPLOY_API_HOST")
	ensureEnvVarExist(t, "DENO_DEPLOY_ORGANIZATION_ID")
}

func ensureEnvVarExist(t *testing.T, name string) {
	if os.Getenv(name) == "" {
		t.Fatalf("missing environment variable: %s", name)
	}
}
