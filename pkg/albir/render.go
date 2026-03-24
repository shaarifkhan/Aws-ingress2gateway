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
		builder.WriteString(fmt.Sprintf("- %s/%s\n", gateway.Namespace, gateway.Name))
	}

	builder.WriteString("httpRoutes:\n")
	for _, route := range model.HTTPRoutes {
		builder.WriteString(fmt.Sprintf("- %s/%s\n", route.Namespace, route.Name))
	}

	return builder.String()
}
