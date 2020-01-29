package main

import (
	"fmt"
	"log"
	// "os"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/kudobuilder/kudo/pkg/kudoctl/kube"
	"github.com/kudobuilder/kudo/pkg/kudoctl/kudoinit"
	"github.com/kudobuilder/kudo/pkg/kudoctl/kudoinit/setup"
	"github.com/kudobuilder/kudo/pkg/kudoctl/util/kudo"
)

// Need to add other options for connecting to kubernetes:
/*
provider "kubernetes" {
  host                   = element(concat(data.aws_eks_cluster.cluster[*].endpoint, list("")), 0)
  cluster_ca_certificate = base64decode(element(concat(data.aws_eks_cluster.cluster[*].certificate_authority.0.data, list("")), 0))
  token                  = element(concat(data.aws_eks_cluster_auth.cluster[*].token, list("")), 0)
  load_config_file       = false
  version                = "~> 1.10"
}

*/

//Provider implements the *schema.Provider interface
func Provider() *schema.Provider {
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"kudo_operator": resourceOperator(),
			"kudo_instance": resourceInstance(),
		},
		Schema: map[string]*schema.Schema{
			"kubeconfig": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Kubeconfig to use to talk to Kubernetes Cluster",
			},
			//Other kube config possibilities
			"host": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Explicit host to connect to Kubernetes",
			},
			"cluster_ca_certificate": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Certificate Authority for Kubernetes",
			},
			"token": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Token to connect to Kubernetes",
			},
			//KUDO specific configs
			"image": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "kudobuilder/controller",
				Description: "Override KUDO controller base image",
			},
			"service_account": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "kudo-manager",
				Description: "Override the default serviceAccount kudo-manager",
			},
			"wait": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Block until KUDO manager is running and ready to recieve requests",
			},
			"wait_timeout": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Wait timeout to be used.",
				Default:     300,
			},
			"webhooks": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Comma separated list of webhooks to install",
			},
			"kudo_version": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "KUDO version to install",
			},
			"namespace": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "kudo-system",
				Description: "Namespace to install KUDO into",
			},
		},
		ConfigureFunc: kudoConfigureFunc,
	}
}

//Config captures the configuration of the KUDO provider
type Config struct {
	Kubeconfig string

	//other connection options
	CertificateAuthority string
	Token                string
	Host                 string

	KudoImage      *string
	Webhooks       *string
	Version        *string
	Wait           bool
	WaitTimeout    int
	ServiceAccount *string
	CRDsOnly       bool
	Namespace      string
}

//GetKudoClient returns a KUDO client from the configuration object
func (c Config) GetKudoClient() (*kudo.Client, error) {
	//TODO add options for using Host/CA/token
	return kudo.NewClient(c.Kubeconfig, 0, true)
}

//GetKubernetesClient returns a Kubernetes client from the configuration object
func (c Config) GetKubernetesClient() (*kube.Client, error) {
	//TODO add options for using Host/CA/token
	return kube.GetKubeClient(c.Kubeconfig)
}

//ToKUDOOpts returns a KUDO Options object for installing KUDO
func (c Config) ToKUDOOpts() kudoinit.Options {
	opts := kudoinit.Options{
		Version:                       *c.Version,
		Namespace:                     c.Namespace,
		TerminationGracePeriodSeconds: 300,
		Image:                         fmt.Sprintf("%v:v%v", *c.KudoImage, *c.Version),
		ServiceAccount:                *c.ServiceAccount,
	}
	if c.Webhooks != nil && *c.Webhooks != "" {
		opts.Webhooks = strings.Split(*c.Webhooks, ",")
	}
	return opts
}

func kudoConfigureFunc(data *schema.ResourceData) (interface{}, error) {
	log.Println("[DEBUG] kudo provider configure:")
	c := Config{}
	if kubeconfig, ok := data.GetOk("kubeconfig"); ok {
		//get config based on this config file
		c.Kubeconfig = kubeconfig.(string)
	} else {
		//build a config from the following
		c.CertificateAuthority = data.Get("cluster_ca_certificate").(string)
		c.Host = data.Get("host").(string)
		c.Token = data.Get("token").(string)
	}

	kubeconfig := data.Get("kubeconfig").(string)

	image := data.Get("image").(string)
	//fix this for later.
	// When empty, we get this error
	/*
		2020-01-26T22:14:52.864-0500 [DEBUG] plugin.terraform-provider-kudo: 2020/01/26 22:14:52 [ERROR] There was an error running the install of KUDO: prerequisites: failed to install: Error when creating resource selfsigned-issuer/kudo-system. the server could not find the requested resource
	*/
	webhooks := data.Get("webhooks").(string)
	version := data.Get("kudo_version").(string)
	wait := data.Get("wait").(bool)
	waitTimeout := data.Get("wait_timeout").(int)
	serviceAccount := data.Get("service_account").(string)
	namespace := data.Get("namespace").(string)
	log.Printf("[DEBUG] KUDO provider kubeconfig: %v", kubeconfig)

	c.KudoImage = &image
	c.Webhooks = &webhooks
	c.Version = &version
	c.Wait = wait
	c.WaitTimeout = waitTimeout
	c.ServiceAccount = &serviceAccount
	c.CRDsOnly = false
	c.Namespace = namespace

	log.Printf("[DEBUG] Config %+v", c)

	//create KubeClient from kubeconfig
	client, err := c.GetKubernetesClient()
	if err != nil {
		return c, err
	}
	log.Printf("[DEBUG] Created the kube client from the config")
	//install KUDO
	log.Printf("[DEBUG] Running install")
	opts := c.ToKUDOOpts()
	log.Printf("[DEBUG] KUDO Opts: %+v", opts)
	err = setup.Install(client, opts, false)
	log.Printf("[DEBUG] Install was run")
	if err != nil {
		log.Printf("[ERROR] There was an error running the install of KUDO: %v", err)
	}
	return c, err
}
