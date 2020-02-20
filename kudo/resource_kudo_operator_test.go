package main

import (
	"fmt"

	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func testAccCheckOperatorExists(name, namespace string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// _, ok := s.RootModule().Resources[id(name, namespace)]
		// if !ok {
		// 	return fmt.Errorf("Not found: %s/%s", name, namespace)
		// }
		config := testAccProvider.Meta().(Config)
		client := config.GetKudoClient()
		// name, namespace, err := idParts(rs.Primary.ID)
		// if err != nil {
		// 	return err
		// }
		_, err := client.GetOperatorVersion(name, namespace)
		return err
	}
}

func TestKudoOperator_create(t *testing.T) {
	// var conf *v1beta1.OperatorVersion

	resource.Test(t, resource.TestCase{
		// PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "kudo_operator.test",
		Providers:     testAccProviders,
		// CheckDestroy:  testAccCheckKudoOperatorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testOperator_basic("kafka"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckOperatorExists("kafka-1.2.0", "default"),
					resource.TestCheckResourceAttr("kudo_operator.test", "object_name", "kafka-1.2.0"),
					resource.TestCheckResourceAttr("kudo_operator.test", "operator_name", "kafka"),
					resource.TestCheckResourceAttr("kudo_operator.test", "operator_version", "1.2.0"),
					// testAccCheckMetaAnnotations(&conf.ObjectMeta, map[string]string{"TestAnnotationOne": "one", "TestAnnotationTwo": "two"}),
					resource.TestCheckResourceAttr("kudo_operator.test", "id", id("kafka-1.2.0", "default")),
					resource.TestCheckResourceAttr("kudo_operator.test", "skip_instance", "true"),
					resource.TestCheckResourceAttr("kudo_operator.test", "repo", "community"),
				),
			},
		},
	})
}

func TestKudoOperator_update(t *testing.T) {

}

func testOperator_basic(name string) string {
	return fmt.Sprintf(`
resource "kudo_operator" "test" {
    operator_name = "%s"
}
`, name)
}
