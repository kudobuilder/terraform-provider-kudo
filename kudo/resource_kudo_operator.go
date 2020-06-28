package main

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/spf13/afero"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kudobuilder/kudo/pkg/kudoctl/env"
	"github.com/kudobuilder/kudo/pkg/kudoctl/kudohome"
	"github.com/kudobuilder/kudo/pkg/kudoctl/packages"
	pkgresolver "github.com/kudobuilder/kudo/pkg/kudoctl/packages/resolver"
	"github.com/kudobuilder/kudo/pkg/kudoctl/util/kudo"
	"github.com/kudobuilder/kudo/pkg/kudoctl/util/repo"
)

func resourceOperator() *schema.Resource {
	return &schema.Resource{
		Create: resourceOperatorCreate,
		Read:   resourceOperatorRead,
		Update: resourceOperatorUpdate,
		Delete: resourceOperatorDelete,
		Exists: resourceOperatorExists,
		Schema: map[string]*schema.Schema{
			"operator_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"operator_version": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"operator_namespace": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "default",
				Description: "Namespace to install the Operator Version",
			},
			"repo": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of Repository in KUDO repo config file",
				Computed:    true,
			},
			"object_name": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func getOperatorVersionFromRepo(d *schema.ResourceData, m interface{}) (*packages.Package, error) {

	repoName := d.Get("repo").(string)
	opearatorVersion := d.Get("operator_version").(string)
	name := d.Get("operator_name").(string)
	// initialization of filesystem for all commands
	fs := afero.NewOsFs()

	repository, err := repo.ClientFromSettings(fs, kudohome.Home(env.DefaultKudoHome), repoName)
	if err != nil {
		return nil, fmt.Errorf("could not build operator repository: %w", err)
	}
	d.Set("repo", repository.Config.Name)

	resolver := pkgresolver.New(repository)
	//not sure if the versions are used here or not.
	return resolver.Resolve(name, "", opearatorVersion)
}

func resourceOperatorCreate(d *schema.ResourceData, m interface{}) error {
	log.Printf("resourceOperatorCreate: %v %v\n", d, m)
	name := d.Get("operator_name").(string)
	namespace := d.Get("operator_namespace").(string)
	repoName := d.Get("repo").(string)
	version := d.Get("operator_version").(string)
	log.Printf("[%v] Operator Name: %v", name, name)
	log.Printf("[%v] Operator Namespace: %v", name, namespace)
	log.Printf("[%v] Repo: %v", name, repoName)
	log.Printf("[%v] Operator Version: %v", name, version)
	config := m.(Config)
	kudoClient := config.GetKudoClient()

	pkg, err := getOperatorVersionFromRepo(d, m)

	if err != nil {
		return fmt.Errorf("failed to resolve operator package for: %s %w", name, err)
	}
	log.Printf("[KUDO] [%v] Version pulled from repo: %+v", name, pkg.Resources.OperatorVersion.Spec.Version)

	d.Set("operator_version", pkg.Resources.OperatorVersion.Spec.Version)
	d.SetId(id(pkg.Resources.OperatorVersion.ObjectMeta.Name, namespace))
	log.Printf("[KUDO] [%v] id set okay!", d.Id())
	d.Set("object_name", pkg.Resources.OperatorVersion.ObjectMeta.Name)

	err = applyPackage(kudoClient, pkg, namespace)

	if err != nil {
		return err
	}
	return resourceOperatorRead(d, m)
}

func printOperatorConfig(d *schema.ResourceData) {
	for k, v := range d.ConnInfo() {
		log.Printf("[%v] %v -> %v", d.Get("operator_name"), k, v)
	}
}

func resourceOperatorExists(d *schema.ResourceData, m interface{}) (bool, error) {
	config := m.(Config)

	obj, def := d.GetOk("object_name")
	if !def {
		return false, nil
	}

	client := config.GetKudoClient()

	_, err := client.GetOperatorVersion(obj.(string), d.Get("operator_namespace").(string))

	return err == nil, err
}

func resourceOperatorRead(d *schema.ResourceData, m interface{}) error {
	log.Printf("resourceOperatorCreate: %v %v\n", d, m)
	// return nil
	namespace := d.Get("operator_namespace").(string)
	version := d.Get("operator_version").(string)
	ovName := d.Get("object_name").(string)
	if version == "" || ovName == "" {
		log.Printf("[KUDO] could not find Version (%v) or Ovname(%v) value", version, ovName)
		d.Partial(true)
		d.SetId("")
		d.Partial(false) // Not installed yet
		return nil
	}

	config := m.(Config)
	client := config.GetKudoClient()

	ov, err := client.GetOperatorVersion(ovName, namespace)
	if err != nil {
		return err
	}
	if ov == nil {
		d.SetId("")
		return nil
	}

	d.Set("operator_version", ov.Spec.Version)
	d.Set("operator_name", ov.Spec.Operator.Name)
	d.Set("object_name", ov.Name)
	return nil
}

func resourceOperatorUpdate(d *schema.ResourceData, m interface{}) error {
	// return resourceOperatorCreate(d, m)
	name := d.Get("operator_name").(string)
	namespace := d.Get("operator_namespace").(string)
	repoName := d.Get("repo").(string)
	version := d.Get("operator_version").(string)
	// ovName := d.Get("operator_version_name").(string)

	config := m.(Config)

	// initialization of filesystem for all commands
	fs := afero.NewOsFs()

	repository, err := repo.ClientFromSettings(fs, kudohome.Home(env.DefaultKudoHome), repoName)
	if err != nil {
		return fmt.Errorf("could not build operator repository: %w", err)
	}
	log.Printf("[KUDO] setting repo name to %v", repository.Config.Name)
	d.Partial(true)
	d.Set("repo", repository.Config.Name)
	d.Partial(false)
	kudoClient := config.GetKudoClient()

	resolver := pkgresolver.New(repository)
	//not sure if the versions are used here or not.
	pkg, err := resolver.Resolve(name, "", version)
	if err != nil {
		return fmt.Errorf("failed to resolve operator package for: %s %w", name, err)
	}

	err = applyPackage(kudoClient, pkg, namespace)
	if err != nil {
		return err
	}

	log.Println("OperatorUpdate: ")
	printOperatorConfig(d)
	return resourceOperatorRead(d, m)
}

func applyPackage(kudoClient *kudo.Client, pkg *packages.Package, namespace string) error {
	if kudoClient.OperatorExistsInCluster(pkg.Resources.Operator.Name, namespace) {
		log.Printf("[KUDO] Operator %v already exists in the cluster.  Updates not supported", pkg.Resources.Operator.Name)
	} else {
		operator, err := kudoClient.InstallOperatorObjToCluster(pkg.Resources.Operator, namespace)
		if err != nil {
			log.Printf("[KUDO] [%v] Error installing Operator: %v", operator.Name, err)
			return err
		}
	}
	if ov, err := kudoClient.GetOperatorVersion(pkg.Resources.OperatorVersion.Name, namespace); err == nil && ov != nil {
		log.Printf("[KUDO] OperatorVersion %v already exists in the cluster.  Updates not supported", ov.Name)
	} else {
		ov, err = kudoClient.InstallOperatorVersionObjToCluster(pkg.Resources.OperatorVersion, namespace)
		if err != nil {
			log.Printf("[KUDO] [%v] Error installing OperatorVersion: %v", pkg.Resources.OperatorVersion.Name, err)
			return err
		}
	}
	return nil
}

//TODO implement uninstall here
func resourceOperatorDelete(d *schema.ResourceData, m interface{}) error {
	log.Printf("resourceOperatorCreate: %v %v\n", d, m)
	name := d.Get("object_name").(string)
	namespace := d.Get("operator_namespace").(string)
	config := m.(Config)

	kudoClientset := config.RawKudoClient

	propagationPolicy := metav1.DeletePropagationBackground
	options := &metav1.DeleteOptions{
		PropagationPolicy: &propagationPolicy,
	}

	return kudoClientset.KudoV1beta1().OperatorVersions(namespace).Delete(name, options)

}
