package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/serenityzn/terraform-provider-deepgram/internal/deepgram"
)

var _ datasource.DataSource = &projectsDataSource{}

type projectsDataSource struct {
	client *deepgram.Client
}

type projectsDataSourceModel struct {
	Projects types.List `tfsdk:"projects"`
}

var projectAttrTypes = map[string]attr.Type{
	"project_id": types.StringType,
	"name":       types.StringType,
	"mip_opt_out": types.BoolType,
}

func NewProjectsDataSource() datasource.DataSource {
	return &projectsDataSource{}
}

func (d *projectsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_projects"
}

func (d *projectsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all Deepgram projects associated with the API key.",
		Attributes: map[string]schema.Attribute{
			"projects": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The list of projects.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"project_id": schema.StringAttribute{
							Computed:    true,
							Description: "The unique identifier of the project.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the project.",
						},
						"mip_opt_out": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the project has opted out of the Model Improvement Program.",
						},
					},
				},
			},
		},
	}
}

func (d *projectsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *projectsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	listed, err := d.client.ListProjects(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list Deepgram projects", err.Error())
		return
	}

	projectObjects := make([]attr.Value, 0, len(listed.Projects))
	for _, p := range listed.Projects {
		obj, diags := types.ObjectValue(projectAttrTypes, map[string]attr.Value{
			"project_id":  types.StringValue(p.ProjectID),
			"name":        types.StringValue(p.Name),
			"mip_opt_out": types.BoolValue(p.MIPOptOut),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		projectObjects = append(projectObjects, obj)
	}

	projectList, diags := types.ListValue(
		types.ObjectType{AttrTypes: projectAttrTypes},
		projectObjects,
	)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &projectsDataSourceModel{
		Projects: projectList,
	})...)
}
