package albir

import (
	"testing"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestConvertIngress(t *testing.T) {
	pathType := networkingv1.PathTypePrefix
	ingress := networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo",
			Namespace: "default",
			Annotations: map[string]string{
				ALBSchemeAnnotation:     "internet-facing",
				ALBTargetTypeAnnotation: "ip",
			},
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: "demo.example.com",
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: "demo-service",
											Port: networkingv1.ServiceBackendPort{
												Number: 80,
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
	}

	model := ConvertIngress(ingress)

	if len(model.Gateways) != 1 {
		t.Fatalf("got %d gateways, want 1", len(model.Gateways))
	}

	if len(model.HTTPRoutes) != 1 {
		t.Fatalf("got %d http routes, want 1", len(model.HTTPRoutes))
	}

	if len(model.LoadBalancerConfigurations) != 1 {
		t.Fatalf("got %d load balancer configurations, want 1", len(model.LoadBalancerConfigurations))
	}

	if model.LoadBalancerConfigurations[0].Scheme != "internet-facing" {
		t.Fatalf("load balancer config scheme = %q, want internet-facing", model.LoadBalancerConfigurations[0].Scheme)
	}

	gateway := model.Gateways[0]
	if gateway.Name != ingress.Name || gateway.Namespace != ingress.Namespace {
		t.Fatalf("gateway identity = %s/%s, want %s/%s", gateway.Namespace, gateway.Name, ingress.Namespace, ingress.Name)
	}

	if gateway.Scheme != "internet-facing" {
		t.Fatalf("gateway scheme = %q, want internet-facing", gateway.Scheme)
	}

	if len(gateway.Listeners) != 1 {
		t.Fatalf("got %d gateway listeners, want 1", len(gateway.Listeners))
	}

	listener := gateway.Listeners[0]
	if listener.Name != "http" || listener.Protocol != "HTTP" || listener.Port != 80 || listener.Hostname != "demo.example.com" {
		t.Fatalf("listener = %#v, want http/HTTP/80/demo.example.com", listener)
	}

	route := model.HTTPRoutes[0]
	if route.Name != ingress.Name || route.Namespace != ingress.Namespace {
		t.Fatalf("http route identity = %s/%s, want %s/%s", route.Namespace, route.Name, ingress.Namespace, ingress.Name)
	}

	if len(route.Hostnames) != 1 || route.Hostnames[0] != "demo.example.com" {
		t.Fatalf("route hostnames = %#v, want [demo.example.com]", route.Hostnames)
	}

	if len(route.ParentRefs) != 1 {
		t.Fatalf("got %d parent refs, want 1", len(route.ParentRefs))
	}

	parentRef := route.ParentRefs[0]
	if parentRef.Namespace != "default" || parentRef.GatewayName != "demo" || parentRef.SectionName != "http" {
		t.Fatalf("parent ref = %#v, want default/demo#http", parentRef)
	}

	if len(route.Rules) != 1 {
		t.Fatalf("got %d route rules, want 1", len(route.Rules))
	}

	if route.Rules[0].Path != "/" {
		t.Fatalf("route rule path = %q, want /", route.Rules[0].Path)
	}

	if route.Rules[0].PathType == nil || *route.Rules[0].PathType != networkingv1.PathTypePrefix {
		t.Fatal("expected route rule path type to be preserved")
	}

	if len(route.Rules[0].BackendRefs) != 1 {
		t.Fatalf("got %d backend refs, want 1", len(route.Rules[0].BackendRefs))
	}

	backendRef := route.Rules[0].BackendRefs[0]
	if backendRef.Name != "demo-service" || backendRef.PortNumber != 80 {
		t.Fatalf("backend ref = %#v, want demo-service:80", backendRef)
	}

	if backendRef.TargetType != "ip" {
		t.Fatalf("backend target type = %q, want ip", backendRef.TargetType)
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

	if len(model.LoadBalancerConfigurations) != 0 {
		t.Fatalf("got %d load balancer configurations, want 0", len(model.LoadBalancerConfigurations))
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

func TestConvertIngressCollectsLoadBalancerConfiguration(t *testing.T) {
	ingress := networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo",
			Namespace: "default",
			Annotations: map[string]string{
				ALBLoadBalancerNameAnnotation: "demo-alb",
				ALBSchemeAnnotation:           "internet-facing",
				ALBListenPortsAnnotation:      `[{"HTTP":8080},{"HTTPS":8443}]`,
				ALBCertificateARNAnnotation:   "arn:aws:acm:region:acct:cert/one, arn:aws:acm:region:acct:cert/two",
				ALBSSLPolicyAnnotation:        "ELBSecurityPolicy-TLS13-1-2-2021-06",
			},
		},
	}

	model := ConvertIngress(ingress)

	if len(model.LoadBalancerConfigurations) != 1 {
		t.Fatalf("got %d load balancer configurations, want 1", len(model.LoadBalancerConfigurations))
	}

	config := model.LoadBalancerConfigurations[0]
	if config.Name != "demo-lb-config" || config.Namespace != "default" {
		t.Fatalf("config identity = %s/%s, want default/demo-lb-config", config.Namespace, config.Name)
	}

	if config.LoadBalancerName != "demo-alb" {
		t.Fatalf("load balancer name = %q, want demo-alb", config.LoadBalancerName)
	}

	if config.Scheme != "internet-facing" {
		t.Fatalf("scheme = %q, want internet-facing", config.Scheme)
	}

	if len(config.Listeners) != 2 {
		t.Fatalf("got %d load balancer listeners, want 2", len(config.Listeners))
	}

	if config.Listeners[0].Protocol != "HTTP" || config.Listeners[0].Port != 8080 {
		t.Fatalf("first lb listener = %#v, want HTTP:8080", config.Listeners[0])
	}

	if config.Listeners[1].Protocol != "HTTPS" || config.Listeners[1].Port != 8443 {
		t.Fatalf("second lb listener = %#v, want HTTPS:8443", config.Listeners[1])
	}

	if config.Listeners[1].SSLPolicy != "ELBSecurityPolicy-TLS13-1-2-2021-06" {
		t.Fatalf("ssl policy = %q, want ELBSecurityPolicy-TLS13-1-2-2021-06", config.Listeners[1].SSLPolicy)
	}

	if len(config.Listeners[1].Certificates) != 2 {
		t.Fatalf("got %d certificates, want 2", len(config.Listeners[1].Certificates))
	}
}

func TestConvertIngressCollectsInternalScheme(t *testing.T) {
	ingress := networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo",
			Namespace: "default",
			Annotations: map[string]string{
				ALBSchemeAnnotation: "internal",
			},
		},
	}

	model := ConvertIngress(ingress)

	if len(model.Gateways) != 1 {
		t.Fatalf("got %d gateways, want 1", len(model.Gateways))
	}

	if model.Gateways[0].Scheme != "internal" {
		t.Fatalf("gateway scheme = %q, want internal", model.Gateways[0].Scheme)
	}
}

func TestConvertIngressIgnoresUnknownScheme(t *testing.T) {
	ingress := networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo",
			Namespace: "default",
			Annotations: map[string]string{
				ALBSchemeAnnotation: "something-else",
			},
		},
	}

	model := ConvertIngress(ingress)

	if len(model.Gateways) != 1 {
		t.Fatalf("got %d gateways, want 1", len(model.Gateways))
	}

	if model.Gateways[0].Scheme != "" {
		t.Fatalf("gateway scheme = %q, want empty string", model.Gateways[0].Scheme)
	}
}

func TestConvertIngressCollectsInstanceTargetType(t *testing.T) {
	pathType := networkingv1.PathTypePrefix
	ingress := networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo",
			Namespace: "default",
			Annotations: map[string]string{
				ALBTargetTypeAnnotation: "instance",
			},
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: "demo.example.com",
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: "demo-service",
											Port: networkingv1.ServiceBackendPort{
												Number: 80,
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
	}

	model := ConvertIngress(ingress)

	if len(model.HTTPRoutes) != 1 || len(model.HTTPRoutes[0].Rules) != 1 || len(model.HTTPRoutes[0].Rules[0].BackendRefs) != 1 {
		t.Fatal("expected one backend ref in converted route")
	}

	if model.HTTPRoutes[0].Rules[0].BackendRefs[0].TargetType != "instance" {
		t.Fatalf("backend target type = %q, want instance", model.HTTPRoutes[0].Rules[0].BackendRefs[0].TargetType)
	}
}

func TestConvertIngressIgnoresUnknownTargetType(t *testing.T) {
	pathType := networkingv1.PathTypePrefix
	ingress := networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo",
			Namespace: "default",
			Annotations: map[string]string{
				ALBTargetTypeAnnotation: "something-else",
			},
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: "demo.example.com",
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: "demo-service",
											Port: networkingv1.ServiceBackendPort{
												Number: 80,
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
	}

	model := ConvertIngress(ingress)

	if len(model.HTTPRoutes) != 1 || len(model.HTTPRoutes[0].Rules) != 1 || len(model.HTTPRoutes[0].Rules[0].BackendRefs) != 1 {
		t.Fatal("expected one backend ref in converted route")
	}

	if model.HTTPRoutes[0].Rules[0].BackendRefs[0].TargetType != "" {
		t.Fatalf("backend target type = %q, want empty string", model.HTTPRoutes[0].Rules[0].BackendRefs[0].TargetType)
	}
}

func TestConvertIngressAddsHTTPSListenerForTLSHost(t *testing.T) {
	pathType := networkingv1.PathTypePrefix
	ingress := networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo",
			Namespace: "default",
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: "demo.example.com",
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: "demo-service",
											Port: networkingv1.ServiceBackendPort{
												Number: 80,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			TLS: []networkingv1.IngressTLS{
				{
					Hosts: []string{"demo.example.com"},
				},
			},
		},
	}

	model := ConvertIngress(ingress)

	if len(model.Gateways) != 1 {
		t.Fatalf("got %d gateways, want 1", len(model.Gateways))
	}

	if len(model.Gateways[0].Listeners) != 2 {
		t.Fatalf("got %d listeners, want 2", len(model.Gateways[0].Listeners))
	}

	httpListener := model.Gateways[0].Listeners[0]
	httpsListener := model.Gateways[0].Listeners[1]

	if httpListener.Name != "http" || httpListener.Protocol != "HTTP" || httpListener.Port != 80 {
		t.Fatalf("http listener = %#v, want HTTP on port 80", httpListener)
	}

	if httpsListener.Name != "https" || httpsListener.Protocol != "HTTPS" || httpsListener.Port != 443 {
		t.Fatalf("https listener = %#v, want HTTPS on port 443", httpsListener)
	}

	if len(model.HTTPRoutes) != 1 {
		t.Fatalf("got %d http routes, want 1", len(model.HTTPRoutes))
	}

	if len(model.HTTPRoutes[0].ParentRefs) != 2 {
		t.Fatalf("got %d parent refs, want 2", len(model.HTTPRoutes[0].ParentRefs))
	}

	if model.HTTPRoutes[0].ParentRefs[0].SectionName != "http" || model.HTTPRoutes[0].ParentRefs[1].SectionName != "https" {
		t.Fatalf("parent refs = %#v, want http and https sections", model.HTTPRoutes[0].ParentRefs)
	}
}

func TestConvertIngressUsesALBListenPortsAnnotation(t *testing.T) {
	pathType := networkingv1.PathTypePrefix
	ingress := networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo",
			Namespace: "default",
			Annotations: map[string]string{
				ALBListenPortsAnnotation: `[{"HTTP":8080},{"HTTPS":8443}]`,
			},
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: "demo.example.com",
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: "demo-service",
											Port: networkingv1.ServiceBackendPort{
												Number: 80,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			TLS: []networkingv1.IngressTLS{
				{
					Hosts: []string{"demo.example.com"},
				},
			},
		},
	}

	model := ConvertIngress(ingress)

	if len(model.Gateways) != 1 {
		t.Fatalf("got %d gateways, want 1", len(model.Gateways))
	}

	if len(model.Gateways[0].Listeners) != 2 {
		t.Fatalf("got %d listeners, want 2", len(model.Gateways[0].Listeners))
	}

	if model.Gateways[0].Listeners[0].Port != 8080 || model.Gateways[0].Listeners[0].Protocol != "HTTP" {
		t.Fatalf("first listener = %#v, want HTTP:8080", model.Gateways[0].Listeners[0])
	}

	if model.Gateways[0].Listeners[1].Port != 8443 || model.Gateways[0].Listeners[1].Protocol != "HTTPS" {
		t.Fatalf("second listener = %#v, want HTTPS:8443", model.Gateways[0].Listeners[1])
	}

	if len(model.HTTPRoutes[0].ParentRefs) != 2 {
		t.Fatalf("got %d parent refs, want 2", len(model.HTTPRoutes[0].ParentRefs))
	}
}

func TestConvertIngressUsesALBListenPortsAnnotationWithoutTLS(t *testing.T) {
	pathType := networkingv1.PathTypePrefix
	ingress := networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo",
			Namespace: "default",
			Annotations: map[string]string{
				ALBListenPortsAnnotation: `[{"HTTP":8080},{"HTTPS":8443}]`,
			},
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: "demo.example.com",
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: "demo-service",
											Port: networkingv1.ServiceBackendPort{
												Number: 80,
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
	}

	model := ConvertIngress(ingress)

	if len(model.Gateways[0].Listeners) != 2 {
		t.Fatalf("got %d listeners, want 2", len(model.Gateways[0].Listeners))
	}

	if model.Gateways[0].Listeners[1].Protocol != "HTTPS" || model.Gateways[0].Listeners[1].Port != 8443 {
		t.Fatalf("second listener = %#v, want HTTPS:8443", model.Gateways[0].Listeners[1])
	}
}

func TestConvertIngressAddsCatchAllListenerWithoutHost(t *testing.T) {
	pathType := networkingv1.PathTypePrefix
	ingress := networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo",
			Namespace: "default",
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: "demo-service",
											Port: networkingv1.ServiceBackendPort{
												Number: 80,
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
	}

	model := ConvertIngress(ingress)

	if len(model.Gateways) != 1 {
		t.Fatalf("got %d gateways, want 1", len(model.Gateways))
	}

	if len(model.Gateways[0].Listeners) != 1 {
		t.Fatalf("got %d listeners, want 1", len(model.Gateways[0].Listeners))
	}

	listener := model.Gateways[0].Listeners[0]
	if listener.Name != "http" || listener.Hostname != "" || listener.Protocol != "HTTP" || listener.Port != 80 {
		t.Fatalf("listener = %#v, want catch-all HTTP listener on port 80", listener)
	}

	if len(model.HTTPRoutes) != 1 {
		t.Fatalf("got %d http routes, want 1", len(model.HTTPRoutes))
	}

	if len(model.HTTPRoutes[0].ParentRefs) != 1 {
		t.Fatalf("got %d parent refs, want 1", len(model.HTTPRoutes[0].ParentRefs))
	}

	if model.HTTPRoutes[0].ParentRefs[0].SectionName != "http" {
		t.Fatalf("parent ref section = %q, want http", model.HTTPRoutes[0].ParentRefs[0].SectionName)
	}
}

func TestConvertIngressAddsCatchAllHTTPSListenerWhenTLSExists(t *testing.T) {
	pathType := networkingv1.PathTypePrefix
	ingress := networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo",
			Namespace: "default",
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: "demo-service",
											Port: networkingv1.ServiceBackendPort{
												Number: 80,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			TLS: []networkingv1.IngressTLS{
				{
					SecretName: "demo-cert",
				},
			},
		},
	}

	model := ConvertIngress(ingress)

	if len(model.Gateways) != 1 {
		t.Fatalf("got %d gateways, want 1", len(model.Gateways))
	}

	if len(model.Gateways[0].Listeners) != 2 {
		t.Fatalf("got %d listeners, want 2", len(model.Gateways[0].Listeners))
	}

	if model.Gateways[0].Listeners[0].Name != "http" || model.Gateways[0].Listeners[1].Name != "https" {
		t.Fatalf("listeners = %#v, want catch-all http and https", model.Gateways[0].Listeners)
	}

	if len(model.HTTPRoutes[0].ParentRefs) != 2 {
		t.Fatalf("got %d parent refs, want 2", len(model.HTTPRoutes[0].ParentRefs))
	}
}

func TestConvertIngressUsesALBListenPortsAnnotationWithoutHosts(t *testing.T) {
	pathType := networkingv1.PathTypePrefix
	ingress := networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo",
			Namespace: "default",
			Annotations: map[string]string{
				ALBListenPortsAnnotation: `[{"HTTP":8080},{"HTTPS":8443}]`,
			},
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: "demo-service",
											Port: networkingv1.ServiceBackendPort{
												Number: 80,
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
	}

	model := ConvertIngress(ingress)

	if len(model.Gateways) != 1 {
		t.Fatalf("got %d gateways, want 1", len(model.Gateways))
	}

	if len(model.Gateways[0].Listeners) != 2 {
		t.Fatalf("got %d listeners, want 2", len(model.Gateways[0].Listeners))
	}

	if model.Gateways[0].Listeners[0].Port != 8080 || model.Gateways[0].Listeners[1].Port != 8443 {
		t.Fatalf("listeners = %#v, want annotated ports 8080 and 8443", model.Gateways[0].Listeners)
	}
}

func TestRenderSummary(t *testing.T) {
	model := Model{
		Gateways: []Gateway{
			{
				Name:      "demo-gateway",
				Namespace: "default",
				Scheme:    "internet-facing",
				Listeners: []Listener{
					{
						Name:     "http",
						Protocol: "HTTP",
						Port:     80,
						Hostname: "demo.example.com",
					},
					{
						Name:     "https",
						Protocol: "HTTPS",
						Port:     443,
						Hostname: "demo.example.com",
					},
				},
			},
		},
		HTTPRoutes: []HTTPRoute{
			{
				Name:      "demo-route",
				Namespace: "default",
				Hostnames: []string{"demo.example.com"},
				ParentRefs: []ParentRef{
					{
						Namespace:   "default",
						GatewayName: "demo-gateway",
						SectionName: "http",
					},
					{
						Namespace:   "default",
						GatewayName: "demo-gateway",
						SectionName: "https",
					},
				},
				Rules: []HTTPRouteRule{
					{
						Path: "/",
						BackendRefs: []BackendRef{
							{Name: "demo-service", PortNumber: 80, TargetType: "ip"},
						},
					},
				},
			},
		},
		LoadBalancerConfigurations: []LoadBalancerConfiguration{
			{
				Name:             "demo-lb-config",
				Namespace:        "default",
				LoadBalancerName: "demo-alb",
				Scheme:           "internet-facing",
				Listeners: []LoadBalancerListenerConfiguration{
					{
						Protocol: "HTTP",
						Port:     80,
					},
					{
						Protocol:     "HTTPS",
						Port:         443,
						SSLPolicy:    "ELBSecurityPolicy-2016-08",
						Certificates: []string{"arn:one", "arn:two"},
					},
				},
			},
		},
	}

	got := RenderSummary(model)
	want := "gateways:\n- default/demo-gateway scheme=internet-facing listeners=2\n  listener=http protocol=HTTP port=80 hostname=demo.example.com\n  listener=https protocol=HTTPS port=443 hostname=demo.example.com\nhttpRoutes:\n- default/demo-route hosts=demo.example.com parents=default/demo-gateway#http,default/demo-gateway#https rules=1\n  path=/ backend=demo-service:80 targetType=ip\nloadBalancerConfigurations:\n- default/demo-lb-config loadBalancerName=demo-alb scheme=internet-facing listeners=2\n  listener protocol=HTTP port=80 sslPolicy= certificates=\n  listener protocol=HTTPS port=443 sslPolicy=ELBSecurityPolicy-2016-08 certificates=arn:one,arn:two\n"

	if got != want {
		t.Fatalf("render summary = %q, want %q", got, want)
	}
}
