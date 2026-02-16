package provider

import (
	"context"

	"github.com/sagadata-public/terraform-provider-sagadata/internal/datasourceenhancer"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ datasource.DataSource              = &KubernetesClusterDataSource{}
	_ datasource.DataSourceWithConfigure = &KubernetesClusterDataSource{}
)

func NewKubernetesClusterDataSource() datasource.DataSource {
	return &KubernetesClusterDataSource{}
}

// KubernetesClusterDataSource defines the data source implementation.
type KubernetesClusterDataSource struct {
	DataSourceWithClient
	DataSourceWithTimeout
}

func (d *KubernetesClusterDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_cluster"
}

func (d *KubernetesClusterDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Kubernetes cluster data source. Fetches cluster details and credentials.",

		Attributes: map[string]schema.Attribute{
			"id": datasourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The unique ID of the Kubernetes cluster.",
				Required:            true,
			}),
			"name": datasourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The human-readable name for the Kubernetes cluster.",
				Computed:            true,
			}),
			"network": datasourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The network ID for the cluster.",
				Computed:            true,
			}),
			"status": datasourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The Kubernetes cluster status.",
				Computed:            true,
			}),
			"created_at": datasourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The timestamp when this Kubernetes cluster was created in RFC 3339.",
				Computed:            true,
			}),
			"updated_at": datasourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The timestamp when this Kubernetes cluster was last updated in RFC 3339.",
				Computed:            true,
			}),
			"kubeconfig": datasourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The kubeconfig for accessing the Kubernetes cluster.",
				Computed:            true,
				Sensitive:           true,
			}),
			"join_command": datasourceenhancer.Attribute(ctx, schema.StringAttribute{
				MarkdownDescription: "The join command for worker nodes to join the cluster.",
				Computed:            true,
				Sensitive:           true,
			}),

			// Internal
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

func (d *KubernetesClusterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data KubernetesClusterDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel, diag := d.ContextWithTimeout(ctx, data.Timeouts.Read)
	if diag != nil {
		resp.Diagnostics.Append(diag...)
		return
	}
	defer cancel()

	clusterId := data.Id.ValueString()

	// Get cluster details
	clusterResponse, err := d.client.GetKubernetesClusterWithResponse(ctx, clusterId)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("read kubernetes cluster", err))
		return
	}

	clusterData := clusterResponse.JSON200
	if clusterData == nil {
		resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("read kubernetes cluster", ErrorResponse{
			Body:         clusterResponse.Body,
			HTTPResponse: clusterResponse.HTTPResponse,
			Error:        clusterResponse.JSONDefault,
		}))
		return
	}

	resp.Diagnostics.Append(data.PopulateFromClientResponse(ctx, &clusterData.Cluster)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get cluster credentials
	credsResponse, err := d.client.GetKubernetesClusterCredentialsWithResponse(ctx, clusterId)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", generateErrorMessage("read kubernetes cluster credentials", err))
		return
	}

	credsData := credsResponse.JSON200
	if credsData == nil {
		resp.Diagnostics.AddError("Client Error", generateClientErrorMessage("read kubernetes cluster credentials", ErrorResponse{
			Body:         credsResponse.Body,
			HTTPResponse: credsResponse.HTTPResponse,
			Error:        credsResponse.JSONDefault,
		}))
		return
	}

	resp.Diagnostics.Append(data.PopulateCredentialsFromClientResponse(ctx, credsData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
