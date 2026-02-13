#!/bin/bash

set -eo pipefail

version="0.0.0-local"
platform="$(go env GOOS)_$(go env GOARCH)"

mkdir -p ~/.terraform.d/plugins/registry.terraform.io/sagadata-public/sagadata/$version/$platform
cp terraform-provider-sagadata ~/.terraform.d/plugins/registry.terraform.io/sagadata-public/sagadata/$version/$platform/terraform-provider-sagadata

# Cleanup:
# rm -fr ~/.terraform.d/plugins/registry.terraform.io/sagadata-public/
