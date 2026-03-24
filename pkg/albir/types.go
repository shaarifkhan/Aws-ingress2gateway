package albir

import networkingv1 "k8s.io/api/networking/v1"

// Model is the small in-memory result of conversion work.
// It intentionally holds only the core Gateway-shaped objects we need first.
type Model struct {
	Gateways   []Gateway
	HTTPRoutes []HTTPRoute
}

// Gateway is a minimal IR form for the gateway we plan to generate.
// Source points back to the Ingress that produced it.
type Gateway struct {
	Name      string
	Namespace string
	Source    *networkingv1.Ingress
}

// HTTPRoute is a minimal IR form for the route we plan to generate.
// Source points back to the Ingress that produced it.
type HTTPRoute struct {
	Name      string
	Namespace string
	Source    *networkingv1.Ingress
}
