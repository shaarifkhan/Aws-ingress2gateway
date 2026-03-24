package albprovider

import (
	"context"
	"fmt"

	"aws-ingress2gateway/pkg/albir"
	"aws-ingress2gateway/pkg/albreader"
)

// Provider is a small orchestration layer for our AWS-specific flow.
// At this stage it only knows how to read ALB ingresses from the cluster
// and place them into Storage.
type Provider struct {
	reader  *albreader.ClusterReader
	storage *Storage
}

// NewProvider creates a provider with the default reader and empty storage.
func NewProvider() *Provider {
	return &Provider{
		reader:  albreader.NewClusterReader(),
		storage: NewStorage(),
	}
}

// LoadFromCluster reads all ingresses, filters to ALB ingresses, and stores them.
func (p *Provider) LoadFromCluster(ctx context.Context, namespace string) error {
	ingresses, err := p.reader.ListIngresses(ctx, namespace)
	if err != nil {
		return fmt.Errorf("read ingresses from cluster: %w", err)
	}

	albIngresses := albreader.FilterALBIngresses(ingresses)
	p.storage = NewStorage()
	p.storage.AddIngresses(albIngresses)

	return nil
}

// BuildModel converts the provider's stored ALB ingresses into the tiny IR model.
func (p *Provider) BuildModel() albir.Model {
	return albir.ConvertIngresses(p.storage.ListIngresses())
}

// BuildSummary renders a small human-readable view of the current model.
func (p *Provider) BuildSummary() string {
	return albir.RenderSummary(p.BuildModel())
}

// Storage returns the provider's current in-memory state.
func (p *Provider) Storage() *Storage {
	return p.storage
}
