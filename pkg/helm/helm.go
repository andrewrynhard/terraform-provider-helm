package helm

import (
	"fmt"
	"log"

	"github.com/andrewrynhard/terraform-provider-helm/pkg/kubernetes"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/helm/portforwarder"
)

// Helm represents a helm client.
type Helm struct {
	client *helm.Client
	host   string
}

// NewHelm instantiates and returns a Helm struct.
func (h *Helm) NewHelm(kubeconfig, namespace string) (*Helm, error) {
	t, err := tunnel(kubeconfig, namespace)
	if err != nil {
		return nil, err
	}

	return &Helm{
		host:   t,
		client: helm.NewClient(helm.Host(t)),
	}, nil
}

func (h *Helm) Client() *helm.Client {
	return h.client
}

func (h *Helm) Host() string {
	return h.host
}

func tunnel(kubeconfig, namespace string) (string, error) {
	k := kubernetes.NewKubernetes()
	config, err := k.NewKubernetesConfig(kubeconfig)
	if err != nil {
		return "", err
	}

	restclientconfig, err := config.ClientConfig()
	if err != nil {
		return "", err
	}

	client, err := k.NewKubernetesClient(config)
	if err != nil {
		return "", err
	}

	tunnel, err := portforwarder.New(namespace, client, restclientconfig)
	if err != nil {
		return "", fmt.Errorf("Failed to forward ports for Tiller in namespace %s: %s", namespace, err.Error())
	}

	log.Printf("[DEBUG]: Using tunnel %s", fmt.Sprintf("localhost:%d", tunnel.Local))

	return fmt.Sprintf("localhost:%d", tunnel.Local), nil
}
