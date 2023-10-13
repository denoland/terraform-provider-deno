package provider

import (
	"context"
	"fmt"
	"os"
	"time"

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
			"glob": schema.StringAttribute{
				Required:    true,
				Description: "The glob pattern to match the assets to be deployed. e.g. `**/*.ts`, `**/*.{ts,tsx,json}`",
			},
			"output": schema.MapNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"kind": schema.StringAttribute{
							Computed:    true,
							Description: "The kind of the asset. It can be either `file` or `symlink`.",
						},
						"git_sha1": schema.StringAttribute{
							Computed:    true,
							Description: "The git object hash of the asset. It is only available for `file` asset.",
						},
						"target": schema.StringAttribute{
							Computed:    true,
							Description: "The target path of the asset. It is only available for `symlink` asset.",
						},
						"updated_at": schema.StringAttribute{
							Computed:    true,
							Description: "The last updated time of the asset. It is only available for `file` asset.",
						},
					},
				},
			},
		},
	}
}

// assetsResourceModel maps the data source schema data.
type assetsResourceModel struct {
	AssetsGlob     types.String `tfsdk:"glob"`
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

	paths, err := doublestar.FilepathGlob(config.AssetsGlob.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Read Assets %s", config.AssetsGlob.ValueString()),
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
			"kind":       types.StringNull(),
			"git_sha1":   types.StringNull(),
			"target":     types.StringNull(),
			"updated_at": types.StringValue(stat.ModTime().Format(time.RFC3339Nano)),
		}

		if stat.Mode()&os.ModeSymlink == os.ModeSymlink {
			value["kind"] = types.StringValue("symlink")
			linkedTo, err := os.Readlink(path)
			if err != nil {
				resp.Diagnostics.AddError(
					fmt.Sprintf("Unable to Read Assets %s", path),
					fmt.Sprintf("Failed to get the destination path of %s: %s", path, err.Error()),
				)
				return
			}
			value["target"] = types.StringValue(linkedTo)
		} else {
			value["kind"] = types.StringValue("file")

			b, err := os.ReadFile(path)
			if err != nil {
				resp.Diagnostics.AddError(
					fmt.Sprintf("Unable to Calculate Git Object Hash for %s", path),
					err.Error(),
				)
				return
			}

			value["git_sha1"] = types.StringValue(calculateGitSha1(b))
		}

		obj, diags := types.ObjectValue(map[string]attr.Type{
			"kind":       types.StringType,
			"git_sha1":   types.StringType,
			"target":     types.StringType,
			"updated_at": types.StringType,
		}, value)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		metadata[path] = obj
	}

	assetsMetadata, diags := types.MapValue(types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"kind":       types.StringType,
			"git_sha1":   types.StringType,
			"target":     types.StringType,
			"updated_at": types.StringType,
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
