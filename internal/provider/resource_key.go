package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/serenityzn/terraform-provider-deepgram-/internal/deepgram"
)

var _ resource.Resource = &keyResource{}

type keyResource struct {
	client *deepgram.Client
}

type keyResourceModel struct {
	ID             types.String `tfsdk:"id"`
	ProjectID      types.String `tfsdk:"project_id"`
	Comment        types.String `tfsdk:"comment"`
	Scopes         types.Set    `tfsdk:"scopes"`
	Tags           types.List   `tfsdk:"tags"`
	ExpirationDate types.String `tfsdk:"expiration_date"`
	Key            types.String `tfsdk:"key"`
}

func NewKeyResource() resource.Resource {
	return &keyResource{}
}

func (r *keyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_key"
}

func (r *keyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Deepgram project API key.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the API key (api_key_id).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project this key belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"comment": schema.StringAttribute{
				Required:    true,
				Description: "A human-readable label for the API key.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			// Set (not List) because Deepgram returns scopes in arbitrary order
			// and may append additional scopes beyond what was requested.
			// We store only the user-requested scopes; extra API-added scopes
			// are intentionally ignored to keep state consistent.
			"scopes": schema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "Set of permission scopes for this key (e.g. [\"usage:read\", \"keys:write\"]).",
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
			},
			"tags": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Optional list of tags to attach to the key.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"expiration_date": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Optional expiration date/time for the key in RFC 3339 format (e.g. \"2026-01-01T00:00:00Z\").",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"key": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The secret API key value. Only available at creation time.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *keyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*deepgram.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected resource configure type",
			fmt.Sprintf("Expected *deepgram.Client, got: %T", req.ProviderData),
		)
		return
	}
	r.client = client
}

func (r *keyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan keyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	scopes := make([]string, 0, len(plan.Scopes.Elements()))
	resp.Diagnostics.Append(plan.Scopes.ElementsAs(ctx, &scopes, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tags := make([]string, 0, len(plan.Tags.Elements()))
	resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &tags, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := deepgram.CreateKeyRequest{
		Comment: plan.Comment.ValueString(),
		Scopes:  scopes,
		Tags:    tags,
	}
	if !plan.ExpirationDate.IsNull() && !plan.ExpirationDate.IsUnknown() {
		createReq.ExpirationDate = plan.ExpirationDate.ValueString()
	}

	created, err := r.client.CreateKey(ctx, plan.ProjectID.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Deepgram API key", err.Error())
		return
	}

	plan.ID = types.StringValue(created.APIKeyID)
	plan.Key = types.StringValue(created.Key)
	plan.Comment = types.StringValue(created.Comment)

	// Deliberately keep plan.Scopes (not created.Scopes) in state.
	// Deepgram returns scopes in a different order and may add extra scopes
	// beyond what was requested, which would cause a post-apply inconsistency error.
	// The plan scopes are what the user declared and what was sent to the API.

	tagList, diags := types.ListValueFrom(ctx, types.StringType, created.Tags)
	resp.Diagnostics.Append(diags...)
	plan.Tags = tagList

	if created.ExpirationDate != "" {
		plan.ExpirationDate = types.StringValue(created.ExpirationDate)
	} else {
		plan.ExpirationDate = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *keyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state keyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	got, err := r.client.GetKey(ctx, state.ProjectID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read Deepgram API key", err.Error())
		return
	}
	if got == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	apiKey := got.Item.Member.APIKey

	state.Comment = types.StringValue(apiKey.Comment)

	scopeSet, diags := types.SetValueFrom(ctx, types.StringType, apiKey.Scopes)
	resp.Diagnostics.Append(diags...)
	state.Scopes = scopeSet

	if len(apiKey.Tags) > 0 {
		tagList, diags := types.ListValueFrom(ctx, types.StringType, apiKey.Tags)
		resp.Diagnostics.Append(diags...)
		state.Tags = tagList
	} else {
		state.Tags = types.ListValueMust(types.StringType, []attr.Value{})
	}

	if apiKey.ExpirationDate != "" {
		state.ExpirationDate = types.StringValue(apiKey.ExpirationDate)
	} else {
		state.ExpirationDate = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is not supported; all attribute changes require replacement.
func (r *keyResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"Deepgram API keys cannot be updated in place. All changes force a replacement.",
	)
}

func (r *keyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state keyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteKey(ctx, state.ProjectID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete Deepgram API key", err.Error())
		return
	}
}
