package albprovider

import (
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	networkingv1 "k8s.io/api/networking/v1"
)

func TestBuildModel(t *testing.T) {
	provider := NewProvider()
	provider.storage.AddIngresses([]networkingv1.Ingress{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "demo-one",
				Namespace: "default",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "demo-two",
				Namespace: "apps",
			},
		},
	})

	model := provider.BuildModel()

	if len(model.Gateways) != 2 {
		t.Fatalf("got %d gateways, want 2", len(model.Gateways))
	}

	if len(model.HTTPRoutes) != 2 {
		t.Fatalf("got %d http routes, want 2", len(model.HTTPRoutes))
	}

	gatewayNames := map[string]bool{}
	for _, gateway := range model.Gateways {
		gatewayNames[gateway.Namespace+"/"+gateway.Name] = true
		if gateway.Source == nil {
			t.Fatal("expected gateway source ingress pointer to be set")
		}
	}

	if !gatewayNames["default/demo-one"] || !gatewayNames["apps/demo-two"] {
		t.Fatal("expected gateways for both stored ingresses")
	}

	routeNames := map[string]bool{}
	for _, route := range model.HTTPRoutes {
		routeNames[route.Namespace+"/"+route.Name] = true
		if route.Source == nil {
			t.Fatal("expected http route source ingress pointer to be set")
		}
	}

	if !routeNames["default/demo-one"] || !routeNames["apps/demo-two"] {
		t.Fatal("expected http routes for both stored ingresses")
	}
}

func TestBuildSummary(t *testing.T) {
	provider := NewProvider()
	provider.storage.AddIngresses([]networkingv1.Ingress{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "demo-one",
				Namespace: "default",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "demo-two",
				Namespace: "apps",
			},
		},
	})

	summary := provider.BuildSummary()

	if !strings.Contains(summary, "gateways:\n") {
		t.Fatal("expected summary to include gateways header")
	}

	if !strings.Contains(summary, "httpRoutes:\n") {
		t.Fatal("expected summary to include httpRoutes header")
	}

	if !strings.Contains(summary, "- default/demo-one\n") {
		t.Fatal("expected summary to include first ingress identity")
	}

	if !strings.Contains(summary, "- apps/demo-two\n") {
		t.Fatal("expected summary to include second ingress identity")
	}
}
