package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/nais/terraform-provider-fasit/fasit/protogen"
)

var (
	_ resource.Resource                 = &fasitEnvironmentValueResource{}
	_ resource.ResourceWithUpgradeState = &fasitEnvironmentValueResource{}
)

type fasitEnvironmentValueResource struct {
	client protogen.FasitClient
}

func newFasitEnvironmentValueResource() resource.Resource {
	return &fasitEnvironmentValueResource{}
}

func (r *fasitEnvironmentValueResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment_value"
}

func (r *fasitEnvironmentValueResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:             2,
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
			},
			"secret": schema.BoolAttribute{
				MarkdownDescription: "Marks the value as a secret in Fasit. A marked secrets is used for masking, and trigger secret-tainting of computed Helm values. Set to `true` for any sensitive value.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		},
	}
}

func (r *fasitEnvironmentValueResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(protogen.FasitClient)

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
	Secret        types.Bool   `tfsdk:"secret"`
}

func (f fasitEnvironmentValueResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data fasitEnvironmentValueData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	vb, err := json.Marshal(data.Value.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to marshal value", err.Error())
		return
	}

	_, err = f.client.SetEnvironmentValue(ctx, &protogen.SetEnvironmentValueRequest{
		EnvironmentId: data.EnvironmentID.ValueString(),
		Key:           data.Key.ValueString(),
		Value:         vb,
		Secret:        data.Secret.ValueBool(),
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
	err = json.Unmarshal(res.GetEnvironmentValue().GetValue(), &s)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal value", err.Error())
		return
	}

	data.Value = types.StringValue(s)
	data.Secret = types.BoolValue(res.GetEnvironmentValue().GetSecret())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (f fasitEnvironmentValueResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data fasitEnvironmentValueData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	vb, err := json.Marshal(data.Value.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to marshal value", err.Error())
		return
	}

	_, err = f.client.SetEnvironmentValue(ctx, &protogen.SetEnvironmentValueRequest{
		EnvironmentId: data.EnvironmentID.ValueString(),
		Key:           data.Key.ValueString(),
		Value:         vb,
		Secret:        data.Secret.ValueBool(),
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

// fasitEnvironmentValueDataV0 represents the v0 state schema where the field was called "secret".
type fasitEnvironmentValueDataV0 struct {
	EnvironmentID types.String `tfsdk:"environment_id"`
	Key           types.String `tfsdk:"key"`
	Value         types.String `tfsdk:"value"`
	Secret        types.Bool   `tfsdk:"secret"`
}

// fasitEnvironmentValueDataV1 represents the v1 state schema where the field was called "hide_in_fasit".
type fasitEnvironmentValueDataV1 struct {
	EnvironmentID types.String `tfsdk:"environment_id"`
	Key           types.String `tfsdk:"key"`
	Value         types.String `tfsdk:"value"`
	HideInFasit   types.Bool   `tfsdk:"hide_in_fasit"`
}

func (f fasitEnvironmentValueResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		// Upgrade from v0 (secret) to v2 (secret) — field name unchanged, schema version bump only
		0: {
			PriorSchema: &schema.Schema{
				Attributes: map[string]schema.Attribute{
					"environment_id": schema.StringAttribute{
						Required: true,
					},
					"key": schema.StringAttribute{
						Required: true,
					},
					"value": schema.StringAttribute{
						Required: true,
					},
					"secret": schema.BoolAttribute{
						Optional: true,
						Computed: true,
					},
				},
			},
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var prior fasitEnvironmentValueDataV0
				resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
				if resp.Diagnostics.HasError() {
					return
				}

				upgraded := fasitEnvironmentValueData(prior)
				resp.Diagnostics.Append(resp.State.Set(ctx, &upgraded)...)
			},
		},
		// Upgrade from v1 (hide_in_fasit) to v2 (secret)
		1: {
			PriorSchema: &schema.Schema{
				Attributes: map[string]schema.Attribute{
					"environment_id": schema.StringAttribute{
						Required: true,
					},
					"key": schema.StringAttribute{
						Required: true,
					},
					"value": schema.StringAttribute{
						Required: true,
					},
					"hide_in_fasit": schema.BoolAttribute{
						Optional: true,
						Computed: true,
					},
				},
			},
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var prior fasitEnvironmentValueDataV1
				resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
				if resp.Diagnostics.HasError() {
					return
				}

				upgraded := fasitEnvironmentValueData{
					EnvironmentID: prior.EnvironmentID,
					Key:           prior.Key,
					Value:         prior.Value,
					Secret:        prior.HideInFasit,
				}
				resp.Diagnostics.Append(resp.State.Set(ctx, &upgraded)...)
			},
		},
	}
}
