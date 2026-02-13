package provider

import (
	"context"
	"time"

	"github.com/sagadata-public/sagadata-go"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PrivateNetworkResourceModel struct {
	CreatedAt types.String `tfsdk:"created_at"`

	// CidrV4 The IPv4 CIDR block for the private network.
	CidrV4 types.String `tfsdk:"cidr_v4"`

	// CidrV6 The IPv6 CIDR block for the private network.
	CidrV6 types.String `tfsdk:"cidr_v6"`

	// Description The human-readable description for the private network.
	Description types.String `tfsdk:"description"`

	// Id The unique ID of the private network.
	Id types.String `tfsdk:"id"`

	// Name The human-readable name for the private network.
	Name types.String `tfsdk:"name"`

	// Region The region identifier.
	Region types.String `tfsdk:"region"`

	// Status The private network status.
	Status types.String `tfsdk:"status"`

	UpdatedAt types.String `tfsdk:"updated_at"`

	// Timeouts The resource timeouts
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func (data *PrivateNetworkResourceModel) PopulateFromClientResponse(ctx context.Context, network *sagadata.PrivateNetwork) (diag diag.Diagnostics) {
	data.CreatedAt = types.StringValue(network.CreatedAt.Format(time.RFC3339))
	data.Id = types.StringValue(network.Id)
	data.Name = types.StringValue(network.Name)
	data.Description = types.StringValue(network.Description)
	data.Region = types.StringValue(string(network.Region))
	data.Status = types.StringValue(string(network.Status))
	data.UpdatedAt = types.StringValue(network.UpdatedAt.Format(time.RFC3339))

	if network.CidrV4 != nil {
		data.CidrV4 = types.StringValue(*network.CidrV4)
	} else {
		data.CidrV4 = types.StringNull()
	}

	if network.CidrV6 != nil {
		data.CidrV6 = types.StringValue(*network.CidrV6)
	} else {
		data.CidrV6 = types.StringNull()
	}

	return
}
