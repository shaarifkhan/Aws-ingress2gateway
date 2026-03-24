package albir

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	networkingv1 "k8s.io/api/networking/v1"
)

func TestConvertIngress(t *testing.T) {
	ingress := networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo",
			Namespace: "default",
		},
	}

	model := ConvertIngress(ingress)

	if len(model.Gateways) != 1 {
		t.Fatalf("got %d gateways, want 1", len(model.Gateways))
	}

	if len(model.HTTPRoutes) != 1 {
		t.Fatalf("got %d http routes, want 1", len(model.HTTPRoutes))
	}

	gateway := model.Gateways[0]
	if gateway.Name != ingress.Name || gateway.Namespace != ingress.Namespace {
		t.Fatalf("gateway identity = %s/%s, want %s/%s", gateway.Namespace, gateway.Name, ingress.Namespace, ingress.Name)
	}

	route := model.HTTPRoutes[0]
	if route.Name != ingress.Name || route.Namespace != ingress.Namespace {
		t.Fatalf("http route identity = %s/%s, want %s/%s", route.Namespace, route.Name, ingress.Namespace, ingress.Name)
	}

	if gateway.Source == nil || route.Source == nil {
		t.Fatal("expected source ingress pointers to be set")
	}

	if gateway.Source.Name != ingress.Name || route.Source.Name != ingress.Name {
		t.Fatal("expected source ingress pointers to preserve ingress data")
	}

	if gateway.Source != route.Source {
		t.Fatal("expected gateway and route to share the same source ingress pointer")
	}
}

func TestConvertIngresses(t *testing.T) {
	ingresses := []networkingv1.Ingress{
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
	}

	model := ConvertIngresses(ingresses)

	if len(model.Gateways) != 2 {
		t.Fatalf("got %d gateways, want 2", len(model.Gateways))
	}

	if len(model.HTTPRoutes) != 2 {
		t.Fatalf("got %d http routes, want 2", len(model.HTTPRoutes))
	}

	if model.Gateways[0].Name != "demo-one" || model.Gateways[1].Name != "demo-two" {
		t.Fatal("expected gateways to preserve ingress ordering")
	}

	if model.HTTPRoutes[0].Namespace != "default" || model.HTTPRoutes[1].Namespace != "apps" {
		t.Fatal("expected http routes to preserve ingress namespaces")
	}

	if model.Gateways[0].Source == nil || model.Gateways[1].Source == nil {
		t.Fatal("expected gateway source ingress pointers to be set")
	}
}

func TestRenderSummary(t *testing.T) {
	model := Model{
		Gateways: []Gateway{
			{Name: "demo-gateway", Namespace: "default"},
		},
		HTTPRoutes: []HTTPRoute{
			{Name: "demo-route", Namespace: "default"},
		},
	}

	got := RenderSummary(model)
	want := "gateways:\n- default/demo-gateway\nhttpRoutes:\n- default/demo-route\n"

	if got != want {
		t.Fatalf("render summary = %q, want %q", got, want)
	}
}
