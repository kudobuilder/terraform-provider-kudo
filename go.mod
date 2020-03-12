module github.com/kudobuilder/terraform-provider-kudo

go 1.13

require (
	github.com/hashicorp/terraform-plugin-sdk v1.5.0
	github.com/kudobuilder/kudo v0.10.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/runyontr/terraform-provider-kudo v0.0.0-20200309145956-9e76a36dd0ae // indirect
	github.com/spf13/afero v1.2.2
	github.com/stretchr/testify v1.4.0
	k8s.io/api v0.0.0-20191025225708-5524a3672fbb
	k8s.io/apiextensions-apiserver v0.0.0-20191016113550-5357c4baaf65
	k8s.io/apimachinery v0.0.0-20191028221656-72ed19daf4bb
	k8s.io/client-go v12.0.0+incompatible
// indirect
)

replace k8s.io/client-go => k8s.io/client-go v0.0.0-20191016111102-bec269661e48

replace github.com/Azure/go-autorest => github.com/Azure/go-autorest v12.1.0+incompatible
