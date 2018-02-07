package resources

import (
	"log"
	"strings"

	"github.com/andrewrynhard/terraform-provider-helm/pkg/helm/repo"
	"github.com/andrewrynhard/terraform-provider-helm/pkg/meta"
	"github.com/hashicorp/terraform/helper/schema"
	yaml "gopkg.in/yaml.v2"

	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/storage/driver"
)

func ResourceRelease() *schema.Resource {
	return &schema.Resource{
		Create: resourceReleaseCreate,
		Read:   resourceReleaseRead,
		Update: resourceReleaseUpdate,
		Delete: resourceReleaseDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"repo": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"chart": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"namespace": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"tiller_namespace": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"force": &schema.Schema{
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
			"recreate_pods": &schema.Schema{
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
			"values": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"version": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"debug": &schema.Schema{
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
			"metadata": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Status of the deployed release.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name is the name of the release.",
						},
						"revision": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Version is an int32 which represents the version of the release.",
						},
						"namespace": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Namespace is the kubernetes namespace of the release.",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Status of the release.",
						},
						"chart": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the chart.",
						},
						"version": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "A SemVer 2 conformant version string of the chart.",
						},
					},
				},
			},
		},
	}
}

func resourceReleaseCreate(d *schema.ResourceData, m interface{}) error {
	meta := m.(*meta.Meta)
	h, err := meta.NewHelm(meta.ExplicitPath, d.Get("tiller_namespace").(string))
	if err != nil {
		return err
	}
	client := h.Client()

	o := &repo.Options{
		Host:      h.Host(),
		Name:      d.Get("repo").(string),
		Namespace: d.Get("tiller_namespace").(string),
		Chart:     d.Get("chart").(string),
		Version:   d.Get("version").(string),
	}
	chartPath, err := meta.FindChart(o)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Using chart %s", chartPath)

	c, err := chartutil.Load(chartPath)
	if err != nil {
		return err
	}

	rawMap := map[string]interface{}{}
	err = yaml.Unmarshal([]byte(d.Get("values").(string)), &rawMap)
	if err != nil {
		return err
	}

	// TODO: Implement `--set` functionality.
	// err = strvals.ParseInto(d.Get("values").(string), rawMap)
	// if err != nil {
	// 	return err
	// }

	raw, err := yaml.Marshal(rawMap)
	if err != nil {
		return err
	}

	opts := []helm.InstallOption{
		helm.ReleaseName(d.Get("name").(string)),
		helm.ValueOverrides(raw),
		helm.InstallWait(true),
	}

	res, err := client.InstallReleaseFromChart(c, d.Get("namespace").(string), opts...)
	if err != nil {
		return err
	}

	d.Set("metadata", []map[string]interface{}{{
		"name":      res.Release.Name,
		"revision":  res.Release.Version,
		"namespace": res.Release.Namespace,
		"status":    res.Release.Info.Status.Code.String(),
		"chart":     res.Release.Chart.Metadata.Name,
		"version":   res.Release.Chart.Metadata.Version,
	}})

	d.SetId(res.Release.Name)

	return nil
}

func resourceReleaseRead(d *schema.ResourceData, m interface{}) error {
	meta := m.(*meta.Meta)
	h, err := meta.NewHelm(meta.ExplicitPath, d.Get("tiller_namespace").(string))
	if err != nil {
		return err
	}
	client := h.Client()

	releaseHistory, err := client.ReleaseHistory(d.Get("name").(string), helm.WithMaxHistory(1))

	if err != nil && strings.Contains(err.Error(), driver.ErrReleaseNotFound(d.Get("name").(string)).Error()) {
		d.SetId("")
		return nil
	}

	rls := releaseHistory.GetReleases()[0]
	d.Set("metadata", []map[string]interface{}{{
		"name":      rls.Name,
		"revision":  rls.Version,
		"namespace": rls.Namespace,
		"status":    rls.Info.Status.Code.String(),
		"chart":     rls.Chart.Metadata.Name,
		"version":   rls.Chart.Metadata.Version,
	}})

	d.SetId(rls.Name)

	return nil
}

func resourceReleaseUpdate(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceReleaseDelete(d *schema.ResourceData, m interface{}) error {
	meta := m.(*meta.Meta)
	h, err := meta.NewHelm(meta.ExplicitPath, d.Get("tiller_namespace").(string))
	if err != nil {
		return err
	}
	client := h.Client()

	_, err = client.DeleteRelease(d.Get("name").(string), helm.DeletePurge(true))
	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}
