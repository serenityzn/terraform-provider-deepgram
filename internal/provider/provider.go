package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/serenityzn/terraform-provider-deepgram-/internal/deepgram"
)

var _ provider.Provider = &DeepgramProvider{}

type DeepgramProvider struct {
	version string
}

type DeepgramProviderModel struct {
	APIKey  types.String `tfsdk:"api_key"`
	BaseURL types.String `tfsdk:"base_url"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &DeepgramProvider{version: version}
	}
}

func (p *DeepgramProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "deepgram"
	resp.Version = p.version
}

func (p *DeepgramProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with the Deepgram API to manage project API keys.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Deepgram API key. Can also be set via the DEEPGRAM_API_KEY environment variable.",
			},
			"base_url": schema.StringAttribute{
				Optional:    true,
				Description: "Override the Deepgram API base URL. Defaults to https://api.deepgram.com.",
			},
		},
	}
}

func (p *DeepgramProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config DeepgramProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey := os.Getenv("DEEPGRAM_API_KEY")
	if !config.APIKey.IsNull() && !config.APIKey.IsUnknown() {
		apiKey = config.APIKey.ValueString()
	}

	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing Deepgram API key",
			"Set the api_key provider attribute or the DEEPGRAM_API_KEY environment variable.",
		)
		return
	}

	baseURL := ""
	if !config.BaseURL.IsNull() && !config.BaseURL.IsUnknown() {
		baseURL = config.BaseURL.ValueString()
	}

	client := deepgram.NewClient(apiKey, baseURL)
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *DeepgramProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewKeyResource,
	}
}

func (p *DeepgramProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewKeysDataSource,
	}
}
