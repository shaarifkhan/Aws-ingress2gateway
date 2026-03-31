package gatewayapi

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	networkingv1 "k8s.io/api/networking/v1"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	"aws-ingress2gateway/pkg/albir"
)

const DefaultGatewayClassName gatewayv1.ObjectName = "alb"

// Resources holds the typed Gateway API objects produced from the IR model.
type Resources struct {
	Gateways   []gatewayv1.Gateway
	HTTPRoutes []gatewayv1.HTTPRoute
}

// ConvertModel converts the current IR model into typed Gateway API objects.
func ConvertModel(model albir.Model) Resources {
	resources := Resources{
		Gateways:   make([]gatewayv1.Gateway, 0, len(model.Gateways)),
		HTTPRoutes: make([]gatewayv1.HTTPRoute, 0, len(model.HTTPRoutes)),
	}

	for _, gateway := range model.Gateways {
		resources.Gateways = append(resources.Gateways, ConvertGateway(gateway))
	}

	for _, route := range model.HTTPRoutes {
		resources.HTTPRoutes = append(resources.HTTPRoutes, ConvertHTTPRoute(route))
	}

	return resources
}

// ConvertGateway converts one IR gateway into one typed Gateway API object.
func ConvertGateway(gateway albir.Gateway) gatewayv1.Gateway {
	return gatewayv1.Gateway{
		TypeMeta: metav1.TypeMeta{
			APIVersion: gatewayv1.GroupVersion.String(),
			Kind:       "Gateway",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      gateway.Name,
			Namespace: gateway.Namespace,
		},
		Spec: gatewayv1.GatewaySpec{
			GatewayClassName: DefaultGatewayClassName,
			Listeners:        convertListeners(gateway.Listeners),
		},
	}
}

// ConvertHTTPRoute converts one IR route into one typed Gateway API object.
func ConvertHTTPRoute(route albir.HTTPRoute) gatewayv1.HTTPRoute {
	return gatewayv1.HTTPRoute{
		TypeMeta: metav1.TypeMeta{
			APIVersion: gatewayv1.GroupVersion.String(),
			Kind:       "HTTPRoute",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      route.Name,
			Namespace: route.Namespace,
		},
		Spec: gatewayv1.HTTPRouteSpec{
			CommonRouteSpec: gatewayv1.CommonRouteSpec{
				ParentRefs: convertParentRefs(route.ParentRefs),
			},
			Hostnames: convertHostnames(route.Hostnames),
			Rules:     convertHTTPRouteRules(route.Rules),
		},
	}
}

func convertListeners(listeners []albir.Listener) []gatewayv1.Listener {
	typed := make([]gatewayv1.Listener, 0, len(listeners))

	for _, listener := range listeners {
		typedListener := gatewayv1.Listener{
			Name:     gatewayv1.SectionName(listener.Name),
			Port:     gatewayv1.PortNumber(listener.Port),
			Protocol: gatewayv1.ProtocolType(listener.Protocol),
		}

		if listener.Hostname != "" {
			hostname := gatewayv1.Hostname(listener.Hostname)
			typedListener.Hostname = &hostname
		}

		typed = append(typed, typedListener)
	}

	return typed
}

func convertHostnames(hostnames []string) []gatewayv1.Hostname {
	typed := make([]gatewayv1.Hostname, 0, len(hostnames))

	for _, hostname := range hostnames {
		typed = append(typed, gatewayv1.Hostname(hostname))
	}

	return typed
}

func convertParentRefs(parentRefs []albir.ParentRef) []gatewayv1.ParentReference {
	typed := make([]gatewayv1.ParentReference, 0, len(parentRefs))

	for _, parentRef := range parentRefs {
		name := gatewayv1.ObjectName(parentRef.GatewayName)
		namespace := gatewayv1.Namespace(parentRef.Namespace)
		sectionName := gatewayv1.SectionName(parentRef.SectionName)

		typed = append(typed, gatewayv1.ParentReference{
			Name:        name,
			Namespace:   &namespace,
			SectionName: &sectionName,
		})
	}

	return typed
}

func convertHTTPRouteRules(rules []albir.HTTPRouteRule) []gatewayv1.HTTPRouteRule {
	typed := make([]gatewayv1.HTTPRouteRule, 0, len(rules))

	for _, rule := range rules {
		typedRule := gatewayv1.HTTPRouteRule{
			BackendRefs: convertBackendRefs(rule.BackendRefs),
		}

		if rule.Path != "" || rule.PathType != nil {
			matchType := convertPathMatchType(rule.PathType)
			value := rule.Path
			typedRule.Matches = []gatewayv1.HTTPRouteMatch{
				{
					Path: &gatewayv1.HTTPPathMatch{
						Type:  &matchType,
						Value: &value,
					},
				},
			}
		}

		typed = append(typed, typedRule)
	}

	return typed
}

func convertPathMatchType(pathType *networkingv1.PathType) gatewayv1.PathMatchType {
	if pathType == nil {
		return gatewayv1.PathMatchPathPrefix
	}

	switch *pathType {
	case networkingv1.PathTypeExact:
		return gatewayv1.PathMatchExact
	case networkingv1.PathTypePrefix:
		return gatewayv1.PathMatchPathPrefix
	default:
		return gatewayv1.PathMatchPathPrefix
	}
}

func convertBackendRefs(backendRefs []albir.BackendRef) []gatewayv1.HTTPBackendRef {
	typed := make([]gatewayv1.HTTPBackendRef, 0, len(backendRefs))

	for _, backendRef := range backendRefs {
		ref := gatewayv1.HTTPBackendRef{
			BackendRef: gatewayv1.BackendRef{
				BackendObjectReference: gatewayv1.BackendObjectReference{
					Name: gatewayv1.ObjectName(backendRef.Name),
				},
			},
		}

		if backendRef.PortNumber != 0 {
			port := gatewayv1.PortNumber(backendRef.PortNumber)
			ref.Port = &port
		}

		typed = append(typed, ref)
	}

	return typed
}
