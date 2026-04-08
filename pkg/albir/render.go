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
		builder.WriteString(fmt.Sprintf("- %s/%s scheme=%s listeners=%d\n", gateway.Namespace, gateway.Name, gateway.Scheme, len(gateway.Listeners)))
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

	builder.WriteString("loadBalancerConfigurations:\n")
	for _, config := range model.LoadBalancerConfigurations {
		builder.WriteString(fmt.Sprintf("- %s/%s loadBalancerName=%s scheme=%s wafv2ACLARN=%s attributes=%s listenerConfigurations=%d\n", config.Namespace, config.Name, config.LoadBalancerName, config.Scheme, config.WAFv2ACLARN, renderLoadBalancerAttributes(config.LoadBalancerAttributes), len(config.ListenerConfigurations)))
		for _, listener := range config.ListenerConfigurations {
			builder.WriteString(fmt.Sprintf("  listenerConfiguration protocol=%s port=%d sslPolicy=%s certificates=%s\n", listener.Protocol, listener.Port, listener.SSLPolicy, strings.Join(listener.Certificates, ",")))
		}
	}

	builder.WriteString("targetGroupConfigurations:\n")
	for _, config := range model.TargetGroupConfigurations {
		builder.WriteString(fmt.Sprintf("- %s/%s service=%s targetType=%s healthCheckPath=%s\n", config.Namespace, config.Name, config.ServiceName, config.TargetType, config.HealthCheckPath))
	}

	return builder.String()
}

func renderLoadBalancerAttributes(attributes []LoadBalancerAttribute) string {
	parts := make([]string, 0, len(attributes))

	for _, attribute := range attributes {
		parts = append(parts, fmt.Sprintf("%s=%s", attribute.Key, attribute.Value))
	}

	return strings.Join(parts, ",")
}

func renderBackendRefs(backendRefs []BackendRef) string {
	parts := make([]string, 0, len(backendRefs))

	for _, backendRef := range backendRefs {
		var rendered string
		switch {
		case backendRef.PortName != "":
			rendered = fmt.Sprintf("%s:%s", backendRef.Name, backendRef.PortName)
		case backendRef.PortNumber != 0:
			rendered = fmt.Sprintf("%s:%d", backendRef.Name, backendRef.PortNumber)
		default:
			rendered = backendRef.Name
		}

		if backendRef.TargetType != "" {
			rendered += " targetType=" + backendRef.TargetType
		}

		parts = append(parts, rendered)
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
