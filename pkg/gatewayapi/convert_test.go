package gatewayapi

import (
	"testing"

	networkingv1 "k8s.io/api/networking/v1"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	"aws-ingress2gateway/pkg/albir"
)

func TestConvertGateway(t *testing.T) {
	gateway := albir.Gateway{
		Name:      "demo-gateway",
		Namespace: "default",
		Listeners: []albir.Listener{
			{
				Name:     "http",
				Port:     80,
				Protocol: "HTTP",
				Hostname: "demo.example.com",
			},
			{
				Name:     "https",
				Port:     443,
				Protocol: "HTTPS",
				Hostname: "demo.example.com",
			},
		},
	}

	typed := ConvertGateway(gateway)

	if typed.Name != "demo-gateway" || typed.Namespace != "default" {
		t.Fatalf("typed gateway identity = %s/%s, want default/demo-gateway", typed.Namespace, typed.Name)
	}

	if typed.Spec.GatewayClassName != DefaultGatewayClassName {
		t.Fatalf("gateway class name = %q, want %q", typed.Spec.GatewayClassName, DefaultGatewayClassName)
	}

	if len(typed.Spec.Listeners) != 2 {
		t.Fatalf("got %d listeners, want 2", len(typed.Spec.Listeners))
	}

	if typed.Spec.Listeners[0].Protocol != gatewayv1.ProtocolType("HTTP") || typed.Spec.Listeners[0].Port != 80 {
		t.Fatalf("first listener = %#v, want HTTP/80", typed.Spec.Listeners[0])
	}

	if typed.Spec.Listeners[0].Hostname == nil || string(*typed.Spec.Listeners[0].Hostname) != "demo.example.com" {
		t.Fatal("expected first listener hostname to be preserved")
	}
}

func TestConvertHTTPRoute(t *testing.T) {
	pathType := networkingv1.PathTypePrefix
	route := albir.HTTPRoute{
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
						TargetType: "ip",
					},
				},
			},
		},
	}

	typed := ConvertHTTPRoute(route)

	if typed.Name != "demo-route" || typed.Namespace != "default" {
		t.Fatalf("typed route identity = %s/%s, want default/demo-route", typed.Namespace, typed.Name)
	}

	if len(typed.Spec.Hostnames) != 1 || string(typed.Spec.Hostnames[0]) != "demo.example.com" {
		t.Fatalf("typed hostnames = %#v, want [demo.example.com]", typed.Spec.Hostnames)
	}

	if len(typed.Spec.ParentRefs) != 1 {
		t.Fatalf("got %d parent refs, want 1", len(typed.Spec.ParentRefs))
	}

	if typed.Spec.ParentRefs[0].Name != "demo-gateway" {
		t.Fatalf("parent ref name = %q, want demo-gateway", typed.Spec.ParentRefs[0].Name)
	}

	if typed.Spec.ParentRefs[0].SectionName == nil || *typed.Spec.ParentRefs[0].SectionName != "http" {
		t.Fatal("expected section name http")
	}

	if len(typed.Spec.Rules) != 1 {
		t.Fatalf("got %d rules, want 1", len(typed.Spec.Rules))
	}

	if len(typed.Spec.Rules[0].Matches) != 1 || typed.Spec.Rules[0].Matches[0].Path == nil {
		t.Fatal("expected one path match")
	}

	if typed.Spec.Rules[0].Matches[0].Path.Value == nil || *typed.Spec.Rules[0].Matches[0].Path.Value != "/" {
		t.Fatal("expected route path /")
	}

	if len(typed.Spec.Rules[0].BackendRefs) != 1 {
		t.Fatalf("got %d backend refs, want 1", len(typed.Spec.Rules[0].BackendRefs))
	}

	if typed.Spec.Rules[0].BackendRefs[0].Name != "demo-service" {
		t.Fatalf("backend ref name = %q, want demo-service", typed.Spec.Rules[0].BackendRefs[0].Name)
	}

	if typed.Spec.Rules[0].BackendRefs[0].Port == nil || *typed.Spec.Rules[0].BackendRefs[0].Port != 80 {
		t.Fatal("expected backend ref port 80")
	}
}

func TestConvertModel(t *testing.T) {
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

	resources := ConvertModel(model)

	if len(resources.Gateways) != 1 {
		t.Fatalf("got %d gateways, want 1", len(resources.Gateways))
	}

	if len(resources.HTTPRoutes) != 1 {
		t.Fatalf("got %d http routes, want 1", len(resources.HTTPRoutes))
	}
}
