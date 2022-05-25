package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/nais/terraform-provider-fasit/fasit/protogen"
)

var (
	_ tfsdk.ResourceType = fasitTenantResourceType{}
	_ tfsdk.Resource     = fasitTenantResource{}
)

type fasitTenantResourceType struct{}

func (f fasitTenantResourceType) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Tenant resource",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				MarkdownDescription: "Tenant ID",
				Computed:            true,
				Type:                types.StringType,
			},
			"name": {
				MarkdownDescription: "Tenant name",
				Required:            true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.RequiresReplace(),
				},
				Type: types.StringType,
			},
		},
	}, nil
}

func (f fasitTenantResourceType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return fasitTenantResource{
		provider: provider,
	}, diags
}

type fasitTenantResource struct {
	provider provider
}

type fasitTenantData struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (f fasitTenantResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var data fasitTenantData
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	res, err := f.provider.client.CreateTenant(ctx, &protogen.CreateTenantRequest{
		Name: data.Name.Value,
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create tenant, got error: %s", err))
		return
	}

	data.ID = types.String{Value: res.Id}
	tflog.Trace(ctx, "create tenant")

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (f fasitTenantResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var data fasitTenantData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	res, err := f.provider.client.GetTenant(ctx, &protogen.GetTenantRequest{
		Name: data.Name.Value,
	})
	if err != nil {
		if isNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get tenant, got error: %s", err))
		return
	}

	data.ID = types.String{Value: res.Id}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (f fasitTenantResource) Update(context.Context, tfsdk.UpdateResourceRequest, *tfsdk.UpdateResourceResponse) {
}

func (f fasitTenantResource) Delete(context.Context, tfsdk.DeleteResourceRequest, *tfsdk.DeleteResourceResponse) {
}
