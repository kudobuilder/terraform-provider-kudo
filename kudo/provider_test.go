package main

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	// "github.com/terraform-providers/terraform-provider-kubernetes/kubernetes"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider()
	testAccProvider.Schema["kudo_version"].Default = "0.11.0" //Annoying we have to do this
	testAccProviders = map[string]terraform.ResourceProvider{
		"kudo": testAccProvider,
		// "kubernetes": kubernetes.Provider(),
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_configure(t *testing.T) {
	if os.Getenv("TF_ACC") != "" {
		t.Skip("The environment variable TF_ACC is set, and this test prevents acceptance tests" +
			" from running as it alters environment variables - skipping")
	}

	rc := terraform.NewResourceConfigRaw(map[string]interface{}{})
	p := Provider()
	p.Schema["kudo_version"].Default = "0.11.0" //Annoying we have to do this
	err := p.Configure(rc)
	if err != nil {
		t.Fatal(err)
	}

	//Check to make sure the provider is healthy

}
