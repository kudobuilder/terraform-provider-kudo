module github.com/runyontr/terraform-provider-kudo

go 1.13

require (
	github.com/frankban/quicktest v1.4.2 // indirect
	github.com/google/go-cmp v0.3.1
	github.com/gophercloud/gophercloud v0.3.1-0.20190807175045-25a84d593c97 // indirect
	github.com/hashicorp/go-version v1.2.0
	github.com/hashicorp/terraform v0.12.0 // indirect
	github.com/hashicorp/terraform-plugin-sdk v1.5.0
	github.com/hashicorp/vault v1.1.2 // indirect
	github.com/keybase/go-crypto v0.0.0-20190416182011-b785b22cc757 // indirect
	github.com/kudobuilder/kudo v0.10.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pierrec/lz4 v2.3.0+incompatible // indirect
	github.com/robfig/cron v1.2.0
	github.com/spf13/afero v1.2.2
	github.com/ulikunitz/xz v0.5.6 // indirect
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	golang.org/x/tools/gopls v0.2.2 // indirect
	k8s.io/api v0.0.0-20191025225708-5524a3672fbb
	k8s.io/apiextensions-apiserver v0.0.0-20191016113550-5357c4baaf65
	k8s.io/apimachinery v0.0.0-20191028221656-72ed19daf4bb
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/kube-aggregator v0.0.0-20191025230902-aa872b06629d
// indirect
)

replace k8s.io/client-go => k8s.io/client-go v0.0.0-20191016111102-bec269661e48

replace github.com/Azure/go-autorest => github.com/Azure/go-autorest v12.1.0+incompatible
