package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/nais/terraform-provider-fasit/fasit/protogen"
)

var _ resource.Resource = &fasitEnvironmentValueResource{}

type fasitEnvironmentValueResource struct {
	client protogen.ProviderClient
}

func newFasitEnvironmentValueResource() resource.Resource {
	return &fasitEnvironmentValueResource{}
}

func (r *fasitEnvironmentValueResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment_value"
}

func (r *fasitEnvironmentValueResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Resource for creating and managing fasit environment values",
		Attributes: map[string]schema.Attribute{
			"environment_id": schema.StringAttribute{
				MarkdownDescription: "Environment ID",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"key": schema.StringAttribute{
				MarkdownDescription: "Key",
				Required:            true,
			},
			"value": schema.StringAttribute{
				MarkdownDescription: "Value",
				Required:            true,
				Sensitive:           true,
			},
			"hide_in_fasit": schema.BoolAttribute{
				MarkdownDescription: "Whether to hide this value in the Fasit UI",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

func (r *fasitEnvironmentValueResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client
}

type fasitEnvironmentValueData struct {
	EnvironmentID types.String `tfsdk:"environment_id"`
	Key           types.String `tfsdk:"key"`
	Value         types.String `tfsdk:"value"`
	HideInFasit   types.Bool   `tfsdk:"hide_in_fasit"`
}

func (f fasitEnvironmentValueResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data fasitEnvironmentValueData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	vb, err := json.Marshal(data.Value.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to marshal value", err.Error())
		return
	}

	_, err = f.client.CreateOrUpdateEnvironmentValue(ctx, &protogen.CreateOrUpdateEnvironmentValueRequest{
		EnvironmentId: data.EnvironmentID.ValueString(),
		Key:           data.Key.ValueString(),
		Value:         vb,
		Secret:        data.HideInFasit.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create EnvironmentValue, got error: %s", err))
		return
	}

	tflog.Trace(ctx, "create EnvironmentValue")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (f fasitEnvironmentValueResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data fasitEnvironmentValueData

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	res, err := f.client.GetEnvironmentValue(ctx, &protogen.GetEnvironmentValueRequest{
		EnvironmentId: data.EnvironmentID.ValueString(),
		Key:           data.Key.ValueString(),
	})
	if err != nil {
		if isNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get EnvironmentValue, got error: %s", err))
		return
	}

	var s string
	err = json.Unmarshal(res.Value, &s)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal value", err.Error())
		return
	}

	data.Value = types.StringValue(s)
	data.HideInFasit = types.BoolValue(res.Secret)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (f fasitEnvironmentValueResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data fasitEnvironmentValueData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	vb, err := json.Marshal(data.Value.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to marshal value", err.Error())
		return
	}

	_, err = f.client.CreateOrUpdateEnvironmentValue(ctx, &protogen.CreateOrUpdateEnvironmentValueRequest{
		EnvironmentId: data.EnvironmentID.ValueString(),
		Key:           data.Key.ValueString(),
		Value:         vb,
		Secret:        data.HideInFasit.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create EnvironmentValue, got error: %s", err))
		return
	}

	tflog.Trace(ctx, "create EnvironmentValue")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (f fasitEnvironmentValueResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data fasitEnvironmentValueData

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := f.client.DeleteEnvironmentValue(ctx, &protogen.DeleteEnvironmentValueRequest{
		EnvironmentId: data.EnvironmentID.ValueString(),
		Key:           data.Key.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get EnvironmentValue, got error: %s", err))
		return
	}
}
