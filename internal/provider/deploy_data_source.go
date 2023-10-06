package provider

import (
	"context"
	"fmt"

	"terraform-provider-deploy/client"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &deployDataSource{}
	_ datasource.DataSourceWithConfigure = &deployDataSource{}
)

func NewDeployDataSource() datasource.DataSource {
	return &deployDataSource{}
}

type deployDataSource struct {
	client         client.ClientWithResponsesInterface
	organizationID uuid.UUID
}

func (d *deployDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

// Schema defines the schema for the data source.
func (d *deployDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

// deployDataSourceModel maps the data source schema data.
type deployDataSourceModel struct {
	Name types.String `tfsdk:"name"`
}

// Read refreshes the Terraform state with the latest data.
func (d *deployDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state deployDataSourceModel

	org, err := d.client.GetOrganizationWithResponse(ctx, d.organizationID)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Read Organization %s", d.organizationID),
			err.Error(),
		)
		return
	}
	if org.StatusCode() != 200 {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Read Organization %s", d.organizationID),
			org.Status(),
		)
		return
	}

	// Map response body to model
	state.Name = types.StringValue(org.JSON200.Name)

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Configure adds the provider configured client to the data source.
func (d *deployDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*deployProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = providerData.client
	d.organizationID = providerData.organizationID
}
