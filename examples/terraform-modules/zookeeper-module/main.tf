provider "kudo" {
    kubeconfig = "/Users/tom/.kube/config"
    kudo_version = "0.14.0"
}

resource "kudo_operator" "zookeeper" {
    operator_name = "zookeeper"
    //hard coded for now
    skip_instance = true
}

resource "kudo_instance" "zk1" {
    name = var.name
    parameters = {
        "NODE_COUNT": var.node_count,
    }
    operator_version_name = kudo_operator.zookeeper.object_name
}


variable "node_count" {}
variable "name" {}

output "connection_string" {
    value = join(",",formatlist("%s.${kudo_instance.zk1.name}-hs:${kudo_instance.zk1.output_parameters.CLIENT_PORT}", kudo_instance.zk1.pods[*]))
}