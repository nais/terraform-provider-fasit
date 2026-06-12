package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/nais/terraform-provider-fasit/fasit/protogen"
)

var (
	_ resource.Resource                = &fasitEnvironmentResource{}
	_ resource.ResourceWithImportState = &fasitEnvironmentResource{}
)

type fasitEnvironmentResource struct {
	client protogen.FasitClient
}

func newFasitEnvironmentResource() resource.Resource {
	return &fasitEnvironmentResource{}
}

type fasitEnvironmentData struct {
	ID               types.String `tfsdk:"id"`
	TenantID         types.String `tfsdk:"tenant_id"`
	Name             types.String `tfsdk:"name"`
	Kind             types.String `tfsdk:"kind"`
	Labels           types.Map    `tfsdk:"labels"`
	OidcIssuer       types.String `tfsdk:"oidc_issuer"`
	OidcDiscoveryUrl types.String `tfsdk:"oidc_discovery_url"`
}

func (f *fasitEnvironmentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

func (f *fasitEnvironmentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Resource for creating and managing fasit environments",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Tenant ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Environment name",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"tenant_id": schema.StringAttribute{
				MarkdownDescription: "Tenant ID",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"kind": schema.StringAttribute{
				MarkdownDescription: "Environment kind",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"labels": schema.MapAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Environment labels",
				Optional:            true,
			},
			"oidc_issuer": schema.StringAttribute{
				MarkdownDescription: "OIDC issuer for the environment",
				Optional:            true,
			},
			"oidc_discovery_url": schema.StringAttribute{
				MarkdownDescription: "OIDC discovery URL for the environment",
				Optional:            true,
			},
		},
	}
}

func (f *fasitEnvironmentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	f.client = client
}

func (f fasitEnvironmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data fasitEnvironmentData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var kind protogen.EnvironmentKind
	switch data.Kind.ValueString() {
	case "management":
		kind = protogen.EnvironmentKind_MANAGEMENT
	case "tenant":
		kind = protogen.EnvironmentKind_TENANT
	case "onprem":
		kind = protogen.EnvironmentKind_ONPREM
	default:
		resp.Diagnostics.AddAttributeError(path.Root("kind"), "Invalid kind", fmt.Sprintf("Invalid kind: %s", data.Kind.ValueString()))
		return
	}

	labels := labelsToProto(data.Labels)
	res, err := f.client.CreateEnvironment(ctx, &protogen.CreateEnvironmentRequest{
		Name:             data.Name.ValueString(),
		TenantId:         data.TenantID.ValueString(),
		Kind:             kind,
		Labels:           labels,
		OidcIssuer:       data.OidcIssuer.ValueStringPointer(),
		OidcDiscoveryUrl: data.OidcDiscoveryUrl.ValueStringPointer(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create environment, got error: %s", err))
		return
	}

	data.ID = types.StringValue(res.GetEnvironment().GetId())
	data.Labels = labelsFromProto(res.GetEnvironment().GetLabels())
	data.OidcIssuer = stringFromProto(res.GetEnvironment().GetTenantId())
	data.OidcDiscoveryUrl = stringFromProto(res.GetEnvironment().GetOidcDiscoveryUrl())
	data.TenantID = stringFromProto(res.GetEnvironment().GetTenantId())

	tflog.Trace(ctx, "create environment")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (f fasitEnvironmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data fasitEnvironmentData

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	res, err := f.client.GetEnvironment(ctx, &protogen.GetEnvironmentRequest{
		TenantId: data.TenantID.ValueString(),
		Name:     data.Name.ValueString(),
	})
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get environment, got error: %s", err))
		return
	}

	data.ID = types.StringValue(res.GetEnvironment().GetId())
	data.Labels = labelsFromProto(res.GetEnvironment().GetLabels())
	data.OidcIssuer = stringFromProto(res.GetEnvironment().GetOidcIssuer())
	data.OidcDiscoveryUrl = stringFromProto(res.GetEnvironment().GetOidcDiscoveryUrl())
	data.TenantID = stringFromProto(res.GetEnvironment().GetTenantId())
	data.Name = stringFromProto(res.GetEnvironment().GetName())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func labelsFromProto(labels []*protogen.EnvironmentLabel) types.Map {
	ret := map[string]attr.Value{}
	for _, l := range labels {
		ret[l.Key] = types.StringValue(l.Value)
	}
	return types.MapValueMust(types.StringType, ret)
}

func (f fasitEnvironmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var config fasitEnvironmentData
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	var state fasitEnvironmentData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	labels := labelsToProto(config.Labels)
	res, err := f.client.UpdateEnvironment(ctx, &protogen.UpdateEnvironmentRequest{
		EnvironmentId:    state.ID.ValueString(),
		Labels:           labels,
		OidcIssuer:       updateOptionalStringValuePtr(config.OidcIssuer, state.OidcIssuer),
		OidcDiscoveryUrl: updateOptionalStringValuePtr(config.OidcDiscoveryUrl, state.OidcDiscoveryUrl),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update environment, got error: %s", err))
		return
	}

	tflog.Trace(ctx, "update environment")
	state.Labels = labelsFromProto(res.GetEnvironment().GetLabels())
	state.OidcIssuer = stringFromProto(res.GetEnvironment().GetOidcIssuer())
	state.OidcDiscoveryUrl = stringFromProto(res.GetEnvironment().GetOidcDiscoveryUrl())

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func labelsToProto(labels types.Map) []*protogen.EnvironmentLabel {
	if labels.IsNull() || labels.IsUnknown() {
		return nil
	}

	entries := make([]*protogen.EnvironmentLabel, 0, len(labels.Elements()))
	for k, v := range labels.Elements() {
		entries = append(entries, &protogen.EnvironmentLabel{
			Key:   k,
			Value: v.(types.String).ValueString(),
		})
	}
	return entries
}

func stringFromProto(value string) types.String {
	if value == "" {
		return types.StringNull()
	}
	return types.StringValue(value)
}

func updateOptionalStringValuePtr(config types.String, state types.String) *string {
	if config.IsUnknown() {
		return nil
	}

	if config.IsNull() {
		if state.IsNull() || state.IsUnknown() {
			return nil
		}
	}

	return config.ValueStringPointer()
}

func (f fasitEnvironmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("fasit_environment cannot be deleted", "This operation is a no-op")
}

func (f fasitEnvironmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idparts := strings.Split(req.ID, "/")
	if len(idparts) != 3 {
		resp.Diagnostics.AddError("error importing Fasit Environment", "invalid ID specified. Please specify the ID as \"tenant_id/env_name/kind\"")
		return
	}

	res, err := f.client.GetEnvironment(ctx, &protogen.GetEnvironmentRequest{
		TenantId: idparts[0],
		Name:     idparts[1],
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get environment, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx, path.Root("id"), types.StringValue(res.GetEnvironment().GetId()),
	)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx, path.Root("tenant_id"), idparts[0],
	)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx, path.Root("name"), idparts[1],
	)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx, path.Root("kind"), idparts[2],
	)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx, path.Root("labels"), labelsFromProto(res.GetEnvironment().GetLabels()),
	)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx, path.Root("oidc_issuer"), stringFromProto(res.GetEnvironment().GetOidcIssuer()),
	)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(
		ctx, path.Root("oidc_discovery_url"), stringFromProto(res.GetEnvironment().GetOidcDiscoveryUrl()),
	)...)
}
