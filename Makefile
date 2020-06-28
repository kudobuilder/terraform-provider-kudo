all: build terraform-init

# LOGGING=TF_LOG=trace
# LOGGING=

VERSION=0.0.1
BINARY=terraform-provider-kudo_v${VERSION}
PLUGIN_DIR=~/.terraform.d/plugins/

build:
	mkdir -p ${PLUGIN_DIR}
	go build -o ${PLUGIN_DIR}${BINARY} ./kudo

kudo-clean:
	kubectl delete ns kudo-system
	kubectl delete crd instances.kudo.dev operators.kudo.dev operatorversions.kudo.dev
