package albreader

import (
	"context"
	"fmt"

	networkingv1 "k8s.io/api/networking/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// ClusterReader is the smallest possible reader for step 1.
// It knows how to connect to the cluster and list raw Ingress objects.
type ClusterReader struct{}

// NewClusterReader creates a reader with no extra configuration.
func NewClusterReader() *ClusterReader {
	return &ClusterReader{}
}

// ListIngresses returns all Ingress objects visible in the selected namespace.
// If namespace is empty, the client is not namespace-scoped.
func (r *ClusterReader) ListIngresses(ctx context.Context, namespace string) ([]networkingv1.Ingress, error) {
	clusterClient, err := r.newClient(namespace)
	if err != nil {
		return nil, err
	}

	var ingressList networkingv1.IngressList
	if err := clusterClient.List(ctx, &ingressList); err != nil {
		return nil, fmt.Errorf("list ingresses: %w", err)
	}

	return ingressList.Items, nil
}

func (r *ClusterReader) newClient(namespace string) (client.Client, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("get cluster config: %w", err)
	}

	clusterClient, err := client.New(cfg, client.Options{})
	if err != nil {
		return nil, fmt.Errorf("create cluster client: %w", err)
	}

	if namespace == "" {
		return clusterClient, nil
	}

	return client.NewNamespacedClient(clusterClient, namespace), nil
}
