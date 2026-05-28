package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nais/terraform-provider-fasit/fasit/protogen"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &environmentValuesAcrossEnvs{}

func newEnvironmentValuesAcrossEnvs() datasource.DataSource {
	return &environmentValuesAcrossEnvs{}
}

// environmentValuesAcrossEnvs defines the data source implementation.
type environmentValuesAcrossEnvs struct {
	client protogen.ProviderClient
}

// environmentValuesAcrossEnvsModel describes the data source data model.
type environmentValuesAcrossEnvsModel struct {
	Key     types.String                         `tfsdk:"key"`
	Results []environmentValuesAcrossEnvsResults `tfsdk:"results"`
}

// environmentValuesAcrossEnvsResults describes the results data model.
type environmentValuesAcrossEnvsResults struct {
	Key             types.String `tfsdk:"key"`
	Value           types.String `tfsdk:"value"`
	Secret          types.Bool   `tfsdk:"secret"`
	TenantID        types.String `tfsdk:"tenant_id"`
	TenantName      types.String `tfsdk:"tenant_name"`
	EnvironmentID   types.String `tfsdk:"environment_id"`
	EnvironmentName types.String `tfsdk:"environment_name"`
}

func (d *environmentValuesAcrossEnvs) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment_values_across_envs"
}

func (d *environmentValuesAcrossEnvs) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Get a list of all environment values across all environments for a given key",

		Attributes: map[string]schema.Attribute{
			"key": schema.StringAttribute{
				MarkdownDescription: "The key to get environment values for",
				Required:            true,
			},

			"results": schema.ListNestedAttribute{
				MarkdownDescription: "The environment values for the given key",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							MarkdownDescription: "The key to get environment values for",
							Computed:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: "The value of the environment value. JSON encoded",
							Computed:            true,
						},
						"secret": schema.BoolAttribute{
							MarkdownDescription: "Marks the value as a secret in Fasit. A marked secrets is used for masking, and trigger secret-tainting of computed Helm values. Set to `true` for any sensitive value.",
							Computed:            true,
						},
						"tenant_id": schema.StringAttribute{
							MarkdownDescription: "The ID of the tenant the environment value belongs to",
							Computed:            true,
						},
						"tenant_name": schema.StringAttribute{
							MarkdownDescription: "The name of the tenant the environment value belongs to",
							Computed:            true,
						},
						"environment_id": schema.StringAttribute{
							MarkdownDescription: "The ID of the environment the environment value belongs to",
							Computed:            true,
						},
						"environment_name": schema.StringAttribute{
							MarkdownDescription: "The name of the environment the environment value belongs to",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *environmentValuesAcrossEnvs) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(protogen.ProviderClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *environmentValuesAcrossEnvs) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data environmentValuesAcrossEnvsModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	res, err := d.client.GetEnvironmentValuesAcrossEnvs(ctx, &protogen.GetEnvironmentValuesAcrossEnvsRequest{Key: data.Key.ValueString()})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
		return
	}

	for _, env := range res.GetValues() {
		data.Results = append(data.Results, environmentValuesAcrossEnvsResults{
			Key:             types.StringValue(env.Key),
			Value:           types.StringValue(string(env.GetValue())),
			Secret:          types.BoolValue(env.Secret),
			TenantID:        types.StringValue(env.TenantId),
			TenantName:      types.StringValue(env.TenantName),
			EnvironmentID:   types.StringValue(env.EnvironmentId),
			EnvironmentName: types.StringValue(env.EnvironmentName),
		})
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
