package gatewayapi

import (
	"strings"
	"testing"

	networkingv1 "k8s.io/api/networking/v1"

	"aws-ingress2gateway/pkg/albir"
)

func TestRenderResourcesYAML(t *testing.T) {
	pathType := networkingv1.PathTypePrefix
	resources := ConvertModel(albir.Model{
		Gateways: []albir.Gateway{
			{
				Name:      "demo-gateway",
				Namespace: "default",
				Listeners: []albir.Listener{
					{
						Name:     "http",
						Port:     80,
						Protocol: "HTTP",
						Hostname: "demo.example.com",
					},
				},
			},
		},
		HTTPRoutes: []albir.HTTPRoute{
			{
				Name:      "demo-route",
				Namespace: "default",
				Hostnames: []string{"demo.example.com"},
				ParentRefs: []albir.ParentRef{
					{
						GatewayName: "demo-gateway",
						SectionName: "http",
						Namespace:   "default",
					},
				},
				Rules: []albir.HTTPRouteRule{
					{
						Path:     "/",
						PathType: &pathType,
						BackendRefs: []albir.BackendRef{
							{
								Name:       "demo-service",
								PortNumber: 80,
							},
						},
					},
				},
			},
		},
	})

	rendered, err := RenderResourcesYAML(resources)
	if err != nil {
		t.Fatalf("render resources yaml returned error: %v", err)
	}

	if !strings.Contains(rendered, "apiVersion: gateway.networking.k8s.io/v1\nkind: Gateway\n") {
		t.Fatal("expected rendered yaml to include a Gateway document")
	}

	if !strings.Contains(rendered, "apiVersion: gateway.networking.k8s.io/v1\nkind: HTTPRoute\n") {
		t.Fatal("expected rendered yaml to include an HTTPRoute document")
	}

	if !strings.Contains(rendered, "name: demo-gateway\n") {
		t.Fatal("expected rendered yaml to include gateway name")
	}

	if !strings.Contains(rendered, "name: demo-route\n") {
		t.Fatal("expected rendered yaml to include route name")
	}

	if !strings.Contains(rendered, "hostname: demo.example.com\n") {
		t.Fatal("expected rendered yaml to include listener hostname")
	}

	if !strings.Contains(rendered, "sectionName: http\n") {
		t.Fatal("expected rendered yaml to include parent ref section name")
	}

	if !strings.Contains(rendered, "---\n") {
		t.Fatal("expected rendered yaml to include document separator")
	}
}

func TestRenderResourcesYAMLEmpty(t *testing.T) {
	rendered, err := RenderResourcesYAML(Resources{})
	if err != nil {
		t.Fatalf("render empty resources returned error: %v", err)
	}

	if rendered != "" {
		t.Fatalf("rendered empty resources = %q, want empty string", rendered)
	}
}

func TestRenderModelYAML(t *testing.T) {
	model := albir.Model{
		Gateways: []albir.Gateway{
			{
				Name:      "demo-gateway",
				Namespace: "default",
			},
		},
		HTTPRoutes: []albir.HTTPRoute{
			{
				Name:      "demo-route",
				Namespace: "default",
			},
		},
	}

	rendered, err := RenderModelYAML(model)
	if err != nil {
		t.Fatalf("render model yaml returned error: %v", err)
	}

	if !strings.Contains(rendered, "kind: Gateway\n") || !strings.Contains(rendered, "kind: HTTPRoute\n") {
		t.Fatal("expected rendered model yaml to include both Gateway and HTTPRoute")
	}
}
