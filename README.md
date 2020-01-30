# Terraform Provider for KUDO


## Problem Statement

* KUDO doesn't orchestrate meta-apps (applications composed of KUDO Opeartors)
  * This could be done by creating an Operator who's templates are other instances
* KUDO doesn't have a framework for referencing other `Instances` as parameter values
  * This can't currently be done with KUDO


## Other Products

https://github.com/garden-io/garden/blob/master/docs/dashboard.gif

From a developers perspective: https://garden.io/




## Things to do:

1. Implement the provider configuration.  Currently only works with hardcoded path to KUBECONFIG

2. Get boolean for configuring KUDO as part of the configuration


5. Look at Whether webhook installs work correctly

6. Add all flags for Instance install

8. Implement OV  and O deletions


## KUDO improvements

* KUDO Client improvements
  * CreateInstanceAndWait(timeout)
  * UpdateInstanceAndWait(timeout)
* Consistent labeling of objects


## Blog Post

### Overview of KUDO

This terraform module looks to solve two primary issues with how KUDO is used to deploy instances:

1. Using the outputs of one instance as the input to another.  I.e. Dependencies
2. Rolling upgrades of dependencies so when an instance is updated, its downstream dependees get updated with new values


Consider the example situation of using a `Zookeeper` `Instance` as the state store for a `Kafka` `Instance`:

```yaml
apiVersion: kudo.dev/v1beta1
kind: Instance
metadata:
    name: zook
    namespace: default
spec:
    operatorVersion:
        name: zookeeper-0.3.0
        namespace: default
    parameters:
        NODE_COUNT: "3"
---
apiVersion: kudo.dev/v1beta1
kind: Instance
metadata:
    name: pipes
    namespace: default
spec:
    operatorVersion:
        name: kafka-1.2.0
        namespace: default
    parameters:
        ZOOKEEPER_URI: zook-zookeeper-0.zook-hs:2181,zook-zookeeper-1.zook-hs:2181,zook-zookeeper-2.zook-hs:2181
```

Installing these objects in the cluster has some obvious downsides.  Firstly, `pipes` gets launched at the same time as `zook` and will thrash unhealthily unnecessiarily until `zook` gets healthy.  While these two objects get healthy on initial deployment, updating them both simultaneously may not allow them to ever correctly reconcile.  Secondly, When updating `zook` to have `NODE_COUNT: "5"`, there needs to be user knowledge of how the parameter change in `zook` effects its pod counts, and that `pipes` needs its `ZOOKEEPER_URI` updated when the pod count of `zook` changes. 

Currently, the instances need to be reasoned about individually to identify the correct objects that would be created, and the proper connection information.  The end user would need to look through the depen

### Setup Kubernetes (GKE? EKS?)

```bash
minikube config set memory 12000
minikube config set cpus 5
minikube start
```


### Install KUDO terraform provider

```bash
$ make build
$ cd terraform
$ terraform apply
kudo_operator.zookeeper: Refreshing state... [id=default-zookeeper-0.3.0]
kudo_operator.kafka: Refreshing state... [id=default-kafka-1.2.0]
kudo_instance.zk1: Refreshing state... [id=default_zook]
kudo_instance.kafka: Refreshing state... [id=default_pipes]

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  # kudo_instance.kafka will be created
  + resource "kudo_instance" "kafka" {
      + configmaps                 = (known after apply)
      + deployments                = (known after apply)
      + id                         = (known after apply)
      + name                       = "pipes"
      + namespace                  = "default"
      + operator_version_name      = (known after apply)
      + operator_version_namespace = "default"
      + output_parameters          = (known after apply)
      + parameters                 = (known after apply)
      + pods                       = (known after apply)
      + services                   = (known after apply)
      + statefulsets               = (known after apply)
    }

  # kudo_instance.zk1 will be created
  + resource "kudo_instance" "zk1" {
      + configmaps                 = (known after apply)
      + deployments                = (known after apply)
      + id                         = (known after apply)
      + name                       = "zook"
      + namespace                  = "default"
      + operator_version_name      = (known after apply)
      + operator_version_namespace = "default"
      + output_parameters          = (known after apply)
      + parameters                 = {
          + "NODE_COUNT" = "3"
        }
      + pods                       = (known after apply)
      + services                   = (known after apply)
      + statefulsets               = (known after apply)
    }

  # kudo_operator.kafka will be created
  + resource "kudo_operator" "kafka" {
      + id                 = (known after apply)
      + object_name        = (known after apply)
      + operator_name      = "kafka"
      + operator_namespace = "default"
      + operator_version   = (known after apply)
      + repo               = (known after apply)
      + skip_instance      = true
    }

  # kudo_operator.zookeeper will be created
  + resource "kudo_operator" "zookeeper" {
      + id                 = (known after apply)
      + object_name        = (known after apply)
      + operator_name      = "zookeeper"
      + operator_namespace = "default"
      + operator_version   = (known after apply)
      + repo               = (known after apply)
      + skip_instance      = true
    }

Plan: 4 to add, 0 to change, 0 to destroy.

Do you want to perform these actions?
  Terraform will perform the actions described above.
  Only 'yes' will be accepted to approve.

  Enter a value: yes

kudo_operator.zookeeper: Creating...
kudo_operator.kafka: Creating...
kudo_operator.zookeeper: Creation complete after 1s [id=default-zookeeper-0.3.0]
kudo_operator.kafka: Creation complete after 1s [id=default-kafka-1.2.0]
kudo_instance.zk1: Creating...
kudo_instance.zk1: Still creating... [10s elapsed]
kudo_instance.zk1: Still creating... [20s elapsed]
kudo_instance.zk1: Still creating... [30s elapsed]
kudo_instance.zk1: Still creating... [40s elapsed]
kudo_instance.zk1: Still creating... [50s elapsed]
kudo_instance.zk1: Still creating... [1m0s elapsed]
kudo_instance.zk1: Still creating... [1m10s elapsed]
kudo_instance.zk1: Still creating... [1m20s elapsed]
kudo_instance.zk1: Still creating... [1m30s elapsed]
kudo_instance.zk1: Still creating... [1m40s elapsed]
kudo_instance.zk1: Still creating... [1m50s elapsed]
kudo_instance.zk1: Still creating... [2m0s elapsed]
kudo_instance.zk1: Still creating... [2m10s elapsed]
kudo_instance.zk1: Still creating... [2m20s elapsed]
kudo_instance.zk1: Still creating... [2m30s elapsed]
kudo_instance.zk1: Creation complete after 2m34s [id=default_zook]
kudo_instance.kafka: Creating...
kudo_instance.kafka: Still creating... [10s elapsed]
kudo_instance.kafka: Still creating... [20s elapsed]
kudo_instance.kafka: Still creating... [30s elapsed]
kudo_instance.kafka: Still creating... [40s elapsed]
kudo_instance.kafka: Still creating... [50s elapsed]
kudo_instance.kafka: Still creating... [1m0s elapsed]
kudo_instance.kafka: Still creating... [1m10s elapsed]
kudo_instance.kafka: Still creating... [1m20s elapsed]
kudo_instance.kafka: Still creating... [1m30s elapsed]
kudo_instance.kafka: Still creating... [1m40s elapsed]
kudo_instance.kafka: Still creating... [1m50s elapsed]
kudo_instance.kafka: Still creating... [2m0s elapsed]
kudo_instance.kafka: Still creating... [2m10s elapsed]
kudo_instance.kafka: Creation complete after 2m14s [id=default_pipes]

Apply complete! Resources: 4 added, 0 changed, 0 destroyed.
$ kubectl get instance pipes -o jsonpath="{ .spec.parameters.ZOOKEEPER_URI }"
zook-zookeeper-0.zook-hs:2181,zook-zookeeper-1.zook-hs:2181,zook-zookeeper-2.zook-hs:2181

```

Update `main.tf` to have `NODE_COUNT: 5` and run apply again:

```bash
$ terraform apply -auto-approve
...
Enter a value: yes

kudo_instance.zk1: Modifying... [id=default_zook]
kudo_instance.zk1: Still modifying... [id=default_zook, 10s elapsed]
kudo_instance.zk1: Still modifying... [id=default_zook, 20s elapsed]
kudo_instance.zk1: Still modifying... [id=default_zook, 30s elapsed]
kudo_instance.zk1: Still modifying... [id=default_zook, 40s elapsed]
kudo_instance.zk1: Still modifying... [id=default_zook, 50s elapsed]
kudo_instance.zk1: Still modifying... [id=default_zook, 1m0s elapsed]
kudo_instance.zk1: Still modifying... [id=default_zook, 1m10s elapsed]
kudo_instance.zk1: Still modifying... [id=default_zook, 1m20s elapsed]
kudo_instance.zk1: Still modifying... [id=default_zook, 1m30s elapsed]
kudo_instance.zk1: Still modifying... [id=default_zook, 1m40s elapsed]
kudo_instance.zk1: Still modifying... [id=default_zook, 1m50s elapsed]
kudo_instance.zk1: Still modifying... [id=default_zook, 2m0s elapsed]
kudo_instance.zk1: Still modifying... [id=default_zook, 2m10s elapsed]
kudo_instance.zk1: Still modifying... [id=default_zook, 2m20s elapsed]
kudo_instance.zk1: Still modifying... [id=default_zook, 2m30s elapsed]
kudo_instance.zk1: Still modifying... [id=default_zook, 2m40s elapsed]
kudo_instance.zk1: Still modifying... [id=default_zook, 2m50s elapsed]
kudo_instance.zk1: Still modifying... [id=default_zook, 3m0s elapsed]
kudo_instance.zk1: Modifications complete after 3m2s [id=default_zook]
kudo_instance.kafka: Modifying... [id=default_pipes]
kudo_instance.kafka: Still modifying... [id=default_pipes, 10s elapsed]
kudo_instance.kafka: Still modifying... [id=default_pipes, 20s elapsed]
kudo_instance.kafka: Still modifying... [id=default_pipes, 30s elapsed]
kudo_instance.kafka: Still modifying... [id=default_pipes, 40s elapsed]
kudo_instance.kafka: Still modifying... [id=default_pipes, 50s elapsed]
kudo_instance.kafka: Still modifying... [id=default_pipes, 1m0s elapsed]
kudo_instance.kafka: Modifications complete after 1m3s [id=default_pipes]

Apply complete! Resources: 0 added, 2 changed, 0 destroyed.
$ kubectl get instance pipes -o jsonpath="{ .spec.parameters.ZOOKEEPER_URI }"
zook-zookeeper-0.zook-hs:2181,zook-zookeeper-1.zook-hs:2181,zook-zookeeper-2.zook-hs:2181,zook-zookeeper-3.zook-hs:2181,zook-zookeeper-4.zook-hs:2181
```

### Modules for output variables

