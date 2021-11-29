package handler

import (
	"flag"
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"os"
	"os/user"
	"path/filepath"
)

type KubernetesClientManager struct {
	clients map[string]*kubernetes.Clientset
}

func (k *KubernetesClientManager) getClient(cluster string) (*kubernetes.Clientset, error) {
	_, exists := k.clients[cluster]

	if exists {
		return k.clients[cluster], nil
	}

	cfg, err := k.getRestClientConfig(cluster)
	if err != nil {
		return nil, err
	}

	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	k.clients[cluster] = client

	return client, nil
}

func (k *KubernetesClientManager) getRestClientConfig(cluster string) (*rest.Config, error) {
	return GetConfigWithContext("")
}

var (
	kubeconfig, apiServerURL string
)

func init() {
	// TODO: Fix this to allow double vendoring this library but still register flags on behalf of users
	flag.StringVar(&kubeconfig, "kubeconfig", "",
		"Paths to a kubeconfig. Only required if out-of-cluster.")

	// This flag is deprecated, it'll be removed in a future iteration, please switch to --kubeconfig.
	flag.StringVar(&apiServerURL, "master", "",
		"(Deprecated: switch to `--kubeconfig`) The address of the Kubernetes API server. Overrides any value in kubeconfig. "+
			"Only required if out-of-cluster.")
}

// * $HOME/.kube/config if exists
func GetConfigWithContext(context string) (*rest.Config, error) {
	cfg, err := loadConfig(context)
	if err != nil {
		return nil, err
	}

	if cfg.QPS == 0.0 {
		cfg.QPS = 20.0
		cfg.Burst = 30.0
	}

	return cfg, nil
}

// loadConfig loads a REST Config as per the rules specified in GetConfig
func loadConfig(context string) (*rest.Config, error) {

	// If a flag is specified with the config location, use that
	if len(kubeconfig) > 0 {
		return loadConfigWithContext(apiServerURL, kubeconfig, context)
	}
	// If an env variable is specified with the config location, use that
	if len(os.Getenv("KUBECONFIG")) > 0 {
		return loadConfigWithContext(apiServerURL, os.Getenv("KUBECONFIG"), context)
	}
	// If no explicit location, try the in-cluster config
	if c, err := rest.InClusterConfig(); err == nil {
		return c, nil
	}
	// If no in-cluster config, try the default location in the user's home directory
	if usr, err := user.Current(); err == nil {
		if c, err := loadConfigWithContext(apiServerURL, filepath.Join(usr.HomeDir, ".kube", "config"),
			context); err == nil {
			return c, nil
		}
	}

	return nil, fmt.Errorf("could not locate a kubeconfig")
}

func loadConfigWithContext(apiServerURL, kubeconfig, context string) (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig},
		&clientcmd.ConfigOverrides{
			ClusterInfo: clientcmdapi.Cluster{
				Server: apiServerURL,
			},
			CurrentContext: context,
		}).ClientConfig()
}
