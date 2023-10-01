resource "kind_cluster" "default" {
  name = "kubelearn-cluster"
  node_image = "kindest/node:v1.27.1"
}