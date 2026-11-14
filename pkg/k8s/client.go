package k8s

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// LoadKubeConfig resolves the cluster configuration the same way kubectl
// does: the files listed in the KUBECONFIG environment variable, or
// ~/.kube/config (%USERPROFILE%\.kube\config on Windows), using the current
// context. Inside a pod it falls back to the in-cluster configuration.
func LoadKubeConfig() (*rest.Config, error) {
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	).ClientConfig()
	if err != nil {
		return nil, err
	}
	// One evaluation reads about 60 objects in a row. The client's default
	// limit (5 requests per second after a burst of 10) would make it take
	// several seconds, so allow a whole evaluation through at once.
	config.QPS = 50
	config.Burst = 100
	return config, nil
}

func NewClientSet(config *rest.Config) (*kubernetes.Clientset, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}
