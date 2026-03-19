package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/serenityzn/terraform-provider-deepgram/internal/deepgram"
)

var _ datasource.DataSource = &projectDataSource{}

type projectDataSource struct {
	client *deepgram.Client
}

type projectDataSourceModel struct {
	Name      types.String `tfsdk:"name"`
	ProjectID types.String `tfsdk:"project_id"`
	MIPOptOut types.Bool   `tfsdk:"mip_opt_out"`
}

func NewProjectDataSource() datasource.DataSource {
	return &projectDataSource{}
}

func (d *projectDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (d *projectDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single Deepgram project by name.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the project to look up.",
			},
			"project_id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the project.",
			},
			"mip_opt_out": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the project has opted out of the Model Improvement Program.",
			},
		},
	}
}

func (d *projectDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*deepgram.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected data source configure type",
			fmt.Sprintf("Expected *deepgram.Client, got: %T", req.ProviderData),
		)
		return
	}
	d.client = client
}

func (d *projectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config projectDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	listed, err := d.client.ListProjects(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list Deepgram projects", err.Error())
		return
	}

	name := config.Name.ValueString()
	var matches []deepgram.Project
	for _, p := range listed.Projects {
		if p.Name == name {
			matches = append(matches, p)
		}
	}

	switch len(matches) {
	case 0:
		resp.Diagnostics.AddError(
			"Project not found",
			fmt.Sprintf("No Deepgram project found with name %q.", name),
		)
		return
	case 1:
		// expected — fall through
	default:
		resp.Diagnostics.AddError(
			"Ambiguous project name",
			fmt.Sprintf("Found %d projects named %q. Use the deepgram_projects data source to list all and select by project_id.", len(matches), name),
		)
		return
	}

	p := matches[0]
	resp.Diagnostics.Append(resp.State.Set(ctx, &projectDataSourceModel{
		Name:      types.StringValue(p.Name),
		ProjectID: types.StringValue(p.ProjectID),
		MIPOptOut: types.BoolValue(p.MIPOptOut),
	})...)
}
