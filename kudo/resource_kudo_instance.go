package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/kudobuilder/kudo/pkg/apis/kudo/v1beta1"
	"github.com/kudobuilder/kudo/pkg/client/clientset/versioned"
)

//TODO add all parameter values to the state, even defaults

func resourceInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceInstanceCreate,
		Read:   resourceInstanceRead,
		Update: resourceInstanceUpdate,
		Delete: resourceInstanceDelete,
		Exists: resourceInstanceExists,
		//Flag these parameters as "newly computed" if any of the parameters change
		CustomizeDiff: customdiff.All(
			customdiff.ComputedIf("pods", func(d *schema.ResourceDiff, meta interface{}) bool {
				return d.HasChange("parameters")
			}),
			customdiff.ComputedIf("services", func(d *schema.ResourceDiff, meta interface{}) bool {
				return d.HasChange("parameters")
			}),
			customdiff.ComputedIf("statefulsets", func(d *schema.ResourceDiff, meta interface{}) bool {
				return d.HasChange("parameters")
			}),
			customdiff.ComputedIf("deployments", func(d *schema.ResourceDiff, meta interface{}) bool {
				return d.HasChange("parameters")
			}),
			customdiff.ComputedIf("configmaps", func(d *schema.ResourceDiff, meta interface{}) bool {
				return d.HasChange("parameters")
			}),
			customdiff.ComputedIf("output_parameters", func(d *schema.ResourceDiff, meta interface{}) bool {
				return d.HasChange("parameters")
			}),
		),

		//customdiff.ComputedIf("version", func(d *schema.ResourceDiff, meta interface{}) bool {
		//	return d.HasChange("content") || d.HasChange("content_type")
		//},
		Schema: map[string]*schema.Schema{
			"parameters": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
			},
			"output_parameters": &schema.Schema{
				Type:     schema.TypeMap,
				Computed: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"namespace": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "default",
			},
			"labels": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
			},
			"operator_version_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"operator_version_namespace": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "default",
			},
			"pods": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"services": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"deployments": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"statefulsets": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"configmaps": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"pvcs": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"cleanup_pvcs": &schema.Schema{
				Type:        schema.TypeBool,
				Default:     true,
				Description: "If true, deleting the object in terraform will cleanup StatefulSet PVCs",
				Optional:    true,
			},
		},
	}
}

// type CustomizeDiffFunc func(*ResourceDiff, interface{}) error
func customizeInstanceDiff(diff *schema.ResourceDiff, m interface{}) error {

	return nil
}

func resourceInstanceExists(d *schema.ResourceData, m interface{}) (bool, error) {
	config := m.(Config)

	client := config.GetKudoClient()

	_, err := client.GetInstance(d.Get("name").(string), d.Get("namespace").(string))

	return err == nil, err
}

func resourceInstanceCreate(d *schema.ResourceData, m interface{}) error {
	log.Printf("resourceInstanceCreate: %v %v\n", d, m)
	name := d.Get("name").(string)
	namespace := d.Get("namespace").(string)
	operatorVersionName := d.Get("operator_version_name").(string)
	operatorVersionNamespace := namespace
	//Default to instance namespace
	if ovnamespace, ok := d.GetOk("operator_version_namespace"); ok {
		operatorVersionNamespace = ovnamespace.(string)
	} else {
		d.Set("operator_version_namespace", operatorVersionNamespace)
	}
	//
	labels := make(map[string]string)
	inLabels, ok := d.GetOk("labels")
	if ok {
		for k, v := range inLabels.(map[string]interface{}) {
			s, o := v.(string)
			if o {
				labels[k] = s
			}
		}
	}

	// operatorVersionNamespace := d.Get("operator_version_namespace").(string)
	parametersI := d.Get("parameters").(map[string]interface{})
	parameters := make(map[string]string)
	for k, v := range parametersI {
		parameters[k] = v.(string)
	}

	config := m.(Config)
	kudoClient := config.GetKudoClient()

	instance := &v1beta1.Instance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: v1beta1.InstanceSpec{
			Parameters: parameters,
			OperatorVersion: corev1.ObjectReference{
				Namespace: operatorVersionNamespace,
				Name:      operatorVersionName,
			},
		},
	}

	instance, err := kudoClient.InstallInstanceObjToCluster(instance, instance.Namespace)
	if err != nil {
		return fmt.Errorf("Error installing instance: %v", err)
	}

	d.SetId(fmt.Sprintf("%v_%v", namespace, name))

	time.Sleep(3 * time.Second)
	//Wait for plan to run
	return waitForInstance(d, m, name, namespace, nil)

}

//https://www.terraform.io/docs/extend/writing-custom-providers.html#implementing-read
func resourceInstanceRead(d *schema.ResourceData, m interface{}) error {
	log.Printf("resourceInstanceCreate: %v %v\n", d, m)
	name := d.Get("name").(string)
	namespace := d.Get("namespace").(string)
	operatorVersionName := d.Get("operator_version_name").(string)
	operatorVersionNamespace := namespace
	//Default to instance namespace
	if ovnamespace, ok := d.GetOk("operator_version_namespace"); ok {
		operatorVersionNamespace = ovnamespace.(string)
	} else {
		d.Set("operator_version_namespace", operatorVersionNamespace)
	}

	// operatorVersionNamespace := d.Get("operator_version_namespace").(string)
	parametersI := d.Get("parameters").(map[string]interface{})
	parameters := make(map[string]string)
	for k, v := range parametersI {
		parameters[k] = v.(string)
	}

	config := m.(Config)

	//if cluster is not present, mark as "need to install"
	kudoClient := config.GetKudoClient()

	instance, err := kudoClient.GetInstance(name, namespace)
	if err != nil {
		d.SetId("")
		return nil
		// return fmt.Errorf("Error getting instance: %w", err)
	}
	if instance == nil {
		d.SetId("")
		return nil //not present
	}
	labels := make(map[string]string)
	for k, v := range instance.Labels {
		labels[k] = v
	}
	d.Set("labels", labels)

	operatorVersionName = instance.Spec.OperatorVersion.Name
	operatorVersionNamespace = instance.Spec.OperatorVersion.Namespace

	ov, err := kudoClient.GetOperatorVersion(operatorVersionName, operatorVersionNamespace)
	if err != nil {
		return fmt.Errorf("could not get OperatorVersion: %w", err)
	}

	if ov == nil {
		return fmt.Errorf("Could not find OV %v/%v: %v", operatorVersionNamespace, operatorVersionName, err)
	}

	fmt.Println("------------")

	//Set defaults from
	for _, param := range ov.Spec.Parameters {

		if param.Default != nil {
			parameters[param.Name] = *param.Default
		}
	}

	inParams := make(map[string]string)

	//Set things from the instance
	for k, v := range instance.Spec.Parameters {
		parameters[k] = v
		inParams[k] = v
	}

	d.Set("parameters", inParams)
	d.Set("output_parameters", parameters)
	d.Set("operator_version_name", operatorVersionName)
	d.Set("operator_version_namespace", operatorVersionNamespace)

	//Query resources!

	d.SetId(fmt.Sprintf("%v_%v", namespace, name))
	// read the instance from the server and see what's changed.

	// Cluster Resource

	kubeClient := config.GetKubernetesClient()

	// the two common ways objects seem to be labeled
	labelSelector1 := fmt.Sprintf("instance=%s", name)
	labelSelector2 := fmt.Sprintf("kudo.dev/instance=%s", name)

	listOptions1 := metav1.ListOptions{
		LabelSelector: labelSelector1,
		Limit:         100,
	}
	listOptions2 := metav1.ListOptions{
		LabelSelector: labelSelector2,
		Limit:         100,
	}

	//Pods
	podNames := make([]string, 0)
	log.Printf("Searching for Pods in namespace=%v\n", namespace)
	//Get pods for instance (with label instance=name)
	pods, err := kubeClient.CoreV1().Pods(namespace).List(listOptions1)
	if err != nil {
		return fmt.Errorf("Error getting pods: %v", err)
	}
	log.Printf("Found %v pods with label instance=%v:\n", len(pods.Items), name)
	for _, p := range pods.Items {
		log.Printf("Pod: %v\n", p.Name)
		podNames = append(podNames, p.Name)
	}
	pods, err = kubeClient.CoreV1().Pods(namespace).List(listOptions2)
	log.Printf("Found %v pods with label kudo.dev/instance=%v:\n", len(pods.Items), name)
	if err != nil {
		return fmt.Errorf("Error getting pods: %v", err)
	}
	for _, p := range pods.Items {
		log.Printf("Pod: %v\n", p.Name)
		podNames = append(podNames, p.Name)
	}
	d.Set("pods", deduplicate(podNames))

	//Services
	serviceNames := make([]string, 0)

	svcs, err := kubeClient.CoreV1().Services(namespace).List(listOptions1)
	if err != nil {
		return fmt.Errorf("Error getting services: %v", err)
	}
	for _, svc := range svcs.Items {
		serviceNames = append(serviceNames, svc.Name)
	}
	svcs, err = kubeClient.CoreV1().Services(namespace).List(listOptions2)
	if err != nil {
		return fmt.Errorf("Error getting services: %v", err)
	}
	for _, svc := range svcs.Items {
		serviceNames = append(serviceNames, svc.Name)
	}

	d.Set("services", deduplicate(serviceNames))

	//Deployments
	deployNames := make([]string, 0)

	deploys, err := kubeClient.AppsV1().Deployments(namespace).List(listOptions1)
	if err != nil {
		return fmt.Errorf("Error getting deployments: %v", err)
	}
	for _, deploy := range deploys.Items {
		deployNames = append(deployNames, deploy.Name)
	}
	deploys, err = kubeClient.AppsV1().Deployments(namespace).List(listOptions2)
	if err != nil {
		return fmt.Errorf("Error getting deployments: %v", err)
	}
	for _, deploy := range deploys.Items {
		deployNames = append(deployNames, deploy.Name)
	}

	d.Set("deployments", deduplicate(deployNames))

	//ConfigMaps
	cmNames := make([]string, 0)

	cms, err := kubeClient.CoreV1().ConfigMaps(namespace).List(listOptions1)
	if err != nil {
		return fmt.Errorf("Error getting configmaps: %v", err)
	}
	for _, o := range cms.Items {
		cmNames = append(cmNames, o.Name)
	}
	cms, err = kubeClient.CoreV1().ConfigMaps(namespace).List(listOptions2)
	if err != nil {
		return fmt.Errorf("Error getting configmaps: %v", err)
	}
	for _, o := range cms.Items {
		cmNames = append(cmNames, o.Name)
	}

	d.Set("configmaps", deduplicate(cmNames))

	//StatefulSets
	ssNames := make([]string, 0)

	sss, err := kubeClient.AppsV1().StatefulSets(namespace).List(listOptions1)
	if err != nil {
		return fmt.Errorf("Error getting statefulSets: %v", err)
	}
	for _, o := range sss.Items {
		ssNames = append(ssNames, o.Name)
	}
	sss, err = kubeClient.AppsV1().StatefulSets(namespace).List(listOptions2)
	if err != nil {
		return fmt.Errorf("Error getting statefulSets: %v", err)
	}
	for _, o := range sss.Items {
		ssNames = append(ssNames, o.Name)
	}

	d.Set("statefulsets", deduplicate(ssNames))

	//PVCs
	pvcNames := make([]string, 0)

	pvcs, err := kubeClient.CoreV1().PersistentVolumeClaims(namespace).List(listOptions1)
	if err != nil {
		return fmt.Errorf("Error getting pvcs: %v", err)
	}
	for _, o := range pvcs.Items {
		pvcNames = append(pvcNames, o.Name)
	}
	pvcs, err = kubeClient.CoreV1().PersistentVolumeClaims(namespace).List(listOptions2)
	if err != nil {
		return fmt.Errorf("Error getting statefulSets: %v", err)
	}
	for _, o := range pvcs.Items {
		pvcNames = append(pvcNames, o.Name)
	}

	d.Set("pvcs", deduplicate(pvcNames))

	return nil
}

func deduplicate(array []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range array {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func resourceInstanceUpdate(d *schema.ResourceData, m interface{}) error {
	log.Printf("resourceInstanceCreate: %v %v\n", d, m)
	name := d.Get("name").(string)
	namespace := d.Get("namespace").(string)
	operatorVersionName := d.Get("operator_version_name").(string)
	operatorVersionNamespace := namespace
	//Default to instance namespace
	if ovnamespace, ok := d.GetOk("operator_version_namespace"); ok {
		operatorVersionNamespace = ovnamespace.(string)
	} else {
		d.Set("operator_version_namespace", operatorVersionNamespace)
	}

	// operatorVersionNamespace := d.Get("operator_version_namespace").(string)
	parametersI := d.Get("parameters").(map[string]interface{})
	parameters := make(map[string]string)
	for k, v := range parametersI {
		parameters[k] = v.(string)
	}

	config := m.(Config)
	kudoClient := config.GetKudoClient()

	old, err := kudoClient.GetInstance(name, namespace)
	if err != nil {
		return fmt.Errorf("Calling Update on non-existant Instance: %v", err)
	}

	same := true
	for k := range old.Spec.Parameters {
		same = same && old.Spec.Parameters[k] == parameters[k]
	}
	for k := range parameters {
		same = same && parameters[k] == old.Spec.Parameters[k]
	}

	//
	labels := make(map[string]string)
	inLabels, ok := d.GetOk("labels")
	if ok {

		for k, v := range inLabels.(map[string]interface{}) {
			s, o := v.(string)
			if o {
				labels[k] = s
			}
		}
	}
	newPlan := !same

	//check labels
	for k, v := range old.Labels {
		same = same && labels[k] == v
	}
	for k, v := range labels {
		same = same && old.Labels[k] == v
	}

	if same {
		//everything was the same, so don't actually update
		return resourceInstanceRead(d, m)
	}
	err = patchInstance(config.RawKudoClient, name, namespace, parameters, labels, operatorVersionName)
	// err = kudoClient.UpdateInstance(name, namespace, &operatorVersionName, parameters)

	if err != nil {
		return fmt.Errorf("Error updating instance: %v", err)
	}
	if newPlan { // change in parameters trigger a new plan
		return waitForInstance(d, m, name, namespace, old)
	}
	return resourceInstanceRead(d, m)

}

func patchInstance(c *versioned.Clientset, instanceName, namespace string, parameters, labels map[string]string, ovName string) error {
	instanceSpec := v1beta1.InstanceSpec{}
	metadata := metav1.ObjectMeta{}
	if parameters != nil {
		instanceSpec.Parameters = parameters
	}
	instanceSpec.OperatorVersion.Name = ovName
	if labels != nil {
		metadata.Labels = labels
	}
	serializedPatch, err := json.Marshal(struct {
		Spec     *v1beta1.InstanceSpec `json:"spec"`
		Metadata *metav1.ObjectMeta    `json:"metadata"`
	}{
		&instanceSpec,
		&metadata,
	})
	if err != nil {
		return err
	}

	_, err = c.KudoV1beta1().Instances(namespace).Patch(instanceName, types.MergePatchType, serializedPatch)
	return err
}

func waitForInstance(d *schema.ResourceData, m interface{}, name, namespace string, oldInstance *v1beta1.Instance) error {
	//Wait for status plan to be done
	config := m.(Config)
	kudoClient := config.GetKudoClient()
	err := kudoClient.WaitForInstance(name, namespace, oldInstance, time.Second*300)
	if err != nil {
		return err
	}
	return resourceInstanceRead(d, m)
}

func resourceInstanceDelete(d *schema.ResourceData, m interface{}) error {
	log.Printf("resourceInstanceCreate: %v %v\n", d, m)

	name := d.Get("name").(string)
	namespace := d.Get("namespace").(string)
	config := m.(Config)

	kudoClientset := config.RawKudoClient

	propagationPolicy := metav1.DeletePropagationForeground
	options := &metav1.DeleteOptions{
		PropagationPolicy: &propagationPolicy,
	}

	err := kudoClientset.KudoV1beta1().Instances(namespace).Delete(name, options)
	if err != nil {
		return err
	}

	wait := true
	for wait {
		_, err = kudoClientset.KudoV1beta1().Instances(namespace).Get(name, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			wait = false
		}
	}

	cleanupPVCs := d.Get("cleanup_pvcs").(bool)
	if !cleanupPVCs {
		return nil
	}

	// get the PVCs too please
	pvcs, ok := d.GetOk("pvcs")
	if ok {
		pvcList := pvcs.([]interface{})
		kubeClient := config.GetKubernetesClient()

		propagationPolicy := metav1.DeletePropagationForeground
		options := &metav1.DeleteOptions{
			PropagationPolicy: &propagationPolicy,
		}

		for _, pvc := range pvcList {

			err = kubeClient.CoreV1().PersistentVolumeClaims(namespace).Delete(pvc.(string), options)
			wait := true
			for wait {
				_, err = kubeClient.CoreV1().PersistentVolumeClaims(namespace).Get(pvc.(string), metav1.GetOptions{})
				if errors.IsNotFound(err) {
					wait = false
				}
			}
		}
	}

	return nil
}
