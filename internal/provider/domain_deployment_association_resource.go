package provider

import (
	"context"
	"fmt"
	"terraform-provider-deno/client"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &DomainDeploymentAssociationResource{}
	_ resource.ResourceWithConfigure = &DomainDeploymentAssociationResource{}
)

// NewDomainDeploymentAssociationResource is a helper function to simplify the provider implementation.
func NewDomainDeploymentAssociationResource() resource.Resource {
	return &DomainDeploymentAssociationResource{}
}

// DomainDeploymentAssociationResource is the resource implementation.
type DomainDeploymentAssociationResource struct {
	client         client.ClientWithResponsesInterface
	organizationID uuid.UUID
}

// DomainDeploymentAssociationResourceModel maps the resource schema data.
type DomainDeploymentAssociationResourceModel struct {
	DomainID     types.String `tfsdk:"domain_id"`
	DeploymentID types.String `tfsdk:"deployment_id"`
}

// Metadata returns the resource type name.
func (r *DomainDeploymentAssociationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_deployment_association"
}

// Schema defines the schema for the resource.
func (r *DomainDeploymentAssociationResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `
A resource for an association between a custom domain and a deployment.

This resource associates a custom domain with a deployment, allowing for access to the deployment via the custom domain. For the whole example of setting up a custom domain, please refer to the doc of deno_domain resource.
		`,
		Attributes: map[string]schema.Attribute{
			"domain_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the domain to verify ownership of.",
			},
			"deployment_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the deployment to associate the domain with.",
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *DomainDeploymentAssociationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan DomainDeploymentAssociationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domainID, err := uuid.Parse(plan.DomainID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Associate the Domain %s with the Deployment %s", plan.DomainID, plan.DeploymentID),
			fmt.Sprintf("Could not parse domain ID %s: %s", plan.DomainID, err.Error()),
		)
		return
	}

	result, err := r.client.UpdateDomainAssociationWithResponse(ctx, domainID, client.UpdateDomainAssociationRequest{})
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Associate the Domain %s with the Deployment %s", plan.DomainID, plan.DeploymentID),
			fmt.Sprintf("update domain association API returned error: %s", err.Error()),
		)
		return
	}
	if client.RespIsError(result) {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Associate the Domain %s with the Deployment %s", plan.DomainID, plan.DeploymentID),
			client.APIErrorDetail(result.HTTPResponse, result.Body),
		)
		return
	}

	// Mark as verified
	plan.Verified = types.BoolValue(true)

	// Set state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *DomainDeploymentAssociationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state DomainDeploymentAssociationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domainID, err := uuid.Parse(state.DomainID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Read Domain Verification %s", state.DomainID),
			fmt.Sprintf("Could not parse domain ID %s: %s", state.DomainID, err.Error()),
		)
		return
	}

	result, err := r.client.VerifyDomainWithResponse(ctx, domainID)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Read Domain Verification %s", state.DomainID),
			fmt.Sprintf("veirfy API returned error: %s", err.Error()),
		)
		return
	}

	if client.RespIsError(result) {
		state.Verified = types.BoolValue(false)
	} else {
		state.Verified = types.BoolValue(true)
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *DomainDeploymentAssociationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Follow the same procedure as Create

	// Retrieve values from plan
	var plan DomainDeploymentAssociationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	const DEFAULT_TIMEOUT = 10 * time.Minute
	timeout, diags := plan.Timeouts.Create(ctx, DEFAULT_TIMEOUT)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domainID, err := uuid.Parse(plan.DomainID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Update Domain Verification %s", plan.DomainID),
			fmt.Sprintf("Could not parse domain ID %s: %s", plan.DomainID, err.Error()),
		)
		return
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

Polling:
	for {
		select {
		// timeout
		case <-ctx.Done():
			resp.Diagnostics.AddError(
				fmt.Sprintf("Unable to Update Domain Verification %s", plan.DomainID),
				fmt.Sprintf("Timed out after %s", timeout),
			)
			return
		// polling
		case <-ticker.C:
			// Call the API
			result, err := r.client.VerifyDomainWithResponse(ctx, domainID)

			if err != nil {
				resp.Diagnostics.AddError(
					fmt.Sprintf("Unable to Update Domain Verification %s", plan.DomainID),
					fmt.Sprintf("veirfy API returned error: %s", err.Error()),
				)
				return
			}
			if client.RespIsError(result) {
				continue
			}

			// Verification completed
			break Polling
		}
	}

	// Mark as verified
	plan.Verified = types.BoolValue(true)

	// Set state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *DomainDeploymentAssociationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// noop
}

// Configure adds the provider configured client to the resource.
func (r *DomainDeploymentAssociationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
