package provider

import (
	"context"
	"fmt"
	"terraform-provider-deno/client"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &projectResource{}
	_ resource.ResourceWithConfigure   = &projectResource{}
	_ resource.ResourceWithImportState = &projectResource{}
)

// NewProjectResource is a helper function to simplify the provider implementation.
func NewProjectResource() resource.Resource {
	return &projectResource{}
}

// projectResource is the resource implementation.
type projectResource struct {
	client         client.ClientWithResponsesInterface
	organizationID uuid.UUID
}

// projectResourceModel maps the resource schema data.
type projectResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

// Metadata returns the resource type name.
func (r *projectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

// Schema defines the schema for the resource.
func (r *projectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `
A resource for a Deno Deploy project.

A project consists of a collection of deployments and belongs to an organization.
		`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The ID of the project.",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The name of the project. This must be globally unique, and must be between 3 and 26 characters, only contain a-z, 0-9 and -, must not start or end with a hyphen (-), and characters after hyphen (-) shouldn't be 8 or 12 in length. If not provided, a random name will be generated.",
			},
			"created_at": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description:         "The time the project was created, formatted in RFC3339.",
				MarkdownDescription: "The time the project was created, formatted in [RFC3339](https://datatracker.ietf.org/doc/html/rfc3339).",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The time the project was last updated, formatted in [RFC3339](https://datatracker.ietf.org/doc/html/rfc3339).",
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *projectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan projectResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request body from plan

	projName := plan.Name.ValueString()
	var projNameForAPICall *string
	if projName == "" {
		projNameForAPICall = nil
	} else {
		projNameForAPICall = &projName
	}
	proj, err := r.client.CreateProjectWithResponse(ctx, r.organizationID, client.CreateProjectJSONRequestBody{
		Name: projNameForAPICall,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Create Project %s", projName),
			err.Error(),
		)
		return
	}
	if client.RespIsError(proj) {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Create Project %s", projName),
			client.APIErrorDetail(proj.HTTPResponse, proj.Body),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(proj.JSON200.Id.String())
	plan.Name = types.StringValue(proj.JSON200.Name)
	plan.CreatedAt = types.StringValue(proj.JSON200.CreatedAt.Format(time.RFC3339))
	plan.UpdatedAt = types.StringValue(proj.JSON200.UpdatedAt.Format(time.RFC3339))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *projectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state projectResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Read Project %s", state.ID),
			fmt.Sprintf("Could not parse project ID %s: %s", state.ID, err.Error()),
		)
		return
	}

	// Get refreshed order value from Deno Deploy
	proj, err := r.client.GetProjectWithResponse(ctx, projID)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Read Project %s", state.ID),
			err.Error(),
		)
		return
	}
	if client.RespIsError(proj) {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Read Project %s", state.ID),
			client.APIErrorDetail(proj.HTTPResponse, proj.Body),
		)
		return
	}

	// Overwtite state with refreshed values
	state.ID = types.StringValue(proj.JSON200.Id.String())
	state.Name = types.StringValue(proj.JSON200.Name)
	state.CreatedAt = types.StringValue(proj.JSON200.CreatedAt.Format(time.RFC3339))
	state.UpdatedAt = types.StringValue(proj.JSON200.UpdatedAt.Format(time.RFC3339))

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *projectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan projectResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projID, err := uuid.Parse(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Update Project %s", plan.ID),
			fmt.Sprintf("Could not parse project ID %s: %s", plan.ID, err.Error()),
		)
		return
	}

	// Update existing project
	proj, err := r.client.UpdateProjectWithResponse(ctx, projID, client.UpdateProjectRequest{
		Name: plan.Name.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Update Project %s", plan.ID),
			err.Error(),
		)
		return
	}
	if client.RespIsError(proj) {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Update Project %s", plan.ID),
			client.APIErrorDetail(proj.HTTPResponse, proj.Body),
		)
		return
	}

	// Update resource state with updates values
	plan.Name = types.StringValue(proj.JSON200.Name)
	plan.UpdatedAt = types.StringValue(proj.JSON200.UpdatedAt.Format(time.RFC3339))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *projectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state projectResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Update Project %s", state.ID),
			fmt.Sprintf("Could not parse project ID %s: %s", state.ID, err.Error()),
		)
		return
	}

	// Delete existing project
	result, err := r.client.DeleteProjectWithResponse(ctx, projID)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Delete Project %s", state.ID),
			err.Error(),
		)
		return
	}
	if client.RespIsError(result) {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Delete Project %s", state.ID),
			client.APIErrorDetail(result.HTTPResponse, result.Body),
		)
		return
	}
}

// ImportState imports the existing resource into Terraform.
func (r *projectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Configure adds the provider configured client to the resource.
func (r *projectResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*deployProviderData)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.ClientWithResponses, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = providerData.client
	r.organizationID = providerData.organizationID
}
