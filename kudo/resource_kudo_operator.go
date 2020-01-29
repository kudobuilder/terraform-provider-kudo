package main

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/spf13/afero"

	"github.com/kudobuilder/kudo/pkg/kudoctl/env"
	"github.com/kudobuilder/kudo/pkg/kudoctl/kudohome"
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
			"skip_instance": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
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
			"parameters": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				Default:  make(map[string]string),
			},
			"object_name": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

//TODO - Create and Update functions do the same things

func resourceOperatorCreate(d *schema.ResourceData, m interface{}) error {
	log.Printf("resourceOperatorCreate: %v %v\n", d, m)
	name := d.Get("operator_name").(string)
	namespace := d.Get("operator_namespace").(string)
	repoName := d.Get("repo").(string)
	version := d.Get("operator_version").(string)
	parametersI := d.Get("parameters").(map[string]interface{})
	parameters := make(map[string]string)
	for k, v := range parametersI {
		parameters[k] = v.(string)
	}
	skipInstance := d.Get("skip_instance").(bool)
	log.Printf("[%v] Operator Name: %v", name, name)
	log.Printf("[%v] Operator Namespace: %v", name, namespace)
	log.Printf("[%v] Repo: %v", name, repoName)
	log.Printf("[%v] Operator Version: %v", name, version)
	config := m.(Config)
	log.Printf("[DEBUG] [%v]  Kubeconfig: %v", name, config.Kubeconfig)

	// initialization of filesystem for all commands
	fs := afero.NewOsFs()

	repository, err := repo.ClientFromSettings(fs, kudohome.Home(env.DefaultKudoHome), repoName)
	if err != nil {
		return fmt.Errorf("could not build operator repository: %w", err)
	}
	d.Set("repo", repository.Config.Name)
	kudoClient, err := kudo.NewClient(config.Kubeconfig, 0, true)

	if err != nil {
		return fmt.Errorf("could not create kudo client: %w", err)
	}

	resolver := pkgresolver.New(repository)
	//not sure if the versions are used here or not.
	pkg, err := resolver.Resolve(name, "", version)

	if err != nil {
		return fmt.Errorf("failed to resolve operator package for: %s %w", name, err)
	}
	log.Printf("[KUDO] [%v] Version pulled from repo: %+v", name, pkg.Resources.OperatorVersion.Spec.Version)

	d.Set("operator_version", pkg.Resources.OperatorVersion.Spec.Version)
	d.SetId(fmt.Sprintf("%v-%v-%v", namespace, name, pkg.Resources.OperatorVersion.Spec.Version))
	log.Printf("[KUDO] [%v] id set okay!")
	d.Set("object_name", pkg.Resources.OperatorVersion.ObjectMeta.Name)

	err = kudo.InstallPackage(kudoClient, pkg.Resources, skipInstance, name, namespace, parameters)
	if err != nil {
		log.Printf("[KUDO] [%v] Error installing package: %v", name, err)
		return err
	}
	log.Println("OperatorCreate: [%v] ", name)
	printOperatorConfig(d)
	return resourceOperatorRead(d, m)
}

func printOperatorConfig(d *schema.ResourceData) {
	for k, v := range d.ConnInfo() {
		log.Printf("[%v] %v -> %v", d.Get("operator_name"), k, v)
	}
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

	c := m.(Config)
	log.Printf("[DEBUG] Kubeconfig: %v", c.Kubeconfig)

	kudoClient, err := c.GetKudoClient()

	if err != nil {
		return fmt.Errorf("could not obtain KUDO client: %v", err)
	}

	ov, err := kudoClient.GetOperatorVersion(ovName, namespace)
	if err != nil || ov == nil {
		//error getting
		//not present anymore
		log.Printf("[KUDO] could not an OV with name %v in namespace %v", ovName, namespace)
		d.Partial(true)
		d.SetId("")
		d.Partial(false)
		return nil
	}

	version = ov.Spec.Version
	// if ov.Spec.Version != nil {
	// 	version =
	// }
	opName := ov.Spec.Operator.Name
	// if ov.Spec.Operator != nil {
	// 	opName = ov.Spec.Operator.Name
	// }
	d.Partial(true)
	d.Set("operator_version", version)
	d.Set("operator_name", opName)
	d.Set("object_name", ov.Name)
	d.Partial(false)
	log.Println("OperatorRead: ")
	printOperatorConfig(d)
	return nil
}

func resourceOperatorUpdate(d *schema.ResourceData, m interface{}) error {
	// return resourceOperatorCreate(d, m)
	name := d.Get("operator_name").(string)
	namespace := d.Get("operator_namespace").(string)
	repoName := d.Get("repo").(string)
	version := d.Get("operator_version").(string)
	// ovName := d.Get("operator_version_name").(string)
	parameters := make(map[string]string) //d.Get("parameters").(map[string]string)
	skipInstance := d.Get("skip_instance").(bool)

	config := m.(Config)
	log.Printf("[DEBUG] Kubeconfig: %v", config.Kubeconfig)

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
	kudoClient, err := kudo.NewClient(config.Kubeconfig, 0, true)

	if err != nil {
		return fmt.Errorf("could not create kudo client: %w", err)
	}

	resolver := pkgresolver.New(repository)
	//not sure if the versions are used here or not.
	pkg, err := resolver.Resolve(name, "", version)
	if err != nil {
		return fmt.Errorf("failed to resolve operator package for: %s %w", name, err)
	}

	err = kudo.InstallPackage(kudoClient, pkg.Resources, skipInstance, name, namespace, parameters)
	if err != nil {
		return err
	}
	log.Println("OperatorUpdate: ")
	printOperatorConfig(d)
	return resourceOperatorRead(d, m)
}

//TODO implement unistall here
func resourceOperatorDelete(d *schema.ResourceData, m interface{}) error {
	log.Printf("resourceOperatorCreate: %v %v\n", d, m)
	return nil
}
