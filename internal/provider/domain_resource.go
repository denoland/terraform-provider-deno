package provider

import (
	"context"
	"fmt"
	"strings"
	"terraform-provider-deno/client"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &domainResource{}
	_ resource.ResourceWithConfigure = &domainResource{}
)

// NewDomainResource is a helper function to simplify the provider implementation.
func NewDomainResource() resource.Resource {
	return &domainResource{}
}

// domainResource is the resource implementation.
type domainResource struct {
	client         client.ClientWithResponsesInterface
	organizationID uuid.UUID
}

// domainResourceModel maps the resource schema data.
type domainResourceModel struct {
	ID            types.String   `tfsdk:"id"`
	Domain        types.String   `tfsdk:"domain"`
	Token         types.String   `tfsdk:"token"`
	DNSRecords    types.List     `tfsdk:"dns_records"`
	DNSRecordsMap *dnsRecordsMap `tfsfk:"dns_records_map"`
	CreatedAt     types.String   `tfsdk:"created_at"`
	UpdatedAt     types.String   `tfsdk:"updated_at"`
}

// dnsRecordsMap represents A, AAAA, and CNAME records to be configured.
type dnsRecordsMap struct {
	A     dnsRecord `tfsdk:"a"`
	AAAA  dnsRecord `tfsdk:"aaaa"`
	CNAME dnsRecord `tfsdk:"cname"`
}

// dnsRecord represents a single DNS record.
type dnsRecord struct {
	Name    types.String `tfsdk:"name"`
	Content types.String `tfsdk:"content"`
}

// Metadata returns the resource type name.
func (r *domainResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

// Schema defines the schema for the resource.
func (r *domainResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `
A resource for a custom domain.

A custom domain is a per-organization resource, which can be associated with a deployment.
In order to associate a custom domain with a deployment, you need to verify the ownership of the domain, then prepare TLS certificates for the domain. Refer to the example below for practical usage.
		`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "The ID of the domain.",
			},
			"domain": schema.StringAttribute{
				Required:    true,
				Description: "The custom domain, such as `foo.example.com`",
			},
			"token": schema.StringAttribute{
				Computed:    true,
				Description: "The token used for verifying the ownership of the domain.",
			},
			"dns_records_map": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "The DNS records that need to be added to the DNS nameserver.",
				Attributes: map[string]schema.Attribute{
					"a": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"name": schema.StringAttribute{
								Computed:    true,
								Description: "The name of the DNS record.",
							},
							"content": schema.StringAttribute{
								Computed:    true,
								Description: "The content of the DNS record.",
							},
						},
					},
					"aaaa": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"name": schema.StringAttribute{
								Computed:    true,
								Description: "The name of the DNS record.",
							},
							"content": schema.StringAttribute{
								Computed:    true,
								Description: "The content of the DNS record.",
							},
						},
					},
					"cname": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"name": schema.StringAttribute{
								Computed:    true,
								Description: "The name of the DNS record.",
							},
							"content": schema.StringAttribute{
								Computed:    true,
								Description: "The content of the DNS record.",
							},
						},
					},
				},
			},
			"dns_records": schema.ListNestedAttribute{
				Computed:           true,
				Description:        "The DNS records that need to be added to the DNS nameserver.",
				DeprecationMessage: "This attribute is deprecated. Please use `dns_records_map` instead.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Computed:    true,
							Description: "The type of the DNS record such as `A`, `CNAME`, etc.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the DNS record.",
						},
						"content": schema.StringAttribute{
							Computed:    true,
							Description: "The content of the DNS record. The value depends on the type of the DNS record. For example, for `A` record, it is the IP address of the domain.",
						},
					},
				},
			},
			"created_at": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description:         "The time the domain was created, formmatting in RFC3339.",
				MarkdownDescription: "The time the domain was created, formmatting in [RFC3339](https://datatracker.ietf.org/doc/html/rfc3339).",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				Description:         "The time the domain was last updated, formmatting in RFC3339.",
				MarkdownDescription: "The time the domain was updated, formmatting in [RFC3339](https://datatracker.ietf.org/doc/html/rfc3339).",
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *domainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan domainResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call "create domain" API
	domain, err := r.client.CreateDomainWithResponse(ctx, r.organizationID, client.CreateDomainJSONRequestBody{
		Domain: plan.Domain.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Create Domain %s", plan.Domain.ValueString()),
			err.Error(),
		)
		return
	}
	if client.RespIsError(domain) {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Create Domain %s", plan.Domain.ValueString()),
			client.APIErrorDetail(domain.HTTPResponse, domain.Body),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(domain.JSON200.Id.String())
	plan.Domain = types.StringValue(domain.JSON200.Domain)
	plan.Token = types.StringValue(domain.JSON200.Token)
	plan.CreatedAt = types.StringValue(domain.JSON200.CreatedAt.Format(time.RFC3339))
	plan.UpdatedAt = types.StringValue(domain.JSON200.UpdatedAt.Format(time.RFC3339))

	dnsRecords, diags := convertToDNSRecordsList(domain.JSON200.DnsRecords)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.DNSRecords = dnsRecords

	dnsRecordsMap, diags := convertToDNSRecordsMap(domain.JSON200.DnsRecords)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.DNSRecordsMap = &dnsRecordsMap

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func convertToDNSRecordsList(dnsRecords []client.DnsRecord) (types.List, diag.Diagnostics) {
	dnsRecordType := map[string]attr.Type{
		"type":    types.StringType,
		"name":    types.StringType,
		"content": types.StringType,
	}
	ty := types.ObjectType{
		AttrTypes: dnsRecordType,
	}

	records := make([]attr.Value, len(dnsRecords))
	for i, dnsRecord := range dnsRecords {
		elements := map[string]attr.Value{
			"type":    types.StringValue(dnsRecord.Type),
			"name":    types.StringValue(dnsRecord.Name),
			"content": types.StringValue(dnsRecord.Content),
		}
		objectValue, diags := types.ObjectValue(dnsRecordType, elements)
		if diags.HasError() {
			return types.ListNull(ty), diags
		}
		records[i] = objectValue
	}

	dnsRecordsList, diags := types.ListValue(ty, records)
	if diags.HasError() {
		return types.ListNull(ty), diags
	}

	return dnsRecordsList, nil
}

func convertToDNSRecordsMap(dnsRecords []client.DnsRecord) (dnsRecordsMap, diag.Diagnostics) {
	var a *dnsRecord
	var aaaa *dnsRecord
	var cname *dnsRecord

	for _, record := range dnsRecords {
		switch record.Type {
		case "A":
			a = &dnsRecord{
				Name:    types.StringValue(record.Name),
				Content: types.StringValue(record.Content),
			}
		case "AAAA":
			aaaa = &dnsRecord{
				Name:    types.StringValue(record.Name),
				Content: types.StringValue(record.Content),
			}
		case "CNAME":
			cname = &dnsRecord{
				Name:    types.StringValue(record.Name),
				Content: types.StringValue(record.Content),
			}
		}
	}

	missingRecords := []string{}
	if a == nil {
		missingRecords = append(missingRecords, "A")
	}
	if aaaa == nil {
		missingRecords = append(missingRecords, "AAAA")
	}
	if cname == nil {
		missingRecords = append(missingRecords, "CNAME")
	}

	if len(missingRecords) > 0 {
		diags := diag.Diagnostics{}
		diags.AddError("Missing DNS records", fmt.Sprintf("The DNS records obtained from API are missing %s records.", strings.Join(missingRecords, ", ")))
		return dnsRecordsMap{}, diags
	}

	return dnsRecordsMap{
		A:     *a,
		AAAA:  *aaaa,
		CNAME: *cname,
	}, nil
}

// Read refreshes the Terraform state with the latest data.
func (r *domainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state domainResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domainID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Read Domain %s", state.ID),
			fmt.Sprintf("Could not parse domain ID %s: %s", state.ID, err.Error()),
		)
		return
	}

	// Get refreshed order value from Deno Deploy
	domain, err := r.client.GetDomainWithResponse(ctx, domainID)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Read Domain %s", state.ID),
			fmt.Sprintf("Could not find domain with ID %s: %s", state.ID, err.Error()),
		)
		return
	}
	if domain.StatusCode() != 200 {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Read Domain %s", state.ID),
			client.APIErrorDetail(domain.HTTPResponse, domain.Body),
		)
		return
	}

	// Overwtite state with refreshed values
	state.ID = types.StringValue(domain.JSON200.Id.String())
	state.Domain = types.StringValue(domain.JSON200.Domain)
	state.Token = types.StringValue(domain.JSON200.Token)
	state.CreatedAt = types.StringValue(domain.JSON200.CreatedAt.Format(time.RFC3339))
	state.UpdatedAt = types.StringValue(domain.JSON200.UpdatedAt.Format(time.RFC3339))

	dnsRecords, diags := convertToDNSRecordsList(domain.JSON200.DnsRecords)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.DNSRecords = dnsRecords

	dnsRecordsMap, diags := convertToDNSRecordsMap(domain.JSON200.DnsRecords)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.DNSRecordsMap = &dnsRecordsMap

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
// Internally, this is equivalent to deleting the resource and creating a new one.
func (r *domainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan domainResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domainID, err := uuid.Parse(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Update Domain %s", plan.ID),
			fmt.Sprintf("Could not parse domain ID %s: %s", plan.ID, err.Error()),
		)
		return
	}

	// Delete existing domain
	result, err := r.client.DeleteDomainWithResponse(ctx, domainID)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Update Domain %s", plan.ID),
			fmt.Sprintf("Could not delete domain with ID %s: %s", plan.ID, err.Error()),
		)
		return
	}
	if result.StatusCode() != 200 {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to delete Domain %s", plan.ID),
			client.APIErrorDetail(result.HTTPResponse, result.Body),
		)
		return
	}

	// Create a new domain
	domain, err := r.client.CreateDomainWithResponse(ctx, r.organizationID, client.CreateDomainJSONRequestBody{
		Domain: plan.Domain.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Update Domain %s", plan.Domain.ValueString()),
			err.Error(),
		)
		return
	}
	if client.RespIsError(domain) {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Update Domain %s", plan.Domain.ValueString()),
			client.APIErrorDetail(domain.HTTPResponse, domain.Body),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(domain.JSON200.Id.String())
	plan.Domain = types.StringValue(domain.JSON200.Domain)
	plan.Token = types.StringValue(domain.JSON200.Token)
	plan.CreatedAt = types.StringValue(domain.JSON200.CreatedAt.Format(time.RFC3339))
	plan.UpdatedAt = types.StringValue(domain.JSON200.UpdatedAt.Format(time.RFC3339))

	dnsRecords, diags := convertToDNSRecordsList(domain.JSON200.DnsRecords)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.DNSRecords = dnsRecords

	dnsRecordsMap, diags := convertToDNSRecordsMap(domain.JSON200.DnsRecords)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.DNSRecordsMap = &dnsRecordsMap

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *domainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state domainResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domainID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Delete Domain %s", state.ID),
			fmt.Sprintf("Could not parse domain ID %s: %s", state.ID, err.Error()),
		)
		return
	}

	// Delete existing domain
	result, err := r.client.DeleteDomainWithResponse(ctx, domainID)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Delete Domain %s", state.ID),
			fmt.Sprintf("Could not delete domain with ID %s: %s", state.ID, err.Error()),
		)
		return
	}
	if result.StatusCode() != 200 {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Delete Domain %s", state.ID),
			client.APIErrorDetail(result.HTTPResponse, result.Body),
		)
		return
	}
}

// Configure adds the provider configured client to the resource.
func (r *domainResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
