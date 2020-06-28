module github.com/kudobuilder/terraform-provider-kudo

go 1.14

require (
	github.com/coreos/etcd v3.3.15+incompatible // indirect
	github.com/hashicorp/terraform-plugin-sdk v1.8.0
	github.com/kudobuilder/kudo v0.14.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/spf13/afero v1.2.2
	github.com/stretchr/testify v1.5.1
	k8s.io/api v0.17.3
	k8s.io/apiextensions-apiserver v0.17.2
	k8s.io/apimachinery v0.17.3
	k8s.io/client-go v0.17.3
// indirect
)

replace github.com/Azure/go-autorest => github.com/Azure/go-autorest v12.1.0+incompatible
