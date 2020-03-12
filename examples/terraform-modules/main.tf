provider "kudo" {
    kubeconfig = "/Users/tom/.kube/config"
    kudo_version = "0.11.0"
}


module "zookeeper" {
    source = "./zookeeper-module"
    name = "zook"
    node_count = 3
}

module "kafka" {
    source = "./kafka-module"
    name = "pipes"
    zookeeper_uri = module.zookeeper.connection_string
}