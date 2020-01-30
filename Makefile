all: build terraform-init

# LOGGING=TF_LOG=trace
# LOGGING=

VERSION=0.0.1
BINARY=terraform-provider-kudo_v${VERSION}
PLUGIN_DIR=~/.terraform.d/plugins/

build:
	mkdir -p ${PLUGIN_DIR}
	go build -o ${PLUGIN_DIR}${BINARY} ./kudo

terraform-init: build
	cd terraform; ${LOGGING} terraform init;

terraform-plan: terraform-init
	cd terraform; ${LOGGING} terraform plan

terraform-apply: terraform-plan
	cd terraform; ${LOGGING} terraform apply


kudo-clean:
	kubectl delete ns kudo-system
	kubectl delete crd instances.kudo.dev operators.kudo.dev operatorversions.kudo.dev
