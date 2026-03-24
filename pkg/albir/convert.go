package albir

import networkingv1 "k8s.io/api/networking/v1"

// ConvertIngress builds the smallest useful IR from one ALB-backed Ingress.
// It does not validate ALB-specific behavior yet; it only preserves identity
// and a link back to the source object for later conversion steps.
func ConvertIngress(ingress networkingv1.Ingress) Model {
	ingressCopy := ingress

	return Model{
		Gateways: []Gateway{
			{
				Name:      ingress.Name,
				Namespace: ingress.Namespace,
				Source:    &ingressCopy,
			},
		},
		HTTPRoutes: []HTTPRoute{
			{
				Name:      ingress.Name,
				Namespace: ingress.Namespace,
				Source:    &ingressCopy,
			},
		},
	}
}

// ConvertIngresses converts a slice of Ingress objects into one combined model.
func ConvertIngresses(ingresses []networkingv1.Ingress) Model {
	model := Model{
		Gateways:   make([]Gateway, 0, len(ingresses)),
		HTTPRoutes: make([]HTTPRoute, 0, len(ingresses)),
	}

	for _, ingress := range ingresses {
		converted := ConvertIngress(ingress)
		model.Gateways = append(model.Gateways, converted.Gateways...)
		model.HTTPRoutes = append(model.HTTPRoutes, converted.HTTPRoutes...)
	}

	return model
}
