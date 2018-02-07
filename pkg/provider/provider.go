package provider

import (
	"log"

	"github.com/andrewrynhard/terraform-provider-helm/pkg/meta"
	"github.com/andrewrynhard/terraform-provider-helm/pkg/resources"
	"github.com/hashicorp/terraform/helper/schema"
)

// Provider implements the Terraform ResourceProvider API
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"config_path": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"helm_tiller":  resources.ResourceTiller(),
			"helm_release": resources.ResourceRelease(),
		},
		ConfigureFunc: providerConfigureFunc,
	}
}

func providerConfigureFunc(d *schema.ResourceData) (interface{}, error) {
	m := &meta.Meta{
		ExplicitPath: d.Get("config_path").(string),
		Data:         d,
	}

	log.Printf("[DEBUG]: Using kubconfig %s", d.Get("config_path").(string))

	return m, nil
}
