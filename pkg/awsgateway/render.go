package awsgateway

import (
	"strings"

	"sigs.k8s.io/yaml"

	"aws-ingress2gateway/pkg/albir"
)

// RenderResourcesYAML renders typed AWS Gateway customization CRDs as multi-document YAML.
func RenderResourcesYAML(resources Resources) (string, error) {
	documents := make([]string, 0, len(resources.LoadBalancerConfigurations)+len(resources.TargetGroupConfigurations))

	for _, config := range resources.LoadBalancerConfigurations {
		document, err := yaml.Marshal(config)
		if err != nil {
			return "", err
		}
		documents = append(documents, string(document))
	}

	for _, config := range resources.TargetGroupConfigurations {
		document, err := yaml.Marshal(config)
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

// RenderModelYAML converts the IR model and renders typed AWS Gateway customization CRDs as YAML.
func RenderModelYAML(model albir.Model) (string, error) {
	return RenderResourcesYAML(ConvertModel(model))
}
