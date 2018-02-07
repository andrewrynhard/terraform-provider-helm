package kubernetes

import (
	homedir "github.com/mitchellh/go-homedir"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Kubernetes struct{}

func NewKubernetes() *Kubernetes {
	return &Kubernetes{}
}

// NewKubernetesConfig instantiates and returns a Kubernetes config that defers loading.
func (k *Kubernetes) NewKubernetesConfig(kubeconfig string) (clientcmd.ClientConfig, error) {
	loadingrules := clientcmd.NewDefaultClientConfigLoadingRules()

	explicitPath, err := homedir.Expand(kubeconfig)
	if err != nil {
		return nil, err
	}

	loadingrules.ExplicitPath = explicitPath
	loadingrules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	overrides := &clientcmd.ConfigOverrides{}

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingrules, overrides), nil
}

// NewKubernetesClient instantiates and returns a Kubernetes client.
func (k *Kubernetes) NewKubernetesClient(config clientcmd.ClientConfig) (*kubernetes.Clientset, error) {
	restclientconfig, err := config.ClientConfig()
	if err != nil {
		return nil, err
	}

	client, err := kubernetes.NewForConfig(restclientconfig)
	if err != nil {
		return nil, err
	}

	return client, nil
}
