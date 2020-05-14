module github.com/kudobuilder/terraform-provider-kudo

go 1.13

require (
	github.com/coreos/etcd v3.3.15+incompatible // indirect
	github.com/hashicorp/terraform-plugin-sdk v1.8.0
	github.com/kudobuilder/kudo v0.13.0-rc1
	github.com/mitchellh/go-homedir v1.1.0
	github.com/spf13/afero v1.2.2
	github.com/stretchr/testify v1.5.1
	k8s.io/api v0.17.3
	k8s.io/apiextensions-apiserver v0.17.2
	k8s.io/apimachinery v0.17.3
	k8s.io/client-go v0.17.3
	k8s.io/code-generator v0.17.3
	k8s.io/component-base v0.17.3
	k8s.io/kubectl v0.17.3
	sigs.k8s.io/controller-runtime v0.5.1
	sigs.k8s.io/controller-tools v0.2.6
	sigs.k8s.io/yaml v1.1.0
// indirect
)

replace github.com/Azure/go-autorest => github.com/Azure/go-autorest v12.1.0+incompatible
