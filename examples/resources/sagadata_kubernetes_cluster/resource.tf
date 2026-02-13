# Create a private network for the Kubernetes cluster
resource "sagadata_private_network" "cluster_network" {
  name    = "k8s-network"
  region  = "NORD-NO-KRS-1"
  cidr_v4 = "10.1.0.0/24"
}

# Create a Kubernetes cluster with a private network
resource "sagadata_kubernetes_cluster" "example" {
  name    = "my-k8s-cluster"
  network = sagadata_private_network.cluster_network.id
}

# Access cluster credentials via the data source
data "sagadata_kubernetes_cluster" "example" {
  id = sagadata_kubernetes_cluster.example.id
}

# Output the kubeconfig (mark as sensitive)
output "kubeconfig" {
  value     = data.sagadata_kubernetes_cluster.example.kubeconfig
  sensitive = true
}
