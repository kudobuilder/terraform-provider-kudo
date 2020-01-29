all: build terraform-init

# LOGGING=TF_LOG=trace
LOGGING=

build:
	go build -o terraform/terraform-provider-kudo ./kudo
	cp terraform/terraform-provider-kudo terraform-modules/
	cp terraform/terraform-provider-kudo terraform-modules/kafka-module/
	cp terraform/terraform-provider-kudo terraform-modules/zookeeper-module/

terraform-init: build
	cd terraform; ${LOGGING} terraform init;

terraform-plan: terraform-init
	cd terraform; ${LOGGING} terraform plan

terraform-apply: terraform-plan
	cd terraform; ${LOGGING} terraform apply


kudo-clean:
	kubectl delete ns kudo-system
	kubectl delete crd instances.kudo.dev operators.kudo.dev operatorversions.kudo.dev
