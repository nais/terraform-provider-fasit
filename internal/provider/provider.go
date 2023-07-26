package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	URL      types.String `provider:"url"`
	Insecure types.Bool   `provider:"insecure"`
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

	gclient, err := grpc.Dial(data.URL.ValueString(), opts...)
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
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "Example provider attribute",
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
		// NewExampleDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &FasitProvider{
			version: version,
		}
	}
}

// convertProviderType is a helper function for NewResource and NewDataSource
// implementations to associate the concrete provider type. Alternatively,
// this helper can be skipped and the provider type can be directly type
// asserted (e.g. provider: in.(*provider)), however using this can prevent
// potential panics.
func convertProviderType(in provider.Provider) (FasitProvider, diag.Diagnostics) {
	var diags diag.Diagnostics

	p, ok := in.(*FasitProvider)

	if !ok {
		diags.AddError(
			"Unexpected Provider Instance Type",
			fmt.Sprintf("While creating the data source or resource, an unexpected provider type (%T) was received. This is always a bug in the provider code and should be reported to the provider developers.", p),
		)
		return FasitProvider{}, diags
	}

	if p == nil {
		diags.AddError(
			"Unexpected Provider Instance Type",
			"While creating the data source or resource, an unexpected empty provider instance was received. This is always a bug in the provider code and should be reported to the provider developers.",
		)
		return FasitProvider{}, diags
	}

	return *p, diags
}

func isNotFound(err error) bool {
	s, ok := status.FromError(err)
	if ok {
		return s.Code() == codes.NotFound
	}
	return false
}
