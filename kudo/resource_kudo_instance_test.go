package main

import (
	"fmt"
	"strconv"

	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/kudobuilder/kudo/pkg/apis/kudo/v1beta1"
)

func testAccCheckInstancePods(id string, instance *v1beta1.Instance, count int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		i, ok := s.RootModule().Resources[id]
		if !ok {
			return fmt.Errorf("could not find object in state: %s", id)
		}

		podCount, err := strconv.ParseInt(i.Primary.Attributes["pods.#"], 10, 64)
		if err != nil {
			return err
		}
		//check
		if podCount != int64(count) {
			return fmt.Errorf("Expected %v number of pods, saw %v in the state", count, podCount)
		}
		config := testAccProvider.Meta().(Config)
		client := config.GetKubernetesClient()
		for index := 0; index < count; index++ {
			//check pod
			podName := i.Primary.Attributes[fmt.Sprintf("pods.%d", index)]
			// make sure pod
			_, err = client.CoreV1().Pods(i.Primary.Attributes["namespace"]).Get(podName, metav1.GetOptions{})
			if err != nil {
				return err
			}
		}
		//get pods from the state
		// pods := i.
		// Check they exist on the cluster

		return nil
	}
}

func testAccCheckInstanceExists(name, namespace string, i *v1beta1.Instance) resource.TestCheckFunc {
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
		var err error
		i, err = client.GetInstance(name, namespace)
		return err
	}
}

func TestKudoInstance_create(t *testing.T) {
	var instance *v1beta1.Instance

	resource.Test(t, resource.TestCase{
		// PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "kudo_instance.test",
		Providers:     testAccProviders,
		// CheckDestroy:  testAccCheckKudoOperatorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testInstance_zookeeper("zook", 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists("zook", "default", instance),
					// Right number of pods
					// resource.TestCheck
					testAccCheckInstancePods("kudo_instance.test", instance, 1),
					resource.TestCheckResourceAttr("kudo_instance.test", "name", "zook"),
					// correct metadata
					resource.TestCheckResourceAttr("kudo_instance.test", "namespace", "default"),
					resource.TestCheckResourceAttr("kudo_instance.test", "operator_version_name", "zookeeper-0.3.0"),
					resource.TestCheckResourceAttr("kudo_instance.test", "operator_version_namespace", "default"),
					resource.TestCheckResourceAttr("kudo_instance.test", "operator_version_name", "zookeeper-0.3.0"),
					// Parameter values

					// output_parameters

					// testAccCheckMetaAnnotations(&conf.ObjectMeta, map[string]string{"TestAnnotationOne": "one", "TestAnnotationTwo": "two"}),

				),
			},
		},
	})
}

func testInstance_zookeeper(name string, nodes int) string {
	return fmt.Sprintf(`
resource "kudo_operator" "test" {
    operator_name = "zookeeper"
}

resource "kudo_instance" "test" {
	name = "%s"
	parameters = {
		"NODE_COUNT": %d
	}
	operator_version_name = kudo_operator.test.object_name
}
`, name, nodes)
}
