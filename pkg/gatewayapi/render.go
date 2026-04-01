package gatewayapi

import (
	"strings"

	"sigs.k8s.io/yaml"

	"aws-ingress2gateway/pkg/albir"
)

// RenderResourcesYAML renders typed Gateway API resources as multi-document YAML.
func RenderResourcesYAML(resources Resources) (string, error) {
	documents := make([]string, 0, len(resources.Gateways)+len(resources.HTTPRoutes))

	for _, gateway := range resources.Gateways {
		document, err := yaml.Marshal(gateway)
		if err != nil {
			return "", err
		}
		documents = append(documents, string(document))
	}

	for _, route := range resources.HTTPRoutes {
		document, err := yaml.Marshal(route)
		if err != nil {
			return "", err
		}
		documents = append(documents, string(document))
	}

	if len(documents) == 0 {
		return "", nil
	}

	return strings.Join(documents, "---\n"), nil
}

// RenderModelYAML converts the IR model and renders it as multi-document YAML.
func RenderModelYAML(model albir.Model) (string, error) {
	return RenderResourcesYAML(ConvertModel(model))
}
