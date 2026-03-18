package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/serenityzn/terraform-provider-deepgram-/internal/deepgram"
)

var _ datasource.DataSource = &keysDataSource{}

type keysDataSource struct {
	client *deepgram.Client
}

type keysDataSourceModel struct {
	ProjectID types.String `tfsdk:"project_id"`
	APIKeys   types.List   `tfsdk:"api_keys"`
}

// keyItemModel mirrors one entry returned by the list endpoint.
type keyItemModel struct {
	APIKeyID string   `tfsdk:"api_key_id"`
	Comment  string   `tfsdk:"comment"`
	Scopes   []string `tfsdk:"scopes"`
	Created  string   `tfsdk:"created"`
	MemberID string   `tfsdk:"member_id"`
	Email    string   `tfsdk:"email"`
}

var keyItemAttrTypes = map[string]attr.Type{
	"api_key_id": types.StringType,
	"comment":    types.StringType,
	"scopes":     types.SetType{ElemType: types.StringType},
	"created":    types.StringType,
	"member_id":  types.StringType,
	"email":      types.StringType,
}

func NewKeysDataSource() datasource.DataSource {
	return &keysDataSource{}
}

func (d *keysDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_keys"
}

func (d *keysDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all API keys for a Deepgram project.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project whose API keys to list.",
			},
			"api_keys": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The list of API keys for the project.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"api_key_id": schema.StringAttribute{
							Computed:    true,
							Description: "The unique identifier of the API key.",
						},
						"comment": schema.StringAttribute{
							Computed:    true,
							Description: "The comment/label for the API key.",
						},
					"scopes": schema.SetAttribute{
						Computed:    true,
						ElementType: types.StringType,
						Description: "The permission scopes assigned to the key.",
					},
						"created": schema.StringAttribute{
							Computed:    true,
							Description: "The creation timestamp of the API key.",
						},
						"member_id": schema.StringAttribute{
							Computed:    true,
							Description: "The member ID associated with the key.",
						},
						"email": schema.StringAttribute{
							Computed:    true,
							Description: "The email of the member who owns the key.",
						},
					},
				},
			},
		},
	}
}

func (d *keysDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *keysDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state keysDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	listed, err := d.client.ListKeys(ctx, state.ProjectID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to list Deepgram API keys", err.Error())
		return
	}

	keyObjects := make([]attr.Value, 0, len(listed.APIKeys))
	for _, item := range listed.APIKeys {
		scopeVals := make([]attr.Value, 0, len(item.APIKey.Scopes))
		for _, s := range item.APIKey.Scopes {
			scopeVals = append(scopeVals, types.StringValue(s))
		}
		scopeSet, diags := types.SetValue(types.StringType, scopeVals)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		obj, diags := types.ObjectValue(keyItemAttrTypes, map[string]attr.Value{
			"api_key_id": types.StringValue(item.APIKey.APIKeyID),
			"comment":    types.StringValue(item.APIKey.Comment),
			"scopes":     scopeSet,
			"created":    types.StringValue(item.APIKey.Created),
			"member_id":  types.StringValue(item.Member.MemberID),
			"email":      types.StringValue(item.Member.Email),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		keyObjects = append(keyObjects, obj)
	}

	keyList, diags := types.ListValue(
		types.ObjectType{AttrTypes: keyItemAttrTypes},
		keyObjects,
	)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.APIKeys = keyList
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
