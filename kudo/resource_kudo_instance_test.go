package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

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
			fmt.Printf("Error finding pod count from resource:")
			b, _ := json.MarshalIndent(i, "", "\t")
			fmt.Println(string(b))
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

func TestKudoInstance_label(t *testing.T) {
	//create an instance with the label
	var instance *v1beta1.Instance

	// try a few label sets:
	// no labels
	empty := make(map[string]string)
	// one label
	one := map[string]string{
		"foo": "bar",
	}
	// multiple labels
	two := map[string]string{
		"foo": "bar",
		"baz": "zax",
	}
	// update labels

	//testInstance_label
	resource.Test(t, resource.TestCase{
		// PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "kudo_instance.labels",
		Providers:     testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testInstance_labels("empty", empty),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists("empty", "default", instance),
					resource.TestCheckResourceAttr("kudo_instance.labels", "name", "empty"),
					// has no labels
					testAccCheckInstanceHasLabels("empty", "default", empty),
				),
			},
		},
	})
	//testInstance_label
	resource.Test(t, resource.TestCase{
		// PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "kudo_instance.labels",
		Providers:     testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testInstance_labels("one", one),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists("one", "default", instance),
					resource.TestCheckResourceAttr("kudo_instance.labels", "name", "one"),
					// has no labels
					testAccCheckInstanceHasLabels("one", "default", one),
				),
			},
		},
	})
	//testInstance_label
	resource.Test(t, resource.TestCase{
		// PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "kudo_instance.labels",
		Providers:     testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testInstance_labels("two", two),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists("two", "default", instance),
					resource.TestCheckResourceAttr("kudo_instance.labels", "name", "two"),
					// has no labels
					testAccCheckInstanceHasLabels("two", "default", two),
				),
			},
		},
	})
	//testInstance_label
	resource.Test(t, resource.TestCase{
		// PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "kudo_instance.labels",
		Providers:     testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testInstance_labels("update", one),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists("update", "default", instance),
					resource.TestCheckResourceAttr("kudo_instance.labels", "name", "update"),
					// has no labels
					testAccCheckInstanceHasLabels("update", "default", one),
				),
			},
			{
				Config: testInstance_labels("update", two),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists("update", "default", instance),
					resource.TestCheckResourceAttr("kudo_instance.labels", "name", "update"),
					// has no labels
					testAccCheckInstanceHasLabels("update", "default", two),
				),
			},
		},
	})
}

func testAccCheckInstanceHasLabels(name, namespace string, labels map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := testAccProvider.Meta().(Config)
		client := config.GetKudoClient()
		// name, namespace, err := idParts(rs.Primary.ID)
		// if err != nil {
		// 	return err
		// }
		i, err := client.GetInstance(name, namespace)
		if err != nil {
			return err
		}
		for k, v := range i.Labels {
			l, ok := labels[k]
			if !ok || l != v {
				return fmt.Errorf("Label value for key %v did not match object.  Expected %v, recieved %v", k, l, v)
			}
		}
		for k, v := range labels {
			l, ok := i.Labels[k]
			if !ok || l != v {
				return fmt.Errorf("Label value for key %v did not match object.  Expected %v, recieved %v", k, v, l)
			}
		}

		return err
	}
}

func TestKudoInstance_createRedis(t *testing.T) {
	var instance *v1beta1.Instance

	count := 3

	resource.Test(t, resource.TestCase{
		// PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "kudo_instance.test",
		Providers:     testAccProviders,
		// CheckDestroy:  testAccCheckKudoOperatorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testInstance_redis("redis", count),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckInstanceExists("redis", "default", instance),
					// Right number of pods
					// resource.TestCheck
					testAccCheckInstancePods("kudo_instance.test", instance, 2*count+1),
					resource.TestCheckResourceAttr("kudo_instance.test", "name", "redis"),
					// correct metadata
					resource.TestCheckResourceAttr("kudo_instance.test", "namespace", "default"),
					resource.TestCheckResourceAttr("kudo_instance.test", "operator_version_name", "redis-0.2.0"),
					resource.TestCheckResourceAttr("kudo_instance.test", "operator_version_namespace", "default"),
					resource.TestCheckResourceAttr("kudo_instance.test", "operator_version_name", "redis-0.2.0"),
					// Parameter values

					// output_parameters

					// testAccCheckMetaAnnotations(&conf.ObjectMeta, map[string]string{"TestAnnotationOne": "one", "TestAnnotationTwo": "two"}),

				),
			},
		},
	})
}

func testInstance_labels(name string, labels map[string]string) string {
	s := ""
	if len(labels) > 0 {
		ls := make([]string, 0)
		for k, v := range labels {
			ls = append(ls, fmt.Sprintf("\"%s\" = \"%s\"", k, v))
		}
		s = fmt.Sprintf("labels = { \n %s \n }", strings.Join(ls, "\n"))
	}

	obj := fmt.Sprintf(`
resource "kudo_operator" redis {
	operator_name = "redis"
}	

resource "kudo_instance" labels {
	name = "%s"
	operator_version_name = kudo_operator.redis.object_name
	%s
}
`, name, s)
	fmt.Printf("Resource:\n%s", obj)
	return obj
}

func testInstance_redis(name string, nodes int) string {
	return fmt.Sprintf(`
resource "kudo_operator" "test" {
    operator_name = "redis"
}

resource "kudo_instance" "test" {
	name = "%s"
	parameters = {
		"masters": %d
	}
	operator_version_name = kudo_operator.test.object_name
}
`, name, nodes)
}
