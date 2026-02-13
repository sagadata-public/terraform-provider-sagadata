package provider

import (
	"context"

	"github.com/sagadata-public/sagadata-go"
	"github.com/sagadata-public/terraform-provider-sagadata/internal/defaultplanmodifier"
	"github.com/sagadata-public/terraform-provider-sagadata/internal/resourceenhancer"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.Resource                     = &PrivateNetworkResource{}
	_ resource.ResourceWithConfigure        = &PrivateNetworkResource{}
	_ resource.ResourceWithImportState      = &PrivateNetworkResource{}
	_ resource.ResourceWithConfigValidators = &PrivateNetworkResource{}
)

func NewPrivateNetworkResource() resource.Resource {
	return &PrivateNetworkResource{}
}

// PrivateNetworkResource defines the resource implementation.
type PrivateNetworkResource struct {
	ResourceWithClient
	ResourceWithTimeout
}

func (r *PrivateNetworkResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_private_network"
}

func (r *PrivateNetworkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Private network resource",

		Attributes: map[string]schema.Attribute{
			"created_at": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The timestamp when this private network was created in RFC 3339.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			}),
			"cidr_v4": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The IPv4 CIDR block for the private network.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			}),
			"cidr_v6": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The IPv6 CIDR block for the private network.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			}),
			"description": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The human-readable description for the private network.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					defaultplanmodifier.String(""),
				},
			}),
			"id": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The unique ID of the private network.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			}),
			"name": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The human-readable name for the private network.",
				Required:            true,
			}),
			"region": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The region identifier.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(sliceStringify(sagadata.AllRegions)...),
				},
			}),
			"status": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The private network status.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			}),
			"updated_at": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The timestamp when this private network was last updated in RFC 3339.",
				Computed:            true,
			}),

			// Internal
			"timeouts": timeouts.AttributesAll(ctx),
		},
	}
}

func (r *PrivateNetworkResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("cidr_v4"),
			path.MatchRoot("cidr_v6"),
		),
	}
}

func (r *PrivateNetworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PrivateNetworkResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, diag := r.ContextWithTimeout(ctx, data.Timeouts.Create)
	if diag != nil {
		resp.Diagnostics.Append(diag...)
		return
	}
	defer cancel()

	body := sagadata.CreatePrivateNetworkJSONRequestBody{}

	body.Name = data.Name.ValueString()
	body.Region = sagadata.Region(data.Region.ValueString())

	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		body.Description = pointer(data.Description.ValueString())
	}

	if !data.CidrV4.IsNull() && !data.CidrV4.IsUnknown() {
		body.CidrV4 = pointer(data.CidrV4.ValueString())
	}

	if !data.CidrV6.IsNull() && !data.CidrV6.IsUnknown() {
		body.CidrV6 = pointer(data.CidrV6.ValueString())
	}

	response, err := r.client.CreatePrivateNetworkWithResponse(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("create private network", err))
		return
	}

	networkResponse := response.JSON201
	if networkResponse == nil {
		resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("create private network", ErrorResponse{
			Body:         response.Body,
			HTTPResponse: response.HTTPResponse,
			Error:        response.JSONDefault,
		}))
		return
	}

	resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, &networkResponse.PrivateNetwork)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "created a private network resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	networkId := networkResponse.PrivateNetwork.Id

	for {
		err := r.client.PollingWait(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Polling Error", generateErrorMessage("polling private network", err))
			return
		}

		tflog.Trace(ctx, "polling a private network resource")

		response, err := r.client.GetPrivateNetworkWithResponse(ctx, networkId)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", generateErrorMessage("polling private network", err))
			return
		}

		networkResponse := response.JSON200
		if networkResponse == nil {
			resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("polling private network", ErrorResponse{
				Body:         response.Body,
				HTTPResponse: response.HTTPResponse,
				Error:        response.JSONDefault,
			}))
			return
		}

		status := networkResponse.PrivateNetwork.Status
		if status == sagadata.PrivateNetworkStatusCreated || status == sagadata.PrivateNetworkStatusError {
			resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, &networkResponse.PrivateNetwork)...)
			if resp.Diagnostics.HasError() {
				return
			}

			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			if resp.Diagnostics.HasError() {
				return
			}

			if status == sagadata.PrivateNetworkStatusError {
				resp.Diagnostics.AddError("Provisioning Error", generateErrorMessage("polling private network", ErrResourceInErrorState))
			}
			return
		}
	}
}

func (r *PrivateNetworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PrivateNetworkResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, diag := r.ContextWithTimeout(ctx, data.Timeouts.Read)
	if diag != nil {
		resp.Diagnostics.Append(diag...)
		return
	}
	defer cancel()

	networkId := data.Id.ValueString()

	response, err := r.client.GetPrivateNetworkWithResponse(ctx, networkId)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("read private network", err))
		return
	}

	networkResponse := response.JSON200
	if networkResponse == nil {
		resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("read private network", ErrorResponse{
			Body:         response.Body,
			HTTPResponse: response.HTTPResponse,
			Error:        response.JSONDefault,
		}))
		return
	}

	resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, &networkResponse.PrivateNetwork)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "read a private network resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PrivateNetworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PrivateNetworkResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, diag := r.ContextWithTimeout(ctx, data.Timeouts.Update)
	if diag != nil {
		resp.Diagnostics.Append(diag...)
		return
	}
	defer cancel()

	body := sagadata.UpdatePrivateNetworkJSONRequestBody{}

	body.Name = pointer(data.Name.ValueString())
	body.Description = pointer(data.Description.ValueString())

	networkId := data.Id.ValueString()

	response, err := r.client.UpdatePrivateNetworkWithResponse(ctx, networkId, body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("update private network", err))
		return
	}

	networkResponse := response.JSON200
	if networkResponse == nil {
		resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("update private network", ErrorResponse{
			Body:         response.Body,
			HTTPResponse: response.HTTPResponse,
			Error:        response.JSONDefault,
		}))
		return
	}

	resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, &networkResponse.PrivateNetwork)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "updated a private network resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PrivateNetworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PrivateNetworkResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, diag := r.ContextWithTimeout(ctx, data.Timeouts.Delete)
	if diag != nil {
		resp.Diagnostics.Append(diag...)
		return
	}
	defer cancel()

	networkId := data.Id.ValueString()

	response, err := r.client.DeletePrivateNetworkWithResponse(ctx, networkId)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("delete private network", err))
		return
	}

	if response.StatusCode() != 204 {
		resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("delete private network", ErrorResponse{
			Body:         response.Body,
			HTTPResponse: response.HTTPResponse,
			Error:        response.JSONDefault,
		}))
		return
	}

	for {
		err := r.client.PollingWait(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Polling Error", generateErrorMessage("polling private network", err))
			return
		}

		tflog.Trace(ctx, "polling a private network resource")

		response, err := r.client.GetPrivateNetworkWithResponse(ctx, networkId)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", generateErrorMessage("polling private network", err))
			return
		}

		if response.StatusCode() == 404 {
			return
		}
	}
}

func (r *PrivateNetworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
