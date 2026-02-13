# Create a private network with IPv4 CIDR
resource "sagadata_private_network" "example" {
  name        = "my-private-network"
  description = "Private network for my infrastructure"
  region      = "NORD-NO-KRS-1"
  cidr_v4     = "10.0.0.0/24"
}

# Create a private network with IPv6 CIDR
resource "sagadata_private_network" "ipv6_network" {
  name        = "ipv6-network"
  description = "IPv6-only private network"
  region      = "NORD-NO-KRS-1"
  cidr_v6     = "fd00::/64"
}
