package awsgateway

import (
	"strings"
	"testing"

	"aws-ingress2gateway/pkg/albir"
)

func TestRenderResourcesYAML(t *testing.T) {
	resources := ConvertModel(albir.Model{
		LoadBalancerConfigurations: []albir.LoadBalancerConfiguration{
			{
				Name:             "demo-lb-config",
				Namespace:        "default",
				LoadBalancerName: "demo-alb",
				Scheme:           "internet-facing",
				LoadBalancerAttributes: []albir.LoadBalancerAttribute{
					{Key: "idle_timeout.timeout_seconds", Value: "60"},
				},
				WAFv2ACLARN: "arn:aws:wafv2:region:acct:regional/webacl/demo/123",
			},
		},
		TargetGroupConfigurations: []albir.TargetGroupConfiguration{
			{
				Name:            "demo-service-tg-config",
				Namespace:       "default",
				ServiceName:     "demo-service",
				TargetType:      "ip",
				HealthCheckPath: "/healthz",
			},
		},
	})

	rendered, err := RenderResourcesYAML(resources)
	if err != nil {
		t.Fatalf("render resources yaml returned error: %v", err)
	}

	if !strings.Contains(rendered, "apiVersion: gateway.k8s.aws/v1beta1\nkind: LoadBalancerConfiguration\n") {
		t.Fatal("expected rendered yaml to include a LoadBalancerConfiguration document")
	}

	if !strings.Contains(rendered, "apiVersion: gateway.k8s.aws/v1beta1\nkind: TargetGroupConfiguration\n") {
		t.Fatal("expected rendered yaml to include a TargetGroupConfiguration document")
	}

	if !strings.Contains(rendered, "name: demo-lb-config\n") {
		t.Fatal("expected rendered yaml to include load balancer config name")
	}

	if !strings.Contains(rendered, "webACL: arn:aws:wafv2:region:acct:regional/webacl/demo/123\n") {
		t.Fatal("expected rendered yaml to include wafv2 webACL")
	}

	if !strings.Contains(rendered, "key: idle_timeout.timeout_seconds\n") || !strings.Contains(rendered, "value: \"60\"\n") {
		t.Fatal("expected rendered yaml to include load balancer attributes")
	}

	if !strings.Contains(rendered, "name: demo-service-tg-config\n") {
		t.Fatal("expected rendered yaml to include target group config name")
	}

	if !strings.Contains(rendered, "healthCheckPath: /healthz\n") {
		t.Fatal("expected rendered yaml to include target group health check path")
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
		LoadBalancerConfigurations: []albir.LoadBalancerConfiguration{
			{Name: "demo-lb-config", Namespace: "default"},
		},
		TargetGroupConfigurations: []albir.TargetGroupConfiguration{
			{Name: "demo-service-tg-config", Namespace: "default", ServiceName: "demo-service"},
		},
	}

	rendered, err := RenderModelYAML(model)
	if err != nil {
		t.Fatalf("render model yaml returned error: %v", err)
	}

	if !strings.Contains(rendered, "kind: LoadBalancerConfiguration\n") || !strings.Contains(rendered, "kind: TargetGroupConfiguration\n") {
		t.Fatal("expected rendered model yaml to include both AWS customization CRDs")
	}
}
