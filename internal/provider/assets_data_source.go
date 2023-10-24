package provider

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &assetsResource{}
)

func NewAssetsResource() datasource.DataSource {
	return &assetsResource{}
}

type assetsResource struct{}

func (d *assetsResource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_assets"
}

// Schema defines the schema for the data source.
func (d *assetsResource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `
A data source for a list of assets to be deployed.

For how to use this data source with deno_deployment resource, please refer to the doc of deno_deployment resource.
		`,
		Attributes: map[string]schema.Attribute{
			"path": schema.StringAttribute{
				Required:    true,
				Description: "The root directory path of the assets. e.g. `../dist`",
			},
			"pattern": schema.StringAttribute{
				Required:    true,
				Description: "The glob pattern to match the assets within the directory specified in `path`. e.g. `**/*.{js,ts,json}`",
			},
			"target": schema.StringAttribute{
				Optional: true,
				Description: `The target directory path where the assets will be put in the runtime virtual filesystem.
For example, if "target" is set to "foo/bar", then the assets will be placed under the directory "foo/bar" in the runtime virtual filesystem.

If this field is omitted, the assets will be put under the "." directory in the runtime virtual filesystem.
				`,
			},
			"output": schema.MapNestedAttribute{
				Computed:    true,
				Description: "The assets map, whose key is the asset path used in the runtime virtual filesystem, and the value is the asset metadata.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"kind": schema.StringAttribute{
							Computed:    true,
							Description: "The kind of the asset. It can be either `file` or `symlink`.",
						},
						"local_file_path": schema.StringAttribute{
							Computed:    true,
							Description: "The file path of the asset in the local filesystem.",
						},
						"runtime_target_path": schema.StringAttribute{
							Computed:    true,
							Description: "The target file path of the symlink in the the runtime virtual filesystem. It is only available for `symlink` asset.",
						},
					},
				},
			},
		},
	}
}

// assetsResourceModel maps the data source schema data.
type assetsResourceModel struct {
	Path           types.String `tfsdk:"path"`
	Pattern        types.String `tfsdk:"pattern"`
	Target         types.String `tfsdk:"target"`
	AssetsMetadata types.Map    `tfsdk:"output"`
}

// Read refreshes the Terraform state with the latest data.
func (d *assetsResource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Retrieve values from config
	var config assetsResourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	glob := filepath.Join(config.Path.ValueString(), config.Pattern.ValueString())
	paths, err := doublestar.FilepathGlob(glob)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Read Assets %s", glob),
			err.Error(),
		)
		return
	}

	metadata := map[string]attr.Value{}
	for _, path := range paths {
		stat, err := os.Lstat(path)
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Unable to Read Assets %s", path),
				fmt.Sprintf("Failed to get the stat of file %s: %s", path, err.Error()),
			)
			return
		}

		if stat.IsDir() {
			continue
		}

		value := map[string]attr.Value{
			"kind":                types.StringNull(),
			"local_file_path":     types.StringValue(path),
			"runtime_target_path": types.StringNull(),
		}

		if stat.Mode()&os.ModeSymlink == os.ModeSymlink {
			value["kind"] = types.StringValue("symlink")
			linkedTo, err := filepath.EvalSymlinks(path)
			if err != nil {
				resp.Diagnostics.AddError(
					fmt.Sprintf("Unable to Read Assets %s", path),
					fmt.Sprintf("Failed to get the destination path of %s: %s", path, err.Error()),
				)
				return
			}
			symlinkTargetRelpath, err := filepath.Rel(config.Path.ValueString(), linkedTo)
			if err != nil {
				resp.Diagnostics.AddError(
					fmt.Sprintf("Unable to Read Assets %s", path),
					fmt.Sprintf("Failed to get the relative path of %s: %s", path, err.Error()),
				)
				return
			}
			runtimeTargetPath := filepath.Join(config.Target.ValueString(), symlinkTargetRelpath)
			value["runtime_target_path"] = types.StringValue(runtimeTargetPath)
		} else {
			value["kind"] = types.StringValue("file")
		}

		obj, diags := types.ObjectValue(map[string]attr.Type{
			"kind":                types.StringType,
			"local_file_path":     types.StringType,
			"runtime_target_path": types.StringType,
		}, value)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		relpath, err := filepath.Rel(config.Path.ValueString(), path)
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Unable to Read Assets %s", path),
				fmt.Sprintf("Failed to get the relative path of %s: %s", path, err.Error()),
			)
			return
		}
		runtimeFilePath := filepath.Join(config.Target.ValueString(), relpath)
		metadata[runtimeFilePath] = obj
	}

	assetsMetadata, diags := types.MapValue(types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"kind":                types.StringType,
			"local_file_path":     types.StringType,
			"runtime_target_path": types.StringType,
		},
	}, metadata)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	config.AssetsMetadata = assetsMetadata

	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
