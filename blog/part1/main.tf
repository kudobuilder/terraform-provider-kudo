provider "kudo" {
    # Path to your local kubeconfig
    config_path = "~/.kube/config"
    # Tell the provider to load this config file
    load_config_file = true
    # Version of KUDO to install
    kudo_version = "0.10.0"
}

resource "kudo_operator" "zookeeper" {
    operator_name = "zookeeper"
}

resource "kudo_operator" "kafka" {
    operator_name = "kafka"
}

resource "kudo_instance" "zk1" {
    name = "zook"
    parameters = {
        "NODE_COUNT": 5,
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
