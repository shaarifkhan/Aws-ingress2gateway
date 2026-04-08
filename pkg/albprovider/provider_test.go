package albprovider

import (
	"strings"
	"testing"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	if !strings.Contains(summary, "loadBalancerConfigurations:\n") {
		t.Fatal("expected summary to include loadBalancerConfigurations header")
	}

	if !strings.Contains(summary, "targetGroupConfigurations:\n") {
		t.Fatal("expected summary to include targetGroupConfigurations header")
	}

	if !strings.Contains(summary, "- default/demo-one scheme= listeners=1\n") {
		t.Fatal("expected summary to include first gateway")
	}

	if !strings.Contains(summary, "- apps/demo-two scheme= listeners=1\n") {
		t.Fatal("expected summary to include second gateway")
	}

	if !strings.Contains(summary, "- default/demo-one hosts= parents=default/demo-one#http rules=0\n") {
		t.Fatal("expected summary to include first http route")
	}

	if !strings.Contains(summary, "- apps/demo-two hosts= parents=apps/demo-two#http rules=0\n") {
		t.Fatal("expected summary to include second http route")
	}
}

func TestBuildGatewayAPIResources(t *testing.T) {
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

	resources := provider.BuildGatewayAPIResources()

	if len(resources.Gateways) != 2 {
		t.Fatalf("got %d gateways, want 2", len(resources.Gateways))
	}

	if len(resources.HTTPRoutes) != 2 {
		t.Fatalf("got %d http routes, want 2", len(resources.HTTPRoutes))
	}
}

func TestBuildGatewayAPIResourcesAddsLoadBalancerParametersRef(t *testing.T) {
	provider := NewProvider()
	provider.storage.AddIngresses([]networkingv1.Ingress{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "demo",
				Namespace: "default",
				Annotations: map[string]string{
					"alb.ingress.kubernetes.io/scheme": "internet-facing",
				},
			},
		},
	})

	resources := provider.BuildGatewayAPIResources()

	if len(resources.Gateways) != 1 {
		t.Fatalf("got %d gateways, want 1", len(resources.Gateways))
	}

	if resources.Gateways[0].Spec.Infrastructure == nil || resources.Gateways[0].Spec.Infrastructure.ParametersRef == nil {
		t.Fatal("expected gateway infrastructure parametersRef to be set")
	}

	if resources.Gateways[0].Spec.Infrastructure.ParametersRef.Name != "demo-lb-config" {
		t.Fatalf("gateway parametersRef name = %q, want demo-lb-config", resources.Gateways[0].Spec.Infrastructure.ParametersRef.Name)
	}
}

func TestBuildGatewayAPIYAML(t *testing.T) {
	provider := NewProvider()
	provider.storage.AddIngresses([]networkingv1.Ingress{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "demo",
				Namespace: "default",
			},
		},
	})

	rendered, err := provider.BuildGatewayAPIYAML()
	if err != nil {
		t.Fatalf("build gateway api yaml returned error: %v", err)
	}

	if !strings.Contains(rendered, "kind: Gateway\n") {
		t.Fatal("expected rendered yaml to include a Gateway")
	}

	if !strings.Contains(rendered, "kind: HTTPRoute\n") {
		t.Fatal("expected rendered yaml to include an HTTPRoute")
	}

	if !strings.Contains(rendered, "name: demo\n") {
		t.Fatal("expected rendered yaml to include the ingress-derived name")
	}
}

func TestBuildAWSGatewayResources(t *testing.T) {
	provider := NewProvider()
	provider.storage.AddIngresses([]networkingv1.Ingress{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "demo",
				Namespace: "default",
				Annotations: map[string]string{
					"alb.ingress.kubernetes.io/load-balancer-name":       "demo-alb",
					"alb.ingress.kubernetes.io/scheme":                   "internet-facing",
					"alb.ingress.kubernetes.io/load-balancer-attributes": "idle_timeout.timeout_seconds=60",
					"alb.ingress.kubernetes.io/wafv2-acl-arn":            "arn:aws:wafv2:region:acct:regional/webacl/demo/123",
					"alb.ingress.kubernetes.io/target-type":              "ip",
					"alb.ingress.kubernetes.io/healthcheck-path":         "/healthz",
				},
			},
			Spec: networkingv1.IngressSpec{
				Rules: []networkingv1.IngressRule{
					{
						IngressRuleValue: networkingv1.IngressRuleValue{
							HTTP: &networkingv1.HTTPIngressRuleValue{
								Paths: []networkingv1.HTTPIngressPath{
									{
										Backend: networkingv1.IngressBackend{
											Service: &networkingv1.IngressServiceBackend{
												Name: "demo-service",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	})

	resources := provider.BuildAWSGatewayResources()

	if len(resources.LoadBalancerConfigurations) != 1 {
		t.Fatalf("got %d load balancer configs, want 1", len(resources.LoadBalancerConfigurations))
	}

	if len(resources.TargetGroupConfigurations) != 1 {
		t.Fatalf("got %d target group configs, want 1", len(resources.TargetGroupConfigurations))
	}

	if len(resources.LoadBalancerConfigurations[0].Spec.LoadBalancerAttributes) != 1 {
		t.Fatal("expected one load balancer attribute")
	}

	if resources.LoadBalancerConfigurations[0].Spec.WAFv2 == nil {
		t.Fatal("expected wafv2 configuration to be present")
	}
}

func TestBuildAWSGatewayYAML(t *testing.T) {
	provider := NewProvider()
	provider.storage.AddIngresses([]networkingv1.Ingress{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "demo",
				Namespace: "default",
				Annotations: map[string]string{
					"alb.ingress.kubernetes.io/scheme":                   "internet-facing",
					"alb.ingress.kubernetes.io/load-balancer-attributes": "idle_timeout.timeout_seconds=60",
					"alb.ingress.kubernetes.io/wafv2-acl-arn":            "arn:aws:wafv2:region:acct:regional/webacl/demo/123",
					"alb.ingress.kubernetes.io/target-type":              "ip",
					"alb.ingress.kubernetes.io/healthcheck-path":         "/healthz",
				},
			},
			Spec: networkingv1.IngressSpec{
				Rules: []networkingv1.IngressRule{
					{
						IngressRuleValue: networkingv1.IngressRuleValue{
							HTTP: &networkingv1.HTTPIngressRuleValue{
								Paths: []networkingv1.HTTPIngressPath{
									{
										Backend: networkingv1.IngressBackend{
											Service: &networkingv1.IngressServiceBackend{
												Name: "demo-service",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	})

	rendered, err := provider.BuildAWSGatewayYAML()
	if err != nil {
		t.Fatalf("build aws gateway yaml returned error: %v", err)
	}

	if !strings.Contains(rendered, "kind: LoadBalancerConfiguration\n") {
		t.Fatal("expected rendered yaml to include a LoadBalancerConfiguration")
	}

	if !strings.Contains(rendered, "kind: TargetGroupConfiguration\n") {
		t.Fatal("expected rendered yaml to include a TargetGroupConfiguration")
	}

	if !strings.Contains(rendered, "webACL: arn:aws:wafv2:region:acct:regional/webacl/demo/123\n") {
		t.Fatal("expected rendered yaml to include wafv2 webACL")
	}
}

func TestBuildCombinedYAML(t *testing.T) {
	provider := NewProvider()
	provider.storage.AddIngresses([]networkingv1.Ingress{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "demo",
				Namespace: "default",
				Annotations: map[string]string{
					"alb.ingress.kubernetes.io/scheme":           "internet-facing",
					"alb.ingress.kubernetes.io/target-type":      "ip",
					"alb.ingress.kubernetes.io/healthcheck-path": "/healthz",
				},
			},
			Spec: networkingv1.IngressSpec{
				Rules: []networkingv1.IngressRule{
					{
						IngressRuleValue: networkingv1.IngressRuleValue{
							HTTP: &networkingv1.HTTPIngressRuleValue{
								Paths: []networkingv1.HTTPIngressPath{
									{
										Backend: networkingv1.IngressBackend{
											Service: &networkingv1.IngressServiceBackend{
												Name: "demo-service",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	})

	rendered, err := provider.BuildCombinedYAML()
	if err != nil {
		t.Fatalf("build combined yaml returned error: %v", err)
	}

	if !strings.Contains(rendered, "kind: Gateway\n") {
		t.Fatal("expected combined yaml to include a Gateway")
	}

	if !strings.Contains(rendered, "kind: HTTPRoute\n") {
		t.Fatal("expected combined yaml to include an HTTPRoute")
	}

	if !strings.Contains(rendered, "kind: LoadBalancerConfiguration\n") {
		t.Fatal("expected combined yaml to include a LoadBalancerConfiguration")
	}

	if !strings.Contains(rendered, "kind: TargetGroupConfiguration\n") {
		t.Fatal("expected combined yaml to include a TargetGroupConfiguration")
	}
}

func TestFilterStoredIngressesByNameAcrossNamespaces(t *testing.T) {
	provider := NewProvider()
	provider.storage.AddIngresses([]networkingv1.Ingress{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "demo",
				Namespace: "default",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "other",
				Namespace: "apps",
			},
		},
	})

	provider.FilterStoredIngresses("", "other")

	ingresses := provider.Storage().ListIngresses()
	if len(ingresses) != 1 {
		t.Fatalf("got %d ingresses after filter, want 1", len(ingresses))
	}

	if ingresses[0].Namespace != "apps" || ingresses[0].Name != "other" {
		t.Fatalf("filtered ingress = %s/%s, want apps/other", ingresses[0].Namespace, ingresses[0].Name)
	}
}

func TestFilterStoredIngressesByNamespaceAndName(t *testing.T) {
	provider := NewProvider()
	provider.storage.AddIngresses([]networkingv1.Ingress{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "demo",
				Namespace: "default",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "demo",
				Namespace: "apps",
			},
		},
	})

	provider.FilterStoredIngresses("apps", "demo")

	ingresses := provider.Storage().ListIngresses()
	if len(ingresses) != 1 {
		t.Fatalf("got %d ingresses after filter, want 1", len(ingresses))
	}

	if ingresses[0].Namespace != "apps" || ingresses[0].Name != "demo" {
		t.Fatalf("filtered ingress = %s/%s, want apps/demo", ingresses[0].Namespace, ingresses[0].Name)
	}
}
