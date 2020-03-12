provider "kudo" {
        kubeconfig = "/Users/tom/.kube/config"
    kudo_version = "0.11.0"
}


resource "kudo_operator" "kafka" {
    operator_name = "kafka"
    skip_instance = true
}

variable "zookeeper_uri" {}
variable "name" {}

resource "kudo_instance" "kafka" {
    name = var.name
    parameters = {
        "ZOOKEEPER_URI": var.zookeeper_uri
    }
    operator_version_name = kudo_operator.kafka.object_name
}
