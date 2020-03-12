module github.com/kudobuilder/terraform-provider-kudo

go 1.13

require (
	github.com/hashicorp/terraform-plugin-sdk v1.8.0
	github.com/kudobuilder/kudo v0.11.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/spf13/afero v1.2.2
	github.com/stretchr/testify v1.5.1
	k8s.io/api v0.16.6
	k8s.io/apiextensions-apiserver v0.0.0-20191016113550-5357c4baaf65
	k8s.io/apimachinery v0.16.6
	k8s.io/client-go v12.0.0+incompatible
// indirect
)

replace k8s.io/client-go => k8s.io/client-go v0.0.0-20191016111102-bec269661e48

replace github.com/Azure/go-autorest => github.com/Azure/go-autorest v12.1.0+incompatible
