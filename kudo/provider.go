package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	kubernetes "k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	"github.com/kudobuilder/kudo/pkg/client/clientset/versioned"
	"github.com/kudobuilder/kudo/pkg/kudoctl/kube"
	"github.com/kudobuilder/kudo/pkg/kudoctl/kudoinit"
	"github.com/kudobuilder/kudo/pkg/kudoctl/kudoinit/setup"
	"github.com/kudobuilder/kudo/pkg/kudoctl/util/kudo"
	"github.com/kudobuilder/kudo/pkg/version"

	"github.com/mitchellh/go-homedir"
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
	p := &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"kudo_operator": resourceOperator(),
			"kudo_instance": resourceInstance(),
		},
		Schema: map[string]*schema.Schema{
			// Most of these taken to match
			// https://github.com/terraform-providers/terraform-provider-kubernetes/blob/main/kubernetes/provider.go#L25
			"config_path": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc(
					[]string{
						"KUBE_CONFIG",
						"KUBECONFIG",
					},
					"~/.kube/config"),
				Description: "Path to the kube config file, defaults to ~/.kube/config",
			},
			//Other kube config possibilities
			"host": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Explicit host to connect to Kubernetes",
			},
			"insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("KUBE_INSECURE", false),
				Description: "Whether server should be accessed without verifying the TLS certificate.",
			},
			//
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("KUBE_USER", ""),
				Description: "The username to use for HTTP basic authentication when accessing the Kubernetes control plane endpoint.",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("KUBE_PASSWORD", ""),
				Description: "The password to use for HTTP basic authentication when accessing the Kubernetes control plane endpoint.",
			},
			//
			"client_certificate": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("KUBE_CLIENT_CERT_DATA", ""),
				Description: "PEM-encoded client certificate for TLS authentication.",
			},
			"client_key": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("KUBE_CLIENT_KEY_DATA", ""),
				Description: "PEM-encoded client certificate key for TLS authentication.",
			},
			"cluster_ca_certificate": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("KUBE_CLUSTER_CA_CERT_DATA", ""),
				Description: "PEM-encoded root certificates bundle for TLS authentication.",
			},
			//

			"config_context": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("KUBE_CTX", ""),
			},
			"config_context_auth_info": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("KUBE_CTX_AUTH_INFO", ""),
				Description: "",
			},
			"config_context_cluster": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("KUBE_CTX_CLUSTER", ""),
				Description: "",
			},
			"token": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("KUBE_TOKEN", ""),
				Description: "Token to authenticate an service account",
			},
			"load_config_file": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("KUBE_LOAD_CONFIG_FILE", true),
				Description: "Load local kubeconfig.",
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
			"kudo_version": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "KUDO version to install",
				Default:     "0.14.0",
			},
			"namespace": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "kudo-system",
				Description: "Namespace to install KUDO into",
			},
		},
		// ConfigureFunc: kudoConfigureFunc,
	}

	p.ConfigureFunc = func(d *schema.ResourceData) (interface{}, error) {
		terraformVersion := p.TerraformVersion
		if terraformVersion == "" {
			// Terraform 0.12 introduced this field to the protocol
			// We can therefore assume that if it's missing it's 0.10 or 0.11
			terraformVersion = "0.11+compatible"
		}
		return kudoConfigureFunc(d, terraformVersion)
	}

	return p

}

//Config captures the configuration of the KUDO provider
type Config struct {
	KudoImage      string
	Version        string
	Wait           bool
	WaitTimeout    int
	ServiceAccount string
	CRDsOnly       bool
	Namespace      string

	KubernetesClient *kubernetes.Clientset
	KubernetesConfig *restclient.Config

	KudoClient     *kudo.Client
	RawKudoClient  *versioned.Clientset
	KudoKubeClient *kube.Client
}

func (c Config) GetKubernetesClient() *kubernetes.Clientset {
	return c.KubernetesClient
}

//GetKudoClient returns a KUDO client from the configuration object
func (c Config) GetKudoClient() *kudo.Client {
	return c.KudoClient
}

//GetKubernetesClient returns a Kubernetes client from the configuration object
func (c Config) GetKudoKubernetesClient() *kube.Client {
	return c.KudoKubeClient
}

//ToKUDOOpts returns a KUDO Options object for installing KUDO
func (c Config) ToKUDOOpts() kudoinit.Options {
	opts := kudoinit.Options{
		Version:                       c.Version,
		Namespace:                     c.Namespace,
		TerminationGracePeriodSeconds: 300,
		Image:                         fmt.Sprintf("%v:v%v", c.KudoImage, c.Version),
		ServiceAccount:                c.ServiceAccount,
		ImagePullPolicy:               "Always",
		//TODO(@runyontr) use cert manager at some point
		SelfSignedWebhookCA: true,
	}
	return opts
}

func kudoConfigureFunc(data *schema.ResourceData, terraformVersion string) (interface{}, error) {
	// Mostly copied from kubernetes provider:
	var cfg *restclient.Config
	var err error
	if data.Get("load_config_file").(bool) {
		// Config file loading
		cfg, err = tryLoadingConfigFile(data)
	}

	if err != nil {
		return nil, err
	}
	if cfg == nil {
		// Attempt to load in-cluster config
		cfg, err = restclient.InClusterConfig()
		if err != nil {
			// Fallback to standard config if we are not running inside a cluster
			if err == restclient.ErrNotInCluster {
				cfg = &restclient.Config{}
			} else {
				return nil, fmt.Errorf("Failed to configure: %s", err)
			}
		}
	}

	// Overriding with static configuration
	cfg.UserAgent = fmt.Sprintf("HashiCorp/1.0 Terraform/%s", terraformVersion)

	if v, ok := data.GetOk("host"); ok {
		cfg.Host = v.(string)
	}
	if v, ok := data.GetOk("username"); ok {
		cfg.Username = v.(string)
	}
	if v, ok := data.GetOk("password"); ok {
		cfg.Password = v.(string)
	}
	if v, ok := data.GetOk("insecure"); ok {
		cfg.Insecure = v.(bool)
	}
	if v, ok := data.GetOk("cluster_ca_certificate"); ok {
		cfg.CAData = bytes.NewBufferString(v.(string)).Bytes()
	}
	if v, ok := data.GetOk("client_certificate"); ok {
		cfg.CertData = bytes.NewBufferString(v.(string)).Bytes()
	}
	if v, ok := data.GetOk("client_key"); ok {
		cfg.KeyData = bytes.NewBufferString(v.(string)).Bytes()
	}
	if v, ok := data.GetOk("token"); ok {
		cfg.BearerToken = v.(string)
	}

	if logging.IsDebugOrHigher() {
		log.Printf("[DEBUG] Enabling HTTP requests/responses tracing")
		cfg.WrapTransport = func(rt http.RoundTripper) http.RoundTripper {
			return logging.NewTransport("Kubernetes", rt)
		}
	}

	k, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("Failed to configure: %s", err)
	}

	log.Println("[DEBUG] kudo provider configure:")
	c := Config{}
	c.KubernetesConfig = cfg
	c.KubernetesClient = k
	c.RawKudoClient, err = versioned.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("Failed to obtain client to KUDO CRDs: %v", err)
	}
	c.KudoClient = kudo.NewClientFromK8s(c.RawKudoClient, c.KubernetesClient)

	extClient, err := apiextensionsclient.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("could not get Kubernetes client: %s", err)
	}
	dynamicClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("could not create Kubernetes dynamic client: %s", err)
	}

	c.KudoKubeClient = &kube.Client{
		KubeClient:    c.KubernetesClient,
		DynamicClient: dynamicClient,
		ExtClient:     extClient,
	}

	//KUDO installation configurations

	if v, ok := data.GetOk("image"); ok {
		c.KudoImage = v.(string)
	}
	if v, ok := data.GetOk("kudo_version"); ok {
		c.Version = v.(string)
	} else {
		c.Version = version.Get().String()
	}
	if v, ok := data.GetOk("wait"); ok {
		c.Wait = v.(bool)
	}
	if v, ok := data.GetOk("wait_timeout"); ok {
		c.WaitTimeout = v.(int)
	}
	if v, ok := data.GetOk("service_account"); ok {
		c.ServiceAccount = v.(string)
	}
	if v, ok := data.GetOk("namespace"); ok {
		c.Namespace = v.(string)
	}
	c.CRDsOnly = false

	log.Printf("[DEBUG] Config %+v", c)

	//create KubeClient from kubeconfig
	client := c.GetKudoKubernetesClient()
	log.Printf("[DEBUG] Created the kube client from the config")
	//install KUDO
	log.Printf("[DEBUG] Running install")
	opts := c.ToKUDOOpts()
	log.Printf("[DEBUG] KUDO Opts: %+v", opts)

	//try an install, but don't care if it succeeds or not so plans
	// can be run without an active cluster
	installer := setup.NewInstaller(opts, false)
	if installer == nil {
		log.Println("[ERROR] [KUDO] Error creating installer.")
		return c, err
	}
	err = installer.Install(c.KudoKubeClient)
	if err != nil {
		log.Printf("[ERROR] [KUDO] Error installing KUDO: %+v", err)
		return c, err
	}

	err = setup.WatchKUDOUntilReady(c.KubernetesClient, opts, 600)

	if err != nil {
		log.Printf("[ERROR] [KUDO] Error installing KUDO: %+v", err)
		return c, err
	}
	//Wait for health of controller
	return c, waitForControllerHealth(client, opts, time.Minute)
}

func waitForControllerHealth(client *kube.Client, opts kudoinit.Options, timeout time.Duration) error {

	start := time.Now()
	for {
		time.Sleep(50 * time.Millisecond)
		ss, err := client.KubeClient.AppsV1().StatefulSets(opts.Namespace).Get("kudo-controller-manager", metav1.GetOptions{})
		if err != nil {
			log.Printf("[DEBUG] Error getting kudo controller: %v\n", err)
			continue
		}
		if ss.Status.ReadyReplicas == ss.Status.CurrentReplicas {
			//Healthy!
			return nil
		}
		if time.Since(start) > timeout {
			return fmt.Errorf("timed out waiting for healthy KUDO controller")
		}
	}
}

func tryLoadingConfigFile(d *schema.ResourceData) (*restclient.Config, error) {
	path, err := homedir.Expand(d.Get("config_path").(string))
	if err != nil {
		return nil, err
	}

	loader := &clientcmd.ClientConfigLoadingRules{
		ExplicitPath: path,
	}

	overrides := &clientcmd.ConfigOverrides{}
	ctxSuffix := "; default context"

	ctx, ctxOk := d.GetOk("config_context")
	authInfo, authInfoOk := d.GetOk("config_context_auth_info")
	cluster, clusterOk := d.GetOk("config_context_cluster")
	if ctxOk || authInfoOk || clusterOk {
		ctxSuffix = "; overriden context"
		if ctxOk {
			overrides.CurrentContext = ctx.(string)
			ctxSuffix += fmt.Sprintf("; config ctx: %s", overrides.CurrentContext)
			log.Printf("[DEBUG] Using custom current context: %q", overrides.CurrentContext)
		}

		overrides.Context = clientcmdapi.Context{}
		if authInfoOk {
			overrides.Context.AuthInfo = authInfo.(string)
			ctxSuffix += fmt.Sprintf("; auth_info: %s", overrides.Context.AuthInfo)
		}
		if clusterOk {
			overrides.Context.Cluster = cluster.(string)
			ctxSuffix += fmt.Sprintf("; cluster: %s", overrides.Context.Cluster)
		}
		log.Printf("[DEBUG] Using overidden context: %#v", overrides.Context)
	}

	cc := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loader, overrides)
	cfg, err := cc.ClientConfig()
	if err != nil {
		if pathErr, ok := err.(*os.PathError); ok && os.IsNotExist(pathErr.Err) {
			log.Printf("[INFO] Unable to load config file as it doesn't exist at %q", path)
			return nil, nil
		}
		return nil, fmt.Errorf("Failed to load config (%s%s): %s", path, ctxSuffix, err)
	}

	log.Printf("[INFO] Successfully loaded config file (%s%s)", path, ctxSuffix)
	return cfg, nil
}
