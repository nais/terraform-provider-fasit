package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/nais/terraform-provider-fasit/fasit/protogen"
)

var (
	_ tfsdk.ResourceType = fasitEnvironmentValueResourceType{}
	_ tfsdk.Resource     = fasitEnvironmentValueResource{}
)

type fasitEnvironmentValueResourceType struct{}

func (f fasitEnvironmentValueResourceType) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "EnvironmentValue resource",
		Attributes: map[string]tfsdk.Attribute{
			"environment_id": {
				MarkdownDescription: "Environment ID",
				Required:            true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.RequiresReplace(),
				},
				Type: types.StringType,
			},
			"key": {
				MarkdownDescription: "Key",
				Required:            true,
				Type:                types.StringType,
			},
			"value": {
				MarkdownDescription: "Value",
				Required:            true,
				Type:                types.StringType,
			},
		},
	}, nil
}

func (f fasitEnvironmentValueResourceType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return fasitEnvironmentValueResource{
		provider: provider,
	}, diags
}

type fasitEnvironmentValueResource struct {
	provider provider
}

type fasitEnvironmentValueData struct {
	EnvironmentID types.String `tfsdk:"environment_id"`
	Key           types.String `tfsdk:"key"`
	Value         types.String `tfsdk:"value"`
}

func (f fasitEnvironmentValueResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var data fasitEnvironmentValueData
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	vb, err := json.Marshal(data.Value.Value)
	if err != nil {
		resp.Diagnostics.AddError("Failed to marshal value", err.Error())
		return
	}

	_, err = f.provider.client.CreateOrUpdateEnvironmentValue(ctx, &protogen.CreateOrUpdateEnvironmentValueRequest{
		EnvironmentId: data.EnvironmentID.Value,
		Key:           data.Key.Value,
		Value:         vb,
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create EnvironmentValue, got error: %s", err))
		return
	}

	tflog.Trace(ctx, "create EnvironmentValue")

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (f fasitEnvironmentValueResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var data fasitEnvironmentValueData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	res, err := f.provider.client.GetEnvironmentValue(ctx, &protogen.GetEnvironmentValueRequest{
		EnvironmentId: data.EnvironmentID.Value,
		Key:           data.Key.Value,
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

	data.Value = types.String{Value: s}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (f fasitEnvironmentValueResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var data fasitEnvironmentValueData
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	vb, err := json.Marshal(data.Value.Value)
	if err != nil {
		resp.Diagnostics.AddError("Failed to marshal value", err.Error())
		return
	}

	_, err = f.provider.client.CreateOrUpdateEnvironmentValue(ctx, &protogen.CreateOrUpdateEnvironmentValueRequest{
		EnvironmentId: data.EnvironmentID.Value,
		Key:           data.Key.Value,
		Value:         vb,
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create EnvironmentValue, got error: %s", err))
		return
	}

	tflog.Trace(ctx, "create EnvironmentValue")

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (f fasitEnvironmentValueResource) Delete(context.Context, tfsdk.DeleteResourceRequest, *tfsdk.DeleteResourceResponse) {
}
