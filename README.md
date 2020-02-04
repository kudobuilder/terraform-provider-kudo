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

This terraform provider looks to solve two primary issues with how KUDO is used to deploy instances:

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

### Install KUDO terraform provider
Because this provider is not currently an official Terraform provider, we need to install it into the local directory that Terraform looks for providers.  From the top level, execute `make build`:

```bash
$ make build
mkdir -p ~/.terraform.d/plugins/
go build -o ~/.terraform.d/plugins/terraform-provider-kudo_v0.0.1 ./kudo
```


### Setup a Kubernetes Environment

Note that none of these Kubernetes clusters are production ready.  For help setting up and configuring production ready Kubernetes clusters, reach out to the KUDO team.

#### Minikube 

If looking to use Minikube, start Minikube with the following configuration:

```bash
minikube config set memory 12000
minikube config set cpus 5
minikube start
cd examples/minikube
```

#### EKS

To setup an EKS cluster on AWS, naviate to the `examples/aws` folder.  Ensure that there is an appropriate profile set in your `~/.aws/config`.  

```bash
$ cat ~/.aws/config
[profile kudo]
aws_access_key_id = *********************
aws_secret_access_key = *********************
region = us-west-2
```

Using this profile you should be able to execute:
```bash
$  aws sts get-caller-identity --profile kudo
{
    "UserId": "AIDAXDPBYCHH42DGNARPM",
    "Account": "************",
    "Arn": "arn:aws:iam::************:user/kudo"
}
```

To enable Terrafrom to pick up this profile, set the following two environment variables, changing the profile name to your preferred AWS profile:

```bash
$ export AWS_PROFILE=kudo
$ export AWS_SDK_LOAD_CONFIG=1 
```

For the following sections, there will be Terraform output specific to setting up and configuring your AWS resources, but the commands should be the same.  Once the resources have been provisioned, configuring your `kubectl` to talk to this cluster requires executing the command

```bash
aws eks update-kubeconfig --name kudo
```

### Application Layout

In this example there are two pieces of software:  Kafka and Zookeeper.  The Kafka instance needs the connection info from the Zookeeper instance to use as its state store.  

* Look at the graph
* Look at the locals to show how we provide the connection info to Kafka

### Plan Installation

Terraform will show the difference between the current state of the environment and what's expected based on the `main.tf` file by running 


This output is an excerpt from the `aws` example:
```bash
$ terraform plan
Refreshing Terraform state in-memory prior to plan...
The refreshed state will be used to calculate this plan, but will not be
persisted to local or remote state storage.

module.kubernetes_cluster.data.aws_availability_zones.available: Refreshing state...
module.kubernetes_cluster.module.eks.data.aws_caller_identity.current: Refreshing state...
module.kubernetes_cluster.module.eks.data.aws_iam_policy_document.workers_assume_role_policy: Refreshing state...
module.kubernetes_cluster.module.eks.data.aws_iam_policy_document.cluster_assume_role_policy: Refreshing state...
module.kubernetes_cluster.module.eks.data.aws_ami.eks_worker: Refreshing state...
module.kubernetes_cluster.module.eks.data.aws_ami.eks_worker_windows: Refreshing state...

------------------------------------------------------------------------

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  + create
 <= read (data resources)

Terraform will perform the following actions:

  # kudo_instance.kafka will be created
  + resource "kudo_instance" "kafka" {
      + cleanup_pvcs               = true
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
      + pvcs                       = (known after apply)
      + services                   = (known after apply)
      + statefulsets               = (known after apply)
    }

  # kudo_instance.zk1 will be created
  + resource "kudo_instance" "zk1" {
      + cleanup_pvcs               = true
      + configmaps                 = (known after apply)
      + deployments                = (known after apply)
      + id                         = (known after apply)
      + name                       = "zook"
      + namespace                  = "default"
      + operator_version_name      = (known after apply)
      + operator_version_namespace = "default"
      + output_parameters          = (known after apply)
      + parameters                 = {
          + "NODE_COUNT" = "5"
        }
      + pods                       = (known after apply)
      + pvcs                       = (known after apply)
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
...
  # module.kubernetes_cluster.module.vpc.aws_vpc.this[0] will be created
  + resource "aws_vpc" "this" {
      + arn                              = (known after apply)
      + assign_generated_ipv6_cidr_block = false
      + cidr_block                       = "10.0.0.0/16"
      + default_network_acl_id           = (known after apply)
      + default_route_table_id           = (known after apply)
      + default_security_group_id        = (known after apply)
      + dhcp_options_id                  = (known after apply)
      + enable_classiclink               = (known after apply)
      + enable_classiclink_dns_support   = (known after apply)
      + enable_dns_hostnames             = true
      + enable_dns_support               = true
      + id                               = (known after apply)
      + instance_tenancy                 = "default"
      + ipv6_association_id              = (known after apply)
      + ipv6_cidr_block                  = (known after apply)
      + main_route_table_id              = (known after apply)
      + owner_id                         = (known after apply)
      + tags                             = {
          + "Name"                       = "test-vpc"
          + "kubernetes.io/cluster/kudo" = "shared"
        }
    }

Plan: 57 to add, 0 to change, 0 to destroy.

------------------------------------------------------------------------

Note: You didn't specify an "-out" parameter to save this plan, so Terraform
can't guarantee that exactly these actions will be performed if
"terraform apply" is subsequently run.
```


The output from just using Minikube includes only the 4 `kudo` resources that will be created.

### Create All the Things!

Terraform `apply` will again show you the changes to the environment, and prompt for the user to execute those changes by typing `yes`. Adding the argument `-auto-approve` will implement these changes without prompting.  Note the creation of the EKS cluster can take over 15 minutes.  So go make a ☕ at this point.

```bash
$ terraform plan -auto-approve
...
kudo_instance.zk1: Still creating... [2m10s elapsed]
kudo_instance.zk1: Still creating... [2m20s elapsed]
kudo_instance.zk1: Still creating... [2m30s elapsed]
kudo_instance.zk1: Still creating... [2m40s elapsed]
kudo_instance.zk1: Still creating... [2m50s elapsed]
kudo_instance.zk1: Still creating... [3m0s elapsed]
kudo_instance.zk1: Still creating... [3m10s elapsed]
kudo_instance.zk1: Still creating... [3m20s elapsed]
kudo_instance.zk1: Still creating... [3m30s elapsed]
kudo_instance.zk1: Still creating... [3m40s elapsed]
kudo_instance.zk1: Creation complete after 3m42s [id=default_zook]
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
kudo_instance.kafka: Creation complete after 2m13s [id=default_pipes]

Apply complete! Resources: 57 added, 0 changed, 0 destroyed.

```

At this point, we created an EKS cluster in AWS, installed KUDO, installed both the Kafka and Zookeeper Operators, created a Zookeeper instance, used that instance to provide a connection string to Kafka and then created the Kafka instance.




### Rollout an update!


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

