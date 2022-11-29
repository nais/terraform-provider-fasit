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
	_ tfsdk.ResourceType = fasitEnvironmentResourceType{}
	_ tfsdk.Resource     = fasitEnvironmentResource{}
)

type fasitEnvironmentResourceType struct{}

func (f fasitEnvironmentResourceType) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Environment resource",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				MarkdownDescription: "Environment ID",
				Computed:            true,
				Type:                types.StringType,
			},
			"tenant_id": {
				MarkdownDescription: "Tenant ID",
				Required:            true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.RequiresReplace(),
				},
				Type: types.StringType,
			},
			"name": {
				MarkdownDescription: "Environment name",
				Required:            true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.RequiresReplace(),
				},
				Type: types.StringType,
			},
			"kind": {
				MarkdownDescription: "Environment kind",
				Required:            true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.RequiresReplace(),
				},
				Type: types.StringType,
			},
		},
	}, nil
}

func (f fasitEnvironmentResourceType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return fasitEnvironmentResource{
		provider: provider,
	}, diags
}

type fasitEnvironmentResource struct {
	provider provider
}

type fasitEnvironmentData struct {
	ID       types.String `tfsdk:"id"`
	TenantID types.String `tfsdk:"tenant_id"`
	Name     types.String `tfsdk:"name"`
	Kind     types.String `tfsdk:"kind"`
}

func (f fasitEnvironmentResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	var data fasitEnvironmentData
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	kind := protogen.EnvironmentKind_UNKNOWN
	switch data.Kind.Value {
	case "management":
		kind = protogen.EnvironmentKind_MANAGEMENT
	case "tenant":
		kind = protogen.EnvironmentKind_TENANT
	case "onprem":
		kind = protogen.EnvironmentKind_ONPREM
	case "legacy":
		kind = protogen.EnvironmentKind_LEGACY
	default:
		resp.Diagnostics.AddError("Invalid kind", fmt.Sprintf("Invalid kind: %s", data.Kind.Value))
		return
	}

	res, err := f.provider.client.CreateEnvironment(ctx, &protogen.CreateEnvironmentRequest{
		Name:     data.Name.Value,
		TenantId: data.TenantID.Value,
		Kind:     kind,
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create environment, got error: %s", err))
		return
	}

	data.ID = types.String{Value: res.Id}
	tflog.Trace(ctx, "create environment")

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (f fasitEnvironmentResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var data fasitEnvironmentData

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	res, err := f.provider.client.GetEnvironment(ctx, &protogen.GetEnvironmentRequest{
		TenantId: data.TenantID.Value,
		Name:     data.Name.Value,
	})
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get environment, got error: %s", err))
		return
	}

	data.ID = types.String{Value: res.Id}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (f fasitEnvironmentResource) Update(context.Context, tfsdk.UpdateResourceRequest, *tfsdk.UpdateResourceResponse) {
}

func (f fasitEnvironmentResource) Delete(context.Context, tfsdk.DeleteResourceRequest, *tfsdk.DeleteResourceResponse) {
}
