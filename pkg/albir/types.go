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
	Listeners []Listener
	Source    *networkingv1.Ingress
}

// Listener is a small gateway-side view of where traffic can arrive.
type Listener struct {
	Name     string
	Port     int32
	Protocol string
	Hostname string
}

// HTTPRoute is a minimal IR form for the route we plan to generate.
// Source points back to the Ingress that produced it.
type HTTPRoute struct {
	Name       string
	Namespace  string
	Hostnames  []string
	ParentRefs []ParentRef
	Rules      []HTTPRouteRule
	Source     *networkingv1.Ingress
}

// ParentRef is a small route-side link back to the generated gateway listener.
type ParentRef struct {
	GatewayName  string
	SectionName  string
	Namespace    string
}

// HTTPRouteRule is a small flattened view of one Ingress HTTP path rule.
type HTTPRouteRule struct {
	Path        string
	PathType    *networkingv1.PathType
	BackendRefs []BackendRef
}

// BackendRef is the smallest backend shape we need from an Ingress path.
type BackendRef struct {
	Name       string
	PortNumber int32
	PortName   string
}
