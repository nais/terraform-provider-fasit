package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nais/terraform-provider-fasit/fasit/protogen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ provider.Provider = &FasitProvider{}

// FasitProvider satisfies the provider.Provider interface and usually is included
// with all Resource and DataSource implementations.
type FasitProvider struct {
	// client can contain the upstream provider SDK or HTTP client used to
	// communicate with the upstream service. Resource and DataSource
	// implementations can then make calls using this client.
	//
	// TODO: If appropriate, implement upstream provider SDK or HTTP client.
	// client vendorsdk.ExampleClient
	client protogen.ProviderClient

	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// FasitProviderModel can be used to store data from the Terraform configuration.
type FasitProviderModel struct {
	URL      types.String `provider:"url" tfsdk:"url"`
	Insecure types.Bool   `provider:"insecure" tfsdk:"insecure"`
}

func (f *FasitProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data FasitProviderModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	// if data.Example.Null { /* ... */ }

	// If the upstream provider SDK or HTTP client requires configuration, such
	// as authentication or logging, this is a great opportunity to do so.

	if data.URL.IsNull() || data.URL.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("url"), "must be set", "A URL must be set")
		return
	}

	var opts []grpc.DialOption

	if data.Insecure.ValueBool() {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	gclient, err := grpc.NewClient(data.URL.ValueString(), opts...)
	if err != nil {
		resp.Diagnostics.AddError("Failed to connect to provider", err.Error())
		return
	}

	f.client = protogen.NewProviderClient(gclient)

	resp.DataSourceData = f.client
	resp.ResourceData = f.client
}

func (f *FasitProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "fasit"
	resp.Version = f.version
}

func (f *FasitProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				MarkdownDescription: "url",
				Required:            true,
			},
			"insecure": schema.BoolAttribute{
				MarkdownDescription: "insecure",
				Optional:            true,
			},
		},
	}
}

func (f *FasitProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		newFasitTenantResource,
		newFasitEnvironmentResource,
		newFasitEnvironmentValueResource,
	}
}

func (f *FasitProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		newEnvironmentValuesAcrossEnvs,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &FasitProvider{
			version: version,
		}
	}
}

func isNotFound(err error) bool {
	s, ok := status.FromError(err)
	if ok {
		return s.Code() == codes.NotFound
	}
	return false
}
