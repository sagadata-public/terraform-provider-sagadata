resource "sagadata_instance" "example" {
  name   = "example"
  region = "NORD-NO-KRS-1"

  image = "ubuntu-24.04"
  type  = "vcpu-2_memory-4g"

  ssh_key_ids = [
    "my-ssh-key-id"
  ]
}

# Example with private network
resource "sagadata_private_network" "app_network" {
  name    = "app-network"
  region  = "NORD-NO-KRS-1"
  cidr_v4 = "10.0.0.0/24"
}

resource "sagadata_instance" "with_private_network" {
  name   = "app-server"
  region = "NORD-NO-KRS-1"

  image = "ubuntu-24.04"
  type  = "vcpu-2_memory-4g"

  private_network_ids = [sagadata_private_network.app_network.id]

  ssh_key_ids = [
    "my-ssh-key-id"
  ]
}
