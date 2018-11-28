package k8s

import (
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

func NewDynamicClient() (dynamic.Interface, error) {
	// TODO: allow running outside of cluster?
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return dynamic.NewForConfig(cfg)
}
