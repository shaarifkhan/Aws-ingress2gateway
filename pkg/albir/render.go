package albir

import (
	"fmt"
	"strings"
)

// RenderSummary returns a small human-readable view of the current model.
func RenderSummary(model Model) string {
	var builder strings.Builder

	builder.WriteString("gateways:\n")
	for _, gateway := range model.Gateways {
		builder.WriteString(fmt.Sprintf("- %s/%s listeners=%d\n", gateway.Namespace, gateway.Name, len(gateway.Listeners)))
		for _, listener := range gateway.Listeners {
			builder.WriteString(fmt.Sprintf("  listener=%s protocol=%s port=%d hostname=%s\n", listener.Name, listener.Protocol, listener.Port, listener.Hostname))
		}
	}

	builder.WriteString("httpRoutes:\n")
	for _, route := range model.HTTPRoutes {
		builder.WriteString(fmt.Sprintf("- %s/%s hosts=%s parents=%s rules=%d\n", route.Namespace, route.Name, strings.Join(route.Hostnames, ","), renderParentRefs(route.ParentRefs), len(route.Rules)))
		for _, rule := range route.Rules {
			builder.WriteString(fmt.Sprintf("  path=%s backend=%s\n", rule.Path, renderBackendRefs(rule.BackendRefs)))
		}
	}

	return builder.String()
}

func renderBackendRefs(backendRefs []BackendRef) string {
	parts := make([]string, 0, len(backendRefs))

	for _, backendRef := range backendRefs {
		switch {
		case backendRef.PortName != "":
			parts = append(parts, fmt.Sprintf("%s:%s", backendRef.Name, backendRef.PortName))
		case backendRef.PortNumber != 0:
			parts = append(parts, fmt.Sprintf("%s:%d", backendRef.Name, backendRef.PortNumber))
		default:
			parts = append(parts, backendRef.Name)
		}
	}

	return strings.Join(parts, ",")
}

func renderParentRefs(parentRefs []ParentRef) string {
	parts := make([]string, 0, len(parentRefs))

	for _, parentRef := range parentRefs {
		parts = append(parts, fmt.Sprintf("%s/%s#%s", parentRef.Namespace, parentRef.GatewayName, parentRef.SectionName))
	}

	return strings.Join(parts, ",")
}
