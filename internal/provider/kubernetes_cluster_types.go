package provider

import (
	"context"
	"time"

	"github.com/sagadata-public/sagadata-go"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	resourcetimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type KubernetesClusterResourceModel struct {
	CreatedAt types.String `tfsdk:"created_at"`

	// Id The unique ID of the Kubernetes cluster.
	Id types.String `tfsdk:"id"`

	// Name The human-readable name for the Kubernetes cluster.
	Name types.String `tfsdk:"name"`

	// Network The network ID for the cluster.
	Network types.String `tfsdk:"network"`

	// Status The Kubernetes cluster status.
	Status types.String `tfsdk:"status"`

	UpdatedAt types.String `tfsdk:"updated_at"`

	// Timeouts The resource timeouts
	Timeouts resourcetimeouts.Value `tfsdk:"timeouts"`
}

func (data *KubernetesClusterResourceModel) PopulateFromClientResponse(ctx context.Context, cluster *sagadata.KubernetesCluster) (diag diag.Diagnostics) {
	data.CreatedAt = types.StringValue(cluster.CreatedAt.Format(time.RFC3339))
	data.Id = types.StringValue(cluster.Id)
	data.Name = types.StringValue(cluster.Name)
	data.Status = types.StringValue(string(cluster.Status))
	data.UpdatedAt = types.StringValue(cluster.UpdatedAt.Format(time.RFC3339))

	if cluster.Network != nil {
		data.Network = types.StringValue(*cluster.Network)
	} else {
		data.Network = types.StringNull()
	}

	return
}

type KubernetesClusterDataSourceModel struct {
	// Id The unique ID of the Kubernetes cluster.
	Id types.String `tfsdk:"id"`

	// Name The human-readable name for the Kubernetes cluster.
	Name types.String `tfsdk:"name"`

	// Network The network ID for the cluster.
	Network types.String `tfsdk:"network"`

	// Status The Kubernetes cluster status.
	Status types.String `tfsdk:"status"`

	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`

	// Kubeconfig The kubeconfig for accessing the Kubernetes cluster.
	Kubeconfig types.String `tfsdk:"kubeconfig"`

	// JoinCommand The join command for worker nodes to join the cluster.
	JoinCommand types.String `tfsdk:"join_command"`

	// Timeouts The data source timeouts
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (data *KubernetesClusterDataSourceModel) PopulateFromClientResponse(ctx context.Context, cluster *sagadata.KubernetesCluster) (diag diag.Diagnostics) {
	data.CreatedAt = types.StringValue(cluster.CreatedAt.Format(time.RFC3339))
	data.Id = types.StringValue(cluster.Id)
	data.Name = types.StringValue(cluster.Name)
	data.Status = types.StringValue(string(cluster.Status))
	data.UpdatedAt = types.StringValue(cluster.UpdatedAt.Format(time.RFC3339))

	if cluster.Network != nil {
		data.Network = types.StringValue(*cluster.Network)
	} else {
		data.Network = types.StringNull()
	}

	return
}

func (data *KubernetesClusterDataSourceModel) PopulateCredentialsFromClientResponse(ctx context.Context, creds *sagadata.K8sClusterCredentialsResponse) (diag diag.Diagnostics) {
	data.Kubeconfig = types.StringValue(creds.Kubeconfig)

	if creds.JoinCommand != nil {
		data.JoinCommand = types.StringValue(*creds.JoinCommand)
	} else {
		data.JoinCommand = types.StringNull()
	}

	return
}
