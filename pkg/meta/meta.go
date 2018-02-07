package meta

import (
	"github.com/andrewrynhard/terraform-provider-helm/pkg/helm"
	"github.com/andrewrynhard/terraform-provider-helm/pkg/helm/repo"
	"github.com/andrewrynhard/terraform-provider-helm/pkg/kubernetes"
	"github.com/hashicorp/terraform/helper/schema"
)

// Meta is the meta struct used by Terraform.
type Meta struct {
	helm.Helm
	kubernetes.Kubernetes
	repo.Repo
	ExplicitPath string
	Data         *schema.ResourceData
}
