package provider

import (
	"context"

	"github.com/sagadata-public/sagadata-go"
	"github.com/sagadata-public/terraform-provider-sagadata/internal/resourceenhancer"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.Resource                = &KubernetesClusterResource{}
	_ resource.ResourceWithConfigure   = &KubernetesClusterResource{}
	_ resource.ResourceWithImportState = &KubernetesClusterResource{}
)

func NewKubernetesClusterResource() resource.Resource {
	return &KubernetesClusterResource{}
}

// KubernetesClusterResource defines the resource implementation.
type KubernetesClusterResource struct {
	ResourceWithClient
	ResourceWithTimeout
}

func (r *KubernetesClusterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_cluster"
}

func (r *KubernetesClusterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Kubernetes cluster resource",

		Attributes: map[string]schema.Attribute{
			"created_at": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The timestamp when this Kubernetes cluster was created in RFC 3339.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			}),
			"id": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The unique ID of the Kubernetes cluster.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			}),
			"name": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The human-readable name for the Kubernetes cluster.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			}),
			"network": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The network ID for the cluster (private network ID).",
				Optional:            true,
			}),
			"status": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The Kubernetes cluster status.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			}),
			"updated_at": resourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The timestamp when this Kubernetes cluster was last updated in RFC 3339.",
				Computed:            true,
			}),

			// Internal
			"timeouts": timeouts.AttributesAll(ctx),
		},
	}
}

func (r *KubernetesClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data KubernetesClusterResourceModel

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

	body := sagadata.CreateKubernetesClusterJSONRequestBody{}

	// The API uses Name field but the create body doesn't have it - it uses the network
	// Looking at the API spec, CreateKubernetesClusterJSONBody only has Id and Network
	// The Name comes from somewhere else or is auto-generated
	// Let's use the name as the optional ID field
	body.Id = pointer(data.Name.ValueString())

	if !data.Network.IsNull() && !data.Network.IsUnknown() {
		body.Network = data.Network.ValueStringPointer()
	}

	response, err := r.client.CreateKubernetesClusterWithResponse(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("create kubernetes cluster", err))
		return
	}

	clusterResponse := response.JSON201
	if clusterResponse == nil {
		resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("create kubernetes cluster", ErrorResponse{
			Body:         response.Body,
			HTTPResponse: response.HTTPResponse,
			Error:        response.JSONDefault,
		}))
		return
	}

	resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, &clusterResponse.Cluster)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "created a kubernetes cluster resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterId := clusterResponse.Cluster.Id

	for {
		err := r.client.PollingWait(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Polling Error", generateErrorMessage("polling kubernetes cluster", err))
			return
		}

		tflog.Trace(ctx, "polling a kubernetes cluster resource")

		response, err := r.client.GetKubernetesClusterWithResponse(ctx, clusterId)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", generateErrorMessage("polling kubernetes cluster", err))
			return
		}

		clusterResponse := response.JSON200
		if clusterResponse == nil {
			resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("polling kubernetes cluster", ErrorResponse{
				Body:         response.Body,
				HTTPResponse: response.HTTPResponse,
				Error:        response.JSONDefault,
			}))
			return
		}

		status := clusterResponse.Cluster.Status
		if status == sagadata.KubernetesClusterStatusActive || status == sagadata.KubernetesClusterStatusError {
			resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, &clusterResponse.Cluster)...)
			if resp.Diagnostics.HasError() {
				return
			}

			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			if resp.Diagnostics.HasError() {
				return
			}

			if status == sagadata.KubernetesClusterStatusError {
				resp.Diagnostics.AddError("Provisioning Error", generateErrorMessage("polling kubernetes cluster", ErrResourceInErrorState))
			}
			return
		}
	}
}

func (r *KubernetesClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data KubernetesClusterResourceModel

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

	clusterId := data.Id.ValueString()

	response, err := r.client.GetKubernetesClusterWithResponse(ctx, clusterId)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("read kubernetes cluster", err))
		return
	}

	clusterResponse := response.JSON200
	if clusterResponse == nil {
		resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("read kubernetes cluster", ErrorResponse{
			Body:         response.Body,
			HTTPResponse: response.HTTPResponse,
			Error:        response.JSONDefault,
		}))
		return
	}

	resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, &clusterResponse.Cluster)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "read a kubernetes cluster resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KubernetesClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data KubernetesClusterResourceModel

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

	body := sagadata.UpdateKubernetesClusterJSONRequestBody{}

	if !data.Network.IsNull() && !data.Network.IsUnknown() {
		body.Network = data.Network.ValueStringPointer()
	}

	clusterId := data.Id.ValueString()

	response, err := r.client.UpdateKubernetesClusterWithResponse(ctx, clusterId, body)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("update kubernetes cluster", err))
		return
	}

	clusterResponse := response.JSON200
	if clusterResponse == nil {
		resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("update kubernetes cluster", ErrorResponse{
			Body:         response.Body,
			HTTPResponse: response.HTTPResponse,
			Error:        response.JSONDefault,
		}))
		return
	}

	resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, &clusterResponse.Cluster)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "updated a kubernetes cluster resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KubernetesClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data KubernetesClusterResourceModel

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

	clusterId := data.Id.ValueString()

	response, err := r.client.DeleteKubernetesClusterWithResponse(ctx, clusterId)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("delete kubernetes cluster", err))
		return
	}

	if response.StatusCode() != 204 {
		resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("delete kubernetes cluster", ErrorResponse{
			Body:         response.Body,
			HTTPResponse: response.HTTPResponse,
			Error:        response.JSONDefault,
		}))
		return
	}

	for {
		err := r.client.PollingWait(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Polling Error", generateErrorMessage("polling kubernetes cluster", err))
			return
		}

		tflog.Trace(ctx, "polling a kubernetes cluster resource")

		response, err := r.client.GetKubernetesClusterWithResponse(ctx, clusterId)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", generateErrorMessage("polling kubernetes cluster", err))
			return
		}

		if response.StatusCode() == 404 {
			return
		}
	}
}

func (r *KubernetesClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
