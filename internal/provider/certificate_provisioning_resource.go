package provider

import (
	"context"
	"fmt"
	"terraform-provider-deno/client"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &certificateProvisioningResource{}
	_ resource.ResourceWithConfigure = &certificateProvisioningResource{}
)

// NewCertificateProvisioningResource is a helper function to simplify the provider implementation.
func NewCertificateProvisioningResource() resource.Resource {
	return &certificateProvisioningResource{}
}

// certificateProvisioningResource is the resource implementation.
type certificateProvisioningResource struct {
	client         client.ClientWithResponsesInterface
	organizationID uuid.UUID
}

// certificateProvisioningResourceModel maps the resource schema data.
type certificateProvisioningResourceModel struct {
	DomainID           types.String `tfsdk:"domain_id"`
	ProvisioningStatus types.String `tfsdk:"provisioning_status"`
}

// Metadata returns the resource type name.
func (r *certificateProvisioningResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificate_provisioning"
}

// Schema defines the schema for the resource.
func (r *certificateProvisioningResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"domain_id": schema.StringAttribute{
				Required: true,
			},
			"provisioning_status": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *certificateProvisioningResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan certificateProvisioningResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domainID, err := uuid.Parse(plan.DomainID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Provision Certificates for Domain %s", plan.DomainID),
			fmt.Sprintf("Could not parse domain ID %s: %s", plan.DomainID, err.Error()),
		)
		return
	}

	// Call the API to trigger provisioning
	result, err := r.client.ProvisionDomainCertificatesWithResponse(ctx, domainID)

	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Provision Certificates for Domain %s", plan.DomainID),
			fmt.Sprintf("API returned error: %s", err.Error()),
		)
		return
	}
	if client.RespIsError(result) {
		tflog.Debug(ctx, "Provision API returned error", map[string]any{
			"400": result.JSON400,
			"401": result.JSON401,
			"404": result.JSON404,
		})
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Provision Certificates for Domain %s", plan.DomainID),
			client.APIErrorDetail(result.HTTPResponse, result.Body),
		)
		return
	}

	// Call the API to get the provisioning status
	provisioningStatus, diag := r.getCurrentProvisioningStatus(ctx, domainID)
	resp.Diagnostics.Append(diag)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set provisioning status to plan
	plan.ProvisioningStatus = types.StringValue(provisioningStatus)

	// Set state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if provisioningStatus != "success" {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Provision Certificates for Domain %s", domainID),
			fmt.Sprintf("Provisioning status is %s, expected success", provisioningStatus),
		)
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *certificateProvisioningResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state certificateProvisioningResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domainID, err := uuid.Parse(state.DomainID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Read Certificate Provisioning %s", state.DomainID),
			fmt.Sprintf("Could not parse domain ID %s: %s", state.DomainID, err.Error()),
		)
		return
	}

	provisioningStatus, diag := r.getCurrentProvisioningStatus(ctx, domainID)
	resp.Diagnostics.Append(diag)
	if resp.Diagnostics.HasError() {
		return
	}

	state.ProvisioningStatus = types.StringValue(provisioningStatus)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *certificateProvisioningResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Follow the same procedure as Create

	// Retrieve values from plan
	var plan certificateProvisioningResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domainID, err := uuid.Parse(plan.DomainID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Provision Certificates for Domain %s", plan.DomainID),
			fmt.Sprintf("Could not parse domain ID %s: %s", plan.DomainID, err.Error()),
		)
		return
	}

	// Call the API to trigger provisioning
	result, err := r.client.ProvisionDomainCertificatesWithResponse(ctx, domainID)

	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Provision Certificates for Domain %s", plan.DomainID),
			fmt.Sprintf("API returned error: %s", err.Error()),
		)
		return
	}
	if client.RespIsError(result) {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Provision Certificates for Domain %s", plan.DomainID),
			client.APIErrorDetail(result.HTTPResponse, result.Body),
		)
		return
	}

	// Call the API to get the provisioning status
	provisioningStatus, diag := r.getCurrentProvisioningStatus(ctx, domainID)
	resp.Diagnostics.Append(diag)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set provisioning status to plan
	plan.ProvisioningStatus = types.StringValue(provisioningStatus)

	// Set state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if provisioningStatus != "success" {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Provision Certificates for Domain %s", domainID),
			fmt.Sprintf("Provisioning status is %s, expected success", provisioningStatus),
		)
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *certificateProvisioningResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// noop
}

// Configure adds the provider configured client to the resource.
func (r *certificateProvisioningResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *certificateProvisioningResource) getCurrentProvisioningStatus(ctx context.Context, domainID uuid.UUID) (string, diag.Diagnostic) {
	domain, err := r.client.GetDomainWithResponse(ctx, domainID)
	if err != nil {
		d := diag.NewErrorDiagnostic(
			fmt.Sprintf("Failed to Get Domain Info for Domain %s", domainID),
			fmt.Sprintf("GetDomain API returned error: %s", err.Error()),
		)
		return "", d
	}
	if client.RespIsError(domain) {
		d := diag.NewErrorDiagnostic(
			fmt.Sprintf("Failed to Get Domain Info for Domain %s", domainID),
			client.APIErrorDetail(domain.HTTPResponse, domain.Body),
		)
		return "", d
	}

	status, err := domain.JSON200.ProvisioningStatus.ValueByDiscriminator()
	if err != nil {
		d := diag.NewErrorDiagnostic(
			"Failed to Get Provisioning Status",
			err.Error(),
		)
		return "", d
	}

	// Convert provisioning status to string representation
	ret := "unknown"
	switch status.(type) {
	case client.ProvisioningStatusSuccess:
		ret = "success"
	case client.ProvisioningStatusFailed:
		ret = "failed"
	case client.ProvisioningStatusPending:
		ret = "pending"
	case client.ProvisioningStatusManual:
		ret = "manual"
	}

	return ret, nil
}
