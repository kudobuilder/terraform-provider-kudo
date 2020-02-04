
provider "kudo" {
    kudo_version = "0.10.0"
    token = module.kubernetes_cluster.token
    host = module.kubernetes_cluster.cluster_endpoint
    cluster_ca_certificate = module.kubernetes_cluster.cluster_ca_certificate
    load_config_file = false
}


module "kubernetes_cluster" {
    source = "./eks/"
}

resource "kudo_operator" "zookeeper" {
    operator_name = "zookeeper"
    skip_instance = true
}

resource "kudo_operator" "kafka" {
    operator_name = "kafka"
    skip_instance = true
}

resource "kudo_instance" "zk1" {
    name = "zook"
    parameters = {
        "NODE_COUNT": 3,
    }
    operator_version_name = kudo_operator.zookeeper.object_name
}

locals {
    zkConnection = join(",",formatlist("%s.${kudo_instance.zk1.name}-hs:${kudo_instance.zk1.output_parameters.CLIENT_PORT}", kudo_instance.zk1.pods[*]))
}

resource "kudo_instance" "kafka" {
    name = "pipes"
    parameters = {
        "ZOOKEEPER_URI": local.zkConnection
    }
    operator_version_name = kudo_operator.kafka.object_name
}
