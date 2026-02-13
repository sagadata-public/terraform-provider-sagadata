terraform {
  required_providers {
    sagadata = {
      source = "sagadata-public/sagadata"
      # version = "..."
    }
  }
}

provider "sagadata" {
  # optional configuration...

  # set SAGADATA_TOKEN env var or:
  # token = "..."
}

# Create an instance:

locals {
  region = "NORD-NO-KRS-1"
}

resource "sagadata_ssh_key" "alice" {
  name       = "alice"
  public_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBOpdKM8wSI07+PO4xLDL7zW/kNWGbdFXeHyBU1TRlBn alice@example.com"
}

resource "sagadata_security_group" "allow-ssh" {
  name   = "allow-ssh"
  region = local.region
  rules = [
    {
      direction      = "ingress"
      protocol       = "tcp"
      port_range_min = 22
      port_range_max = 22
    },
  ]
}

resource "sagadata_security_group" "allow-http" {
  name   = "allow-http"
  region = local.region
  rules = [
    {
      direction      = "ingress"
      protocol       = "tcp"
      port_range_min = 80
      port_range_max = 80
    }
  ]
}

resource "sagadata_security_group" "allow-https" {
  name   = "allow-https"
  region = local.region
  rules = [
    {
      direction      = "ingress"
      protocol       = "tcp"
      port_range_min = 443
      port_range_max = 443
    }
  ]
}

resource "sagadata_floating_ip" "floating_ip" {
  name    = "terraform-floating-ip"
  region  = local.region
  version = "ipv4"
}

resource "sagadata_instance" "instance" {
  name   = "terraform-instance"
  region = local.region

  image = "ubuntu-24.04"
  type  = "vcpu-4_memory-16g_nvidia-rtx-3080-1"

  ssh_key_ids = [
    sagadata_ssh_key.alice.id,
  ]

  security_group_ids = [
    sagadata_security_group.allow-ssh.id,
    sagadata_security_group.allow-http.id,
    sagadata_security_group.allow-https.id,
  ]

  floating_ip_id = sagadata_floating_ip.floating_ip.id

  disk_size = 128

  metadata = {
    startup_script = <<EOF
#!/bin/bash
set -eo pipefail

# Add startup script

EOF

  }
}

output "connect" {
  value = "ssh ubuntu@${sagadata_instance.instance.public_ip}"
}
