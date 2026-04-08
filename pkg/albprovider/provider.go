package albprovider

import (
	"context"
	"fmt"
	"strings"

	"aws-ingress2gateway/pkg/albir"
	"aws-ingress2gateway/pkg/albreader"
	"aws-ingress2gateway/pkg/awsgateway"
	"aws-ingress2gateway/pkg/gatewayapi"
	networkingv1 "k8s.io/api/networking/v1"
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

// FilterStoredIngresses narrows the in-memory ingress set by optional namespace
// and ingress name. Empty values mean "do not filter by this field".
func (p *Provider) FilterStoredIngresses(namespace, ingressName string) {
	if namespace == "" && ingressName == "" {
		return
	}

	filtered := NewStorage()
	for _, ingress := range p.storage.ListIngresses() {
		if !matchesIngressFilter(ingress, namespace, ingressName) {
			continue
		}
		filtered.AddIngress(ingress)
	}

	p.storage = filtered
}

// BuildModel converts the provider's stored ALB ingresses into the tiny IR model.
func (p *Provider) BuildModel() albir.Model {
	return albir.ConvertIngresses(p.storage.ListIngresses())
}

// BuildSummary renders a small human-readable view of the current model.
func (p *Provider) BuildSummary() string {
	return albir.RenderSummary(p.BuildModel())
}

// BuildGatewayAPIResources converts the current model into typed Gateway API objects.
func (p *Provider) BuildGatewayAPIResources() gatewayapi.Resources {
	return gatewayapi.ConvertModel(p.BuildModel())
}

// BuildGatewayAPIYAML renders the provider state as Gateway API YAML.
func (p *Provider) BuildGatewayAPIYAML() (string, error) {
	return gatewayapi.RenderResourcesYAML(p.BuildGatewayAPIResources())
}

// BuildAWSGatewayResources converts the current model into typed AWS Gateway customization CRDs.
func (p *Provider) BuildAWSGatewayResources() awsgateway.Resources {
	return awsgateway.ConvertModel(p.BuildModel())
}

// BuildAWSGatewayYAML renders the provider state as AWS Gateway customization CRD YAML.
func (p *Provider) BuildAWSGatewayYAML() (string, error) {
	return awsgateway.RenderResourcesYAML(p.BuildAWSGatewayResources())
}

// BuildCombinedYAML renders both standard Gateway API resources and AWS
// customization CRDs as one multi-document YAML stream.
func (p *Provider) BuildCombinedYAML() (string, error) {
	gatewayYAML, err := p.BuildGatewayAPIYAML()
	if err != nil {
		return "", err
	}

	awsYAML, err := p.BuildAWSGatewayYAML()
	if err != nil {
		return "", err
	}

	documents := make([]string, 0, 2)
	if gatewayYAML != "" {
		documents = append(documents, gatewayYAML)
	}
	if awsYAML != "" {
		documents = append(documents, awsYAML)
	}

	if len(documents) == 0 {
		return "", nil
	}

	return strings.Join(documents, "---\n"), nil
}

// Storage returns the provider's current in-memory state.
func (p *Provider) Storage() *Storage {
	return p.storage
}

func matchesIngressFilter(ingress networkingv1.Ingress, namespace, ingressName string) bool {
	if namespace != "" && ingress.Namespace != namespace {
		return false
	}

	if ingressName != "" && ingress.Name != ingressName {
		return false
	}

	return true
}
