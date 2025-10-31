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
	client protogen.ProviderClient
}

func newFasitEnvironmentResource() resource.Resource {
	return &fasitEnvironmentResource{}
}

type fasitEnvironmentData struct {
	ID       types.String `tfsdk:"id"`
	TenantID types.String `tfsdk:"tenant_id"`
	Name     types.String `tfsdk:"name"`
	Kind     types.String `tfsdk:"kind"`
	Labels   types.Map    `tfsdk:"labels"`
}

func (r *fasitEnvironmentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

func (r *fasitEnvironmentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
		},
	}
}

func (r *fasitEnvironmentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (f fasitEnvironmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data fasitEnvironmentData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	kind := protogen.EnvironmentKind_UNKNOWN
	switch data.Kind.ValueString() {
	case "management":
		kind = protogen.EnvironmentKind_MANAGEMENT
	case "tenant":
		kind = protogen.EnvironmentKind_TENANT
	case "onprem":
		kind = protogen.EnvironmentKind_ONPREM
	case "legacy":
		kind = protogen.EnvironmentKind_LEGACY
	default:
		resp.Diagnostics.AddAttributeError(path.Root("kind"), "Invalid kind", fmt.Sprintf("Invalid kind: %s", data.Kind.ValueString()))
		return
	}

	labels := labelsToProto(data.Labels)
	res, err := f.client.CreateEnvironment(ctx, &protogen.CreateEnvironmentRequest{
		Name:     data.Name.ValueString(),
		TenantId: data.TenantID.ValueString(),
		Kind:     kind,
		Labels:   labels,
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create environment, got error: %s", err))
		return
	}

	data.ID = types.StringValue(res.Id)
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

	data.ID = types.StringValue(res.Id)

	if res.Labels != nil {
		data.Labels = labelsFromProto(res.Labels)
	}
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
	_, err := f.client.UpdateEnvironment(ctx, &protogen.UpdateEnvironmentRequest{
		EnvironmentId: state.ID.ValueString(),
		Labels:        labels,
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update environment, got error: %s", err))
		return
	}

	tflog.Trace(ctx, "update environment")
	state.Labels = config.Labels

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func labelsToProto(labels types.Map) []*protogen.EnvironmentLabel {
	entries := make([]*protogen.EnvironmentLabel, 0)
	for k, v := range labels.Elements() {
		entries = append(entries, &protogen.EnvironmentLabel{
			Key:   k,
			Value: v.(types.String).ValueString(),
		})
	}
	return entries
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
		ctx, path.Root("id"), types.StringValue(res.Id),
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
		ctx, path.Root("labels"), res.Labels,
	)...)
}
