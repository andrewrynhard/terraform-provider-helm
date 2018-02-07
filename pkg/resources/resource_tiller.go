package resources

import (
	"fmt"
	"log"
	"time"

	"github.com/andrewrynhard/terraform-provider-helm/pkg/meta"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/helm/cmd/helm/installer"
)

const (
	// This must be kept up to date with upstream since it is not exposed.
	deploymentName = "tiller-deploy"
	// This must be kept up to date with upstream since it is not exposed.
	version = "v2.8.0"
)

func ResourceTiller() *schema.Resource {
	return &schema.Resource{
		Create: resourceTillerCreate,
		Read:   resourceTillerRead,
		Update: resourceTillerUpdate,
		Delete: resourceTillerDelete,

		Schema: map[string]*schema.Schema{
			"roleref": {
				Type:     schema.TypeString,
				Required: true,
			},
			"clusterrolebinding": {
				Type:     schema.TypeString,
				Required: true,
			},
			"image": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "gcr.io/kubernetes-helm/tiller:" + version,
			},
			"namespace": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "kube-system",
				Description: "",
			},
			"service_account": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceTillerCreate(d *schema.ResourceData, m interface{}) error {
	meta := m.(*meta.Meta)

	config, err := meta.NewKubernetesConfig(meta.ExplicitPath)
	if err != nil {
		return err
	}

	client, err := meta.NewKubernetesClient(config)
	if err != nil {
		return err
	}

	serviceaccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name: d.Get("service_account").(string),
		},
	}
	_, err = client.Core().ServiceAccounts(d.Get("namespace").(string)).Create(serviceaccount)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return err
		}
	}

	clusterrolebinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: d.Get("clusterrolebinding").(string),
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     d.Get("roleref").(string),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      d.Get("service_account").(string),
				Namespace: d.Get("namespace").(string),
			},
		},
	}

	_, err = client.Rbac().ClusterRoleBindings().Create(clusterrolebinding)
	if err != nil {
		if !errors.IsAlreadyExists(err) {
			return err
		}
	}

	o := &installer.Options{
		Namespace:      d.Get("namespace").(string),
		ImageSpec:      d.Get("image").(string),
		ServiceAccount: d.Get("service_account").(string),
	}
	if err := installer.Install(client, o); err != nil {
		// TODO: Should we error here and require an import instead?
		if errors.IsAlreadyExists(err) {
			d.SetId(deploymentName + "_" + d.Get("namespace").(string))
			return nil
		}

		return fmt.Errorf("Error installing Tiller: %s", err)
	}

	if err := waitForTiller(client, o); err != nil {
		return err
	}

	log.Printf("[DEBUG] Installed Tiller to namespace %s.", d.Get("namespace").(string))

	d.SetId(deploymentName + "_" + d.Get("namespace").(string))

	return nil
}

func resourceTillerRead(d *schema.ResourceData, m interface{}) error {
	meta := m.(*meta.Meta)

	config, err := meta.NewKubernetesConfig(meta.ExplicitPath)
	if err != nil {
		return err
	}

	client, err := meta.NewKubernetesClient(config)
	if err != nil {
		return err
	}

	obj, err := client.Extensions().Deployments(d.Get("namespace").(string)).Get(deploymentName, metav1.GetOptions{})
	if err != nil {
		// TODO: errors.IsNotFound(err error)
		return err
	}

	d.Set("namespace", obj.Namespace)

	return nil
}

func resourceTillerUpdate(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceTillerDelete(d *schema.ResourceData, m interface{}) error {
	return nil
}

func waitForTiller(client *kubernetes.Clientset, o *installer.Options) error {
	stateConf := &resource.StateChangeConf{
		Target:  []string{"Running"},
		Pending: []string{"Pending"},
		Timeout: 5 * time.Minute,
		Refresh: func() (interface{}, string, error) {
			obj, err := client.Extensions().Deployments(o.Namespace).Get(deploymentName, metav1.GetOptions{})
			if err != nil {
				return obj, "Error", err
			}

			if obj.Status.ReadyReplicas > 0 {
				return obj, "Running", nil
			}

			return obj, "Pending", nil
		},
	}

	_, err := stateConf.WaitForState()

	return err
}
