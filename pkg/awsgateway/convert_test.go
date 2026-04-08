package awsgateway

import (
	"testing"

	awsgatewayv1beta1 "sigs.k8s.io/aws-load-balancer-controller/apis/gateway/v1beta1"

	"aws-ingress2gateway/pkg/albir"
)

func TestConvertLoadBalancerConfiguration(t *testing.T) {
	config := albir.LoadBalancerConfiguration{
		Name:             "demo-lb-config",
		Namespace:        "default",
		LoadBalancerName: "demo-alb",
		Scheme:           "internet-facing",
		LoadBalancerAttributes: []albir.LoadBalancerAttribute{
			{Key: "idle_timeout.timeout_seconds", Value: "60"},
			{Key: "routing.http2.enabled", Value: "true"},
		},
		WAFv2ACLARN: "arn:aws:wafv2:region:acct:regional/webacl/demo/123",
		ListenerConfigurations: []albir.LoadBalancerListenerConfiguration{
			{
				Protocol:     "HTTPS",
				Port:         443,
				SSLPolicy:    "ELBSecurityPolicy-2016-08",
				Certificates: []string{"arn:one", "arn:two"},
			},
		},
	}

	typed := ConvertLoadBalancerConfiguration(config)

	if typed.Name != "demo-lb-config" || typed.Namespace != "default" {
		t.Fatalf("typed load balancer config identity = %s/%s, want default/demo-lb-config", typed.Namespace, typed.Name)
	}

	if typed.Spec.LoadBalancerName == nil || *typed.Spec.LoadBalancerName != "demo-alb" {
		t.Fatal("expected load balancer name to be preserved")
	}

	if typed.Spec.Scheme == nil || *typed.Spec.Scheme != awsgatewayv1beta1.LoadBalancerSchemeInternetFacing {
		t.Fatal("expected scheme internet-facing")
	}

	if len(typed.Spec.LoadBalancerAttributes) != 2 {
		t.Fatalf("got %d load balancer attributes, want 2", len(typed.Spec.LoadBalancerAttributes))
	}

	if typed.Spec.LoadBalancerAttributes[0].Key != "idle_timeout.timeout_seconds" || typed.Spec.LoadBalancerAttributes[0].Value != "60" {
		t.Fatalf("first load balancer attribute = %#v, want idle_timeout.timeout_seconds=60", typed.Spec.LoadBalancerAttributes[0])
	}

	if typed.Spec.WAFv2 == nil || typed.Spec.WAFv2.ACL != "arn:aws:wafv2:region:acct:regional/webacl/demo/123" {
		t.Fatal("expected wafv2 webACL to be preserved")
	}

	if typed.Spec.ListenerConfigurations == nil || len(*typed.Spec.ListenerConfigurations) != 1 {
		t.Fatal("expected one listener configuration")
	}

	listener := (*typed.Spec.ListenerConfigurations)[0]
	if listener.ProtocolPort != "HTTPS:443" {
		t.Fatalf("protocolPort = %q, want HTTPS:443", listener.ProtocolPort)
	}

	if listener.DefaultCertificate == nil || *listener.DefaultCertificate != "arn:one" {
		t.Fatal("expected first certificate to become defaultCertificate")
	}

	if len(listener.Certificates) != 1 || listener.Certificates[0] == nil || *listener.Certificates[0] != "arn:two" {
		t.Fatal("expected remaining certificates to stay on certificates")
	}

	if listener.SslPolicy == nil || *listener.SslPolicy != "ELBSecurityPolicy-2016-08" {
		t.Fatal("expected ssl policy to be preserved")
	}
}

func TestConvertTargetGroupConfiguration(t *testing.T) {
	config := albir.TargetGroupConfiguration{
		Name:            "demo-service-tg-config",
		Namespace:       "default",
		ServiceName:     "demo-service",
		TargetType:      "ip",
		HealthCheckPath: "/healthz",
	}

	typed := ConvertTargetGroupConfiguration(config)

	if typed.Name != "demo-service-tg-config" || typed.Namespace != "default" {
		t.Fatalf("typed target group config identity = %s/%s, want default/demo-service-tg-config", typed.Namespace, typed.Name)
	}

	if typed.Spec.TargetReference.Name != "demo-service" {
		t.Fatalf("target reference name = %q, want demo-service", typed.Spec.TargetReference.Name)
	}

	if typed.Spec.DefaultConfiguration.TargetType == nil || *typed.Spec.DefaultConfiguration.TargetType != awsgatewayv1beta1.TargetTypeIP {
		t.Fatal("expected target type ip")
	}

	if typed.Spec.DefaultConfiguration.HealthCheckConfig == nil || typed.Spec.DefaultConfiguration.HealthCheckConfig.HealthCheckPath == nil {
		t.Fatal("expected health check config path to be present")
	}

	if *typed.Spec.DefaultConfiguration.HealthCheckConfig.HealthCheckPath != "/healthz" {
		t.Fatalf("health check path = %q, want /healthz", *typed.Spec.DefaultConfiguration.HealthCheckConfig.HealthCheckPath)
	}
}

func TestConvertModel(t *testing.T) {
	model := albir.Model{
		LoadBalancerConfigurations: []albir.LoadBalancerConfiguration{
			{Name: "demo-lb-config", Namespace: "default"},
		},
		TargetGroupConfigurations: []albir.TargetGroupConfiguration{
			{Name: "demo-tg-config", Namespace: "default", ServiceName: "demo-service"},
		},
	}

	resources := ConvertModel(model)

	if len(resources.LoadBalancerConfigurations) != 1 {
		t.Fatalf("got %d load balancer configs, want 1", len(resources.LoadBalancerConfigurations))
	}

	if len(resources.TargetGroupConfigurations) != 1 {
		t.Fatalf("got %d target group configs, want 1", len(resources.TargetGroupConfigurations))
	}
}
