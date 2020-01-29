# Terraform Provider for KUDO


## Things to do:

1. Implement the provider configuration.
2. Get boolean for configuring KUDO as part of the configuration
3. Implement the init of KUDO on the call


4. Add all flags to Operator install

5. Look at Whether webhook installs work correctly

6. Add all flags for Instance install


8. Inside of KUDO there's no way to uninstall a framework.  

## Others

- Patch instances
- Outputs for instances.  Can we get POD names owned by an instance for plugging into others?




## KUDO improvements

* KUDO Client improvements
  * CreateInstanceAndWait(timeout)
  * UpdateInstanceAndWait(timeout)
* Consistent labeling of objects



## Blog

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

Installing these objects in the cluster has some obvious downsides.  Firstly, `pipes` gets launched at the same time as `zook` and will thrash unhealthily unnecessiarily until `zook` gets healthy.  

Currently, the instances need to be reasoned about individually to identify the correct objects that would be created, and the proper connection information.  The end user would need to look through the depen

### Setup Kubernetes (GKE? EKS?)

### Install KUDO terraform provider

### Deploy an Operator and Instance

### Outputs of Instance are inputs of Instance

### Cascading Updates

