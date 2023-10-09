package provider

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"terraform-provider-deno/client"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &deployProvider{}
)

const (
	DEFAULT_API_HOST = "https://api.deno.com/v1"
)

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &deployProvider{
			version: version,
		}
	}
}

// deployProvider is the provider implementation.
type deployProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// deployProviderData is the provider-defined data that is intended to pass to
// data sources and resoures as ProviderData.
type deployProviderData struct {
	client         client.ClientWithResponsesInterface
	organizationID uuid.UUID
}

// Metadata returns the provider type name.
func (p *deployProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "deno"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *deployProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `
The terraform provider for Deno Deploy offering management on Deno projects, custom domains, and deployments.
		`,
		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Access token. May be set by the DENO_DEPLOY_TOKEN environment variable. Tokens are created here: https://dash.deno.com/account#access-tokens.",
			},
			"organization_id": schema.StringAttribute{
				Optional:    true,
				Description: "Deploy organization id. May be set by the DENO_DEPLOY_ORGANIZATION_ID environment variable. The organization id is visible in the url of the organization's project list - https://dash.deno.com/orgs/<organization_id>",
			},
			"host": schema.StringAttribute{
				Optional:    true,
				Description: "URI for the Deno API. For normal use cases this value doesn't need to be set, in which case it defaults to https://api.deno.com/v1. May be set by the DENO_API_HOST environment variable.",
			},
		},
	}
}

// deployProviderModel maps provider schema data to a Go type.
type deployProviderModel struct {
	Host           types.String `tfsdk:"host"`
	Token          types.String `tfsdk:"token"`
	OrganizationID types.String `tfsdk:"organization_id"`
}

// Configure prepares a Deploy API client for data sources and resources.
func (p *deployProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring the Deno Deploy API client")

	// TODO(magurotuna): This env var name is aligned with deployctl, but
	// there is inconsistency between host and token on whether prefixed with
	// `DENO_` or not. Maybe we want to prefix host with `DENO_`?
	host := os.Getenv("DEPLOY_API_HOST")
	token := os.Getenv("DENO_DEPLOY_TOKEN")
	rawOrganizationID := os.Getenv("DENO_DEPLOY_ORGANIZATION_ID")

	// Retrieve provider data from configuration
	var config deployProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Values set in terraform configuration take precedence over env vars
	if config.Host.ValueString() != "" {
		host = config.Host.ValueString()
	}

	if config.Token.ValueString() != "" {
		token = config.Token.ValueString()
	}

	if config.OrganizationID.ValueString() != "" {
		rawOrganizationID = config.OrganizationID.ValueString()
	}

	// If host is still empty, set it to the default value
	if host == "" {
		host = DEFAULT_API_HOST
	}

	// If token or organization ID is empty, return an error
	if token == "" {
		resp.Diagnostics.AddError(
			"Missing Deno Deploy API Token",
			"The provider cannot create the Deno Deploy API client as there is a missing or empty value for the Deno Deploy API token. Set the value statically in the configuration, or use the DENO_DEPLOY_TOKEN environment variable.",
		)
	}
	if rawOrganizationID == "" {
		resp.Diagnostics.AddError(
			"Missing Deno Deploy Organization ID",
			"Organization ID needs to be given in order for the provider to interact with the Deno Deploy API. Set the value statically in the configuration, or use the DENO_DEPLOY_ORGANIZATION_ID environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	organizationID, err := uuid.Parse(rawOrganizationID)

	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("organization_id"),
			"Invalid Deno Deploy Organization ID",
			"Valid Organization ID needs to be given in order for the provider to interact with the Deno Deploy API.",
		)
	}

	ctx = tflog.SetField(ctx, "deno_deploy_host", host)
	ctx = tflog.SetField(ctx, "deno_deploy_token", token)
	ctx = tflog.SetField(ctx, "deno_deploy_organization_id", organizationID)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "deno_deploy_token")

	tflog.Debug(ctx, "Creating Deno Deploy API client")

	// Create a new Deno Deploy client using the configuration values
	addAuth := func(ctx context.Context, req *http.Request) error {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
		return nil
	}
	client, err := client.NewClientWithResponses(host, client.WithRequestEditorFn(addAuth))
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create HashiCups API Client",
			"An unexpected error occurred when creating the HashiCups API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"HashiCups Client Error: "+err.Error(),
		)
		return
	}

	data := &deployProviderData{
		client:         client,
		organizationID: organizationID,
	}

	// Make the Deno Deploy client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = data
	resp.ResourceData = data

	tflog.Info(ctx, "Configured Deno Deploy client", map[string]any{"success": true})
}

// DataSources defines the data sources implemented in the provider.
func (p *deployProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDeployDataSource,
		NewAssetsResource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *deployProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewProjectResource,
		NewDomainResource,
		NewDomainVerificationResource,
		NewCertificateProvisioningResource,
		NewDeploymentResource,
	}
}
