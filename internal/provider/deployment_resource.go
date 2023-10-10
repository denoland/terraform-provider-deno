package provider

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"terraform-provider-deno/client"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &deploymentResource{}
	_ resource.ResourceWithConfigure = &deploymentResource{}
)

// NewDeploymentResource is a helper function to simplify the provider implementation.
func NewDeploymentResource() resource.Resource {
	return &deploymentResource{}
}

// deploymentResource is the resource implementation.
type deploymentResource struct {
	client         client.ClientWithResponsesInterface
	organizationID uuid.UUID
}

// deploymentResourceModel maps the resource schema data.
type deploymentResourceModel struct {
	DeploymentID    types.String         `tfsdk:"deployment_id"`
	ProjectID       types.String         `tfsdk:"project_id"`
	Status          types.String         `tfsdk:"status"`
	DomainIDs       types.Set            `tfsdk:"domain_ids"`
	Domains         types.Set            `tfsdk:"domains"`
	EntryPointURL   types.String         `tfsdk:"entry_point_url"`
	ImportMapURL    types.String         `tfsdk:"import_map_url"`
	LockFileURL     types.String         `tfsdk:"lock_file_url"`
	CompilerOptions compilerOptionsModel `tfsdk:"compiler_options"`
	Assets          types.Map            `tfsdk:"assets"`
	UploadedAssets  types.Map            `tfsdk:"uploaded_assets"`
	EnvVars         types.Map            `tfsdk:"env_vars"`
	CreatedAt       types.String         `tfsdk:"created_at"`
	UpdatedAt       types.String         `tfsdk:"updated_at"`
	Timeouts        timeouts.Value       `tfsdk:"timeouts"`
}

// compilerOptionsModel maps the compiler options schema data.
type compilerOptionsModel struct {
	JSX                types.String `tfsdk:"jsx"`
	JSXFactory         types.String `tfsdk:"jsx_factory"`
	JSXFragmentFactory types.String `tfsdk:"jsx_fragment_factory"`
	JSXImportSource    types.String `tfsdk:"jsx_import_source"`
}

// Metadata returns the resource type name.
func (r *deploymentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployment"
}

// Schema defines the schema for the resource.
func (r *deploymentResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `
A resource for a Deno Deploy deployment.

A deployment belongs to a project, is an immutable, invokable snapshot of the project's assets, can be assigned custom domain(s).
		`,
		Attributes: map[string]schema.Attribute{
			"deployment_id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the deployment.",
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The project ID that this deployment belongs to.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: `The status of the deployment, indicating whether the deployment succeeded or not. It can be "failed", "pending", or "success"`,
			},
			"domain_ids": schema.SetAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: `The custom domain IDs to associate with the deployment. To associate, the domains must be verified for their ownership and their certificates must be ready. For further information, please refer to the doc of deno_domain resource.`,
			},
			"domains": schema.SetAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: `The domain(s) that can be used to access the deployment.`,
			},
			"entry_point_url": schema.StringAttribute{
				Required:    true,
				Description: "The path to the file that will be executed when the deployment is invoked.",
			},
			"import_map_url": schema.StringAttribute{
				Optional:    true,
				Description: "The path to the import map file. If this is omitted and a deno config file (`deno.json` or `deno.jsonc`) is found in the assets, the value in the config file will be used.",
			},
			"lock_file_url": schema.StringAttribute{
				Optional:    true,
				Description: "The path to the lock file. If this is omitted and a deno config file (`deno.json` or `deno.jsonc`) is found in the assets, the value in the config will be used.",
			},
			"compiler_options": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Compiler options to be used when building the deployment. If this is omitted and a deno config file (`deno.json` or `deno.jsonc`) is found in the assets, the value in the config file will be used.",
				Attributes: map[string]schema.Attribute{
					"jsx": schema.StringAttribute{
						Optional: true,
					},
					"jsx_factory": schema.StringAttribute{
						Optional: true,
					},
					"jsx_fragment_factory": schema.StringAttribute{
						Optional: true,
					},
					"jsx_import_source": schema.StringAttribute{
						Optional: true,
					},
				},
			},
			"assets": schema.MapNestedAttribute{
				Required:    true,
				Description: "The entities that compose the deployment. A key represents a path to the entity.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"kind": schema.StringAttribute{
							Required:    true,
							Description: `The kind of entity: "file" or "symlink".`,
						},
						"git_sha1": schema.StringAttribute{
							Optional:    true,
							Description: `The git object hash for the file. This is valid only for kind == "file".`,
						},
						"target": schema.StringAttribute{
							Optional:    true,
							Description: `The target file path for the symlink. This is valid only for kind == "symlink".`,
						},
						"updated_at": schema.StringAttribute{
							Optional:    true,
							Description: `The time the file was last updated. This is valid only for kind == "file".`,
						},
					},
				},
			},
			"uploaded_assets": schema.MapNestedAttribute{
				Computed:    true,
				Description: "The assets that have been uploaded in previous deployments, keyed with hash of the content. This is inteneded to be used to avoid uploading the same assets multiple times.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"path": schema.StringAttribute{
							Computed: true,
						},
						"git_sha1": schema.StringAttribute{
							Computed: true,
						},
						"updated_at": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
			"env_vars": schema.MapAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "The environment variables to be set in the runtime environment of the deployment.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				Description:         "The time the deployment was created, formmatting in RFC3339.",
				MarkdownDescription: "The time the deployment was created, formmatting in [RFC3339](https://datatracker.ietf.org/doc/html/rfc3339).",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				Description:         "The time the deployment was last updated, formmatting in RFC3339.",
				MarkdownDescription: "The time the deployment was last updated, formmatting in [RFC3339](https://datatracker.ietf.org/doc/html/rfc3339).",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
			}),
		},
	}
}

// encodePath applies URL encoding to the given path, with directory separator
// `/` preserved.
func encodePath(path string) string {
	arr := []string{}
	for _, part := range strings.Split(path, "/") {
		escaped := url.QueryEscape(part)
		arr = append(arr, escaped)
	}
	return strings.Join(arr, "/")
}

func prepareAssetsForUpload(ctx context.Context, plannedAssets types.Map) (client.Assets, diag.Diagnostic) {
	rootPath := "."
	assets := make(client.Assets)

	for path, metadata := range plannedAssets.Elements() {
		obj, ok := metadata.(types.Object)
		if !ok {
			return nil, diag.NewErrorDiagnostic(
				"Unable to Create Deployment",
				fmt.Sprintf("Could not parse asset metadata for %s", path),
			)
		}
		metadataValues := obj.Attributes()

		relpath, err := filepath.Rel(rootPath, path)
		if err != nil {
			return nil, diag.NewErrorDiagnostic(
				"Unable to Create Deployment",
				fmt.Sprintf("Could not get file path relative to the current directory. target: %s", path),
			)
		}

		kind, ok := metadataValues["kind"].(types.String)
		if !ok {
			return nil, diag.NewErrorDiagnostic(
				"Unable to Create Deployment",
				fmt.Sprintf("Could not parse asset kind for %s. Expected string, but got %s", path, metadataValues["kind"].Type(ctx)),
			)
		}

		switch kind.ValueString() {
		case "file":
			b, err := os.ReadFile(path)
			if err != nil {
				return nil, diag.NewErrorDiagnostic(
					"Unable to Create Deployment",
					fmt.Sprintf("Could not read file content for %s", path),
				)
			}
			var fileAsset client.FileAsset
			var fileContent client.FileAsset0
			if utf8.Valid(b) {
				enc := client.Utf8
				fileContent = client.FileAsset0{
					Content:  string(b),
					Encoding: &enc,
				}
			} else {
				enc := client.Base64
				fileContent = client.FileAsset0{
					Content:  base64.StdEncoding.EncodeToString(b),
					Encoding: &enc,
				}
			}

			err = fileAsset.FromFileAsset0(fileContent)
			if err != nil {
				return nil, diag.NewErrorDiagnostic(
					"Unable to Create Deployment",
					fmt.Sprintf("Internal error happened for %s on FromFileAsset0", path),
				)
			}
			var asset client.Asset
			err = asset.FromFileAsset(fileAsset)
			if err != nil {
				return nil, diag.NewErrorDiagnostic(
					"Unable to Create Deployment",
					fmt.Sprintf("Internal error happened for %s on FromFileAsset", path),
				)
			}

			assets[encodePath(relpath)] = asset
		case "symlink":
			targetPath, ok := metadataValues["target"].(types.String)
			if !ok {
				return nil, diag.NewErrorDiagnostic(
					"Unable to Create Deployment",
					fmt.Sprintf("Could not parse target path for %s. Expected string, but got %s", path, metadataValues["target"].Type(ctx)),
				)
			}

			targetRel, err := filepath.Rel(rootPath, targetPath.ValueString())
			if err != nil {
				return nil, diag.NewErrorDiagnostic(
					"Unable to Create Deployment",
					fmt.Sprintf("Could not get file path relative to the current directory. target: %s", targetPath),
				)
			}
			symlinkAsset := client.SymlinkAsset{
				Target: encodePath(targetRel),
				Kind:   client.SymlinkAssetKindSymlink,
			}

			var asset client.Asset
			err = asset.FromSymlinkAsset(symlinkAsset)
			if err != nil {
				return nil, diag.NewErrorDiagnostic(
					"Unable to Create Deployment",
					fmt.Sprintf("Internal error happened for %s on FromSymlinkAsset", path),
				)
			}

			assets[encodePath(relpath)] = asset
		default:
			return nil, diag.NewErrorDiagnostic(
				"Unable to Create Deployment",
				fmt.Sprintf("Invalid asset kind %s is found for %s. Valid kinds are `file`, `symlink`", kind.ValueString(), path),
			)
		}
	}

	if len(assets) == 0 {
		return nil, diag.NewErrorDiagnostic(
			"Unable to Create Deployment",
			"No assets are found. At least one asset is required.",
		)
	}

	return assets, nil
}

// Create creates the resource and sets the initial Terraform state.
func (r *deploymentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan deploymentResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Do deployment
	diags = r.doDeployment(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *deploymentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state deploymentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deploymentID := state.DeploymentID.ValueString()

	diags = r.updateModel(ctx, deploymentID, &state)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *deploymentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Follow the same procedure as Create

	// Retrieve values from plan
	var plan deploymentResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Do deployment
	diags = r.doDeployment(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *deploymentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// No opeation is needed for deployment since it is immutable,
	// but we disassociate the domain (if any) from the deployment.

	// Get current state
	var state deploymentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// No domain association; nothing to do
	domainIDs := []string{}
	diags = state.DomainIDs.ElementsAs(ctx, &domainIDs, true)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// No domain association; nothing to do
	if len(domainIDs) == 0 {
		return
	}

	for _, rawDomainID := range domainIDs {
		domainID, err := uuid.Parse(rawDomainID)
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Unable to Disassociate the Domain %s from the Deployment %s", rawDomainID, state.DeploymentID),
				fmt.Sprintf("Could not parse domain ID %s: %s", rawDomainID, err.Error()),
			)
			continue
		}

		// Call the API to disassociate the domain
		result, err := r.client.UpdateDomainAssociationWithResponse(ctx, domainID, client.UpdateDomainAssociationRequest{
			DeploymentId: nil,
		})
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Unable to Disassociate the Domain %s from the Deployment %s", rawDomainID, state.DeploymentID),
				err.Error(),
			)
			continue
		}
		if client.RespIsError(result) {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Unable to Disassociate the Domain %s from the Deployment %s", rawDomainID, state.DeploymentID),
				client.APIErrorDetail(result.HTTPResponse, result.Body),
			)
			continue
		}
	}
}

// Configure adds the provider configured client to the resource.
func (r *deploymentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *deploymentResource) doDeployment(ctx context.Context, plan *deploymentResourceModel) diag.Diagnostics {
	accumulatedDiags := diag.Diagnostics{}

	// Validate and parse project ID as UUID
	projectID, err := uuid.Parse(plan.ProjectID.ValueString())
	if err != nil {
		accumulatedDiags.AddError(
			fmt.Sprintf("Unable to Create Deployment for Project %s", plan.ProjectID),
			fmt.Sprintf("Could not parse project ID %s: %s", plan.ProjectID, err.Error()),
		)
		return accumulatedDiags
	}

	// Validate and parse domain IDs as UUID (if present)
	var rawDomainIDs []string
	var domainIDs []uuid.UUID
	diags := plan.DomainIDs.ElementsAs(ctx, &rawDomainIDs, true)
	accumulatedDiags.Append(diags...)
	if accumulatedDiags.HasError() {
		return accumulatedDiags
	}
	for _, rawDomainID := range rawDomainIDs {
		d, err := uuid.Parse(rawDomainID)
		if err != nil {
			accumulatedDiags.AddError(
				fmt.Sprintf("Unable to Create Deployment for Project %s", plan.ProjectID),
				fmt.Sprintf("Could not parse domain ID %s: %s", rawDomainID, err.Error()),
			)
			return accumulatedDiags
		}
		domainIDs = append(domainIDs, d)
	}

	assets, diag := prepareAssetsForUpload(ctx, plan.Assets)
	accumulatedDiags.Append(diag)
	if accumulatedDiags.HasError() {
		return accumulatedDiags
	}

	var envVars map[string]string
	diags = plan.EnvVars.ElementsAs(ctx, &envVars, true)
	accumulatedDiags.Append(diags...)
	if accumulatedDiags.HasError() {
		return accumulatedDiags
	}

	res, err := r.client.CreateDeploymentWithResponse(ctx, projectID, client.CreateDeploymentRequest{
		Assets: assets,
		CompilerOptions: &client.CompilerOptions{
			Jsx:                plan.CompilerOptions.JSX.ValueStringPointer(),
			JsxFactory:         plan.CompilerOptions.JSXFactory.ValueStringPointer(),
			JsxFragmentFactory: plan.CompilerOptions.JSXFragmentFactory.ValueStringPointer(),
			JsxImportSource:    plan.CompilerOptions.JSXImportSource.ValueStringPointer(),
		},
		EntryPointUrl: plan.EntryPointURL.ValueString(),
		EnvVars:       envVars,
		ImportMapUrl:  plan.ImportMapURL.ValueStringPointer(),
		LockFileUrl:   plan.LockFileURL.ValueStringPointer(),
	})
	if err != nil {
		accumulatedDiags.AddError(
			fmt.Sprintf("Unable to Create Deployment for Project %s", plan.ProjectID),
			err.Error(),
		)
		return accumulatedDiags
	}
	if client.RespIsError(res) {
		accumulatedDiags.AddError(
			fmt.Sprintf("Unable to Create Deployment for Project %s", plan.ProjectID),
			client.APIErrorDetail(res.HTTPResponse, res.Body),
		)
		return accumulatedDiags
	}

	deploymentID := res.JSON200.Id

	buildLogs, err := r.client.GetBuildLogsWithResponse(ctx, deploymentID, func(ctx context.Context, req *http.Request) error {
		req.Header.Add("Accept", "application/json")
		return nil
	})
	if err != nil {
		accumulatedDiags.AddError(
			"Deployment Initiated, but Failed to Get Build Logs",
			fmt.Sprintf("Deployment ID: %s, Error: %s", deploymentID, err.Error()),
		)
		return accumulatedDiags
	}
	if client.RespIsError(buildLogs) {
		accumulatedDiags.AddError(
			"Deployment Initiated, but Failed to Get Build Logs",
			fmt.Sprintf("Deployment ID: %s, Error: %s", deploymentID, client.APIErrorDetail(buildLogs.HTTPResponse, buildLogs.Body)),
		)
		return accumulatedDiags
	}

	logs := make([]string, len(*buildLogs.JSON200))
	for i, logline := range *buildLogs.JSON200 {
		logs[i] = fmt.Sprintf("[%s] %s", logline.Level, logline.Message)
	}

	diags = r.updateModel(ctx, deploymentID, plan)
	accumulatedDiags.Append(diags...)
	if accumulatedDiags.HasError() {
		return accumulatedDiags
	}

	// Ensure the deployment has succeeded
	if plan.Status.ValueString() != string(client.DeploymentStatusSuccess) {
		accumulatedDiags.AddError(
			"Deployment Failed",
			fmt.Sprintf(`Deployment ID: %s
Status: %s

Build logs:
%s
`, deploymentID, plan.Status, strings.Join(logs, "\n")),
		)
		return accumulatedDiags
	}

	// Save the uploaded assets to the state so we can avoid duplicate uploads in future deployments.
	// TODO: we haven't implemented the logic to avoid duplicate uploads
	uploadedAssets, diags := types.MapValue(types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"path":       types.StringType,
			"git_sha1":   types.StringType,
			"updated_at": types.StringType,
		},
	}, map[string]attr.Value{})
	accumulatedDiags.Append(diags...)
	if accumulatedDiags.HasError() {
		return accumulatedDiags
	}
	plan.UploadedAssets = uploadedAssets

	// Associate the custom domain (if present)
	if len(domainIDs) > 0 {
		for _, domainID := range domainIDs {
			result, err := r.client.UpdateDomainAssociationWithResponse(ctx, domainID, client.UpdateDomainAssociationRequest{
				DeploymentId: &deploymentID,
			})
			if err != nil {
				accumulatedDiags.AddError(
					fmt.Sprintf("Unable to Associate the Domain %s with the Deployment %s", domainID, deploymentID),
					err.Error(),
				)
				return accumulatedDiags
			}
			if client.RespIsError(result) {
				accumulatedDiags.AddError(
					fmt.Sprintf("Unable to Associate the Domain %s with the Deployment %s", domainID, deploymentID),
					client.APIErrorDetail(result.HTTPResponse, result.Body),
				)
				return accumulatedDiags
			}
		}

		// Custom domain association succeeded; call the get deployment API again to obtain the updated domain list
		r.updateModel(ctx, deploymentID, plan)
	}

	return accumulatedDiags
}

// UpdateModel updates the resource model with the latest information obtained by making a API call.
func (r *deploymentResource) updateModel(ctx context.Context, deploymentID string, model *deploymentResourceModel) diag.Diagnostics {
	accumulatedDiags := diag.Diagnostics{}

	deployment, err := r.client.GetDeploymentWithResponse(ctx, deploymentID)
	if err != nil {
		accumulatedDiags.AddError(
			"Failed to Get Deployment Details",
			fmt.Sprintf("Deployment ID: %s, Error: %s", deploymentID, err.Error()),
		)
		return accumulatedDiags
	}
	if client.RespIsError(deployment) {
		accumulatedDiags.AddError(
			"Failed to Get Deployment Details",
			fmt.Sprintf("Deployment ID: %s, Error: %s", deploymentID, client.APIErrorDetail(deployment.HTTPResponse, deployment.Body)),
		)
		return accumulatedDiags
	}

	model.DeploymentID = types.StringValue(deployment.JSON200.Id)
	model.Status = types.StringValue(string(deployment.JSON200.Status))
	domainElems := []attr.Value{}
	if deployment.JSON200.Domains != nil {
		for _, d := range *deployment.JSON200.Domains {
			domainElems = append(domainElems, types.StringValue(d))
		}
	}
	domainSet, diags := types.SetValue(basetypes.StringType{}, domainElems)
	accumulatedDiags.Append(diags...)
	if accumulatedDiags.HasError() {
		return accumulatedDiags
	}
	model.Domains = domainSet
	model.CreatedAt = types.StringValue(deployment.JSON200.CreatedAt.Format(time.RFC3339))
	model.UpdatedAt = types.StringValue(deployment.JSON200.UpdatedAt.Format(time.RFC3339))

	return accumulatedDiags
}
