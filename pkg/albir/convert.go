package albir

import (
	"encoding/json"
	"strconv"
	"strings"

	networkingv1 "k8s.io/api/networking/v1"
)

const ALBListenPortsAnnotation = "alb.ingress.kubernetes.io/listen-ports"
const ALBSchemeAnnotation = "alb.ingress.kubernetes.io/scheme"
const ALBTargetTypeAnnotation = "alb.ingress.kubernetes.io/target-type"
const ALBHealthCheckPathAnnotation = "alb.ingress.kubernetes.io/healthcheck-path"
const ALBLoadBalancerNameAnnotation = "alb.ingress.kubernetes.io/load-balancer-name"
const ALBCertificateARNAnnotation = "alb.ingress.kubernetes.io/certificate-arn"
const ALBSSLPolicyAnnotation = "alb.ingress.kubernetes.io/ssl-policy"
const ALBLoadBalancerAttributesAnnotation = "alb.ingress.kubernetes.io/load-balancer-attributes"
const ALBWAFv2ACLARNAnnotation = "alb.ingress.kubernetes.io/wafv2-acl-arn"

type listenerConfig struct {
	Protocol string
	Port     int32
}

// ConvertIngress builds the smallest useful IR from one ALB-backed Ingress.
// It does not validate ALB-specific behavior yet; it only preserves identity
// and a link back to the source object for later conversion steps.
func ConvertIngress(ingress networkingv1.Ingress) Model {
	ingressCopy := ingress
	hostnames := collectHostnames(ingress)
	listeners := collectListeners(ingress, hostnames)
	loadBalancerConfiguration := collectLoadBalancerConfiguration(ingress)
	targetGroupConfigurations := collectTargetGroupConfigurations(ingress)
	httpRoute := HTTPRoute{
		Name:       ingress.Name,
		Namespace:  ingress.Namespace,
		Hostnames:  hostnames,
		ParentRefs: collectParentRefs(ingress, listeners),
		Rules:      collectRouteRules(ingress),
		Source:     &ingressCopy,
	}

	model := Model{
		Gateways: []Gateway{
			{
				Name:      ingress.Name,
				Namespace: ingress.Namespace,
				Scheme:    collectScheme(ingress),
				Listeners: listeners,
				Source:    &ingressCopy,
			},
		},
		HTTPRoutes: []HTTPRoute{
			httpRoute,
		},
	}

	if loadBalancerConfiguration != nil {
		model.LoadBalancerConfigurations = []LoadBalancerConfiguration{*loadBalancerConfiguration}
	}

	if len(targetGroupConfigurations) > 0 {
		model.TargetGroupConfigurations = targetGroupConfigurations
	}

	return model
}

// ConvertIngresses converts a slice of Ingress objects into one combined model.
func ConvertIngresses(ingresses []networkingv1.Ingress) Model {
	model := Model{
		Gateways:                   make([]Gateway, 0, len(ingresses)),
		HTTPRoutes:                 make([]HTTPRoute, 0, len(ingresses)),
		LoadBalancerConfigurations: make([]LoadBalancerConfiguration, 0, len(ingresses)),
		TargetGroupConfigurations:  make([]TargetGroupConfiguration, 0, len(ingresses)),
	}

	for _, ingress := range ingresses {
		converted := ConvertIngress(ingress)
		model.Gateways = append(model.Gateways, converted.Gateways...)
		model.HTTPRoutes = append(model.HTTPRoutes, converted.HTTPRoutes...)
		model.LoadBalancerConfigurations = append(model.LoadBalancerConfigurations, converted.LoadBalancerConfigurations...)
		model.TargetGroupConfigurations = append(model.TargetGroupConfigurations, converted.TargetGroupConfigurations...)
	}

	return model
}

func collectLoadBalancerConfiguration(ingress networkingv1.Ingress) *LoadBalancerConfiguration {
	if !hasLoadBalancerConfigurationInput(ingress) {
		return nil
	}

	ingressCopy := ingress
	return &LoadBalancerConfiguration{
		Name:                   ingress.Name + "-lb-config",
		Namespace:              ingress.Namespace,
		LoadBalancerName:       strings.TrimSpace(ingress.Annotations[ALBLoadBalancerNameAnnotation]),
		Scheme:                 collectScheme(ingress),
		LoadBalancerAttributes: collectLoadBalancerAttributes(ingress),
		WAFv2ACLARN:            strings.TrimSpace(ingress.Annotations[ALBWAFv2ACLARNAnnotation]),
		ListenerConfigurations: collectLoadBalancerListenerConfigurations(ingress),
		Source:                 &ingressCopy,
	}
}

func hasLoadBalancerConfigurationInput(ingress networkingv1.Ingress) bool {
	return strings.TrimSpace(ingress.Annotations[ALBLoadBalancerNameAnnotation]) != "" ||
		strings.TrimSpace(ingress.Annotations[ALBSchemeAnnotation]) != "" ||
		strings.TrimSpace(ingress.Annotations[ALBCertificateARNAnnotation]) != "" ||
		strings.TrimSpace(ingress.Annotations[ALBSSLPolicyAnnotation]) != "" ||
		strings.TrimSpace(ingress.Annotations[ALBLoadBalancerAttributesAnnotation]) != "" ||
		strings.TrimSpace(ingress.Annotations[ALBWAFv2ACLARNAnnotation]) != ""
}

func collectLoadBalancerListenerConfigurations(ingress networkingv1.Ingress) []LoadBalancerListenerConfiguration {
	if strings.TrimSpace(ingress.Annotations[ALBCertificateARNAnnotation]) == "" &&
		strings.TrimSpace(ingress.Annotations[ALBSSLPolicyAnnotation]) == "" {
		return nil
	}

	configs := collectListenerConfigs(ingress)
	certificates := collectCertificateARNs(ingress)
	sslPolicy := strings.TrimSpace(ingress.Annotations[ALBSSLPolicyAnnotation])

	listeners := make([]LoadBalancerListenerConfiguration, 0, len(configs))
	for _, config := range configs {
		if config.Protocol != "HTTPS" {
			continue
		}

		listener := LoadBalancerListenerConfiguration{
			Protocol: config.Protocol,
			Port:     config.Port,
		}

		listener.SSLPolicy = sslPolicy
		listener.Certificates = certificates

		listeners = append(listeners, listener)
	}

	return listeners
}

func collectTargetGroupConfigurations(ingress networkingv1.Ingress) []TargetGroupConfiguration {
	targetType := collectTargetType(ingress)
	healthCheckPath := strings.TrimSpace(ingress.Annotations[ALBHealthCheckPathAnnotation])
	if targetType == "" && healthCheckPath == "" {
		return nil
	}

	serviceNames := collectBackendServiceNames(ingress)
	configurations := make([]TargetGroupConfiguration, 0, len(serviceNames))

	for _, serviceName := range serviceNames {
		ingressCopy := ingress
		configurations = append(configurations, TargetGroupConfiguration{
			Name:            serviceName + "-tg-config",
			Namespace:       ingress.Namespace,
			ServiceName:     serviceName,
			TargetType:      targetType,
			HealthCheckPath: healthCheckPath,
			Source:          &ingressCopy,
		})
	}

	return configurations
}

func collectBackendServiceNames(ingress networkingv1.Ingress) []string {
	serviceNames := make([]string, 0)
	seen := make(map[string]struct{})

	for _, ingressRule := range ingress.Spec.Rules {
		if ingressRule.HTTP == nil {
			continue
		}

		for _, path := range ingressRule.HTTP.Paths {
			if path.Backend.Service == nil || path.Backend.Service.Name == "" {
				continue
			}

			serviceName := path.Backend.Service.Name
			if _, ok := seen[serviceName]; ok {
				continue
			}

			serviceNames = append(serviceNames, serviceName)
			seen[serviceName] = struct{}{}
		}
	}

	return serviceNames
}

func collectCertificateARNs(ingress networkingv1.Ingress) []string {
	value := strings.TrimSpace(ingress.Annotations[ALBCertificateARNAnnotation])
	if value == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	certificates := make([]string, 0, len(parts))
	for _, part := range parts {
		certificate := strings.TrimSpace(part)
		if certificate == "" {
			continue
		}
		certificates = append(certificates, certificate)
	}

	return certificates
}

func collectLoadBalancerAttributes(ingress networkingv1.Ingress) []LoadBalancerAttribute {
	value := strings.TrimSpace(ingress.Annotations[ALBLoadBalancerAttributesAnnotation])
	if value == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	attributes := make([]LoadBalancerAttribute, 0, len(parts))
	for _, part := range parts {
		entry := strings.TrimSpace(part)
		if entry == "" {
			continue
		}

		keyValue := strings.SplitN(entry, "=", 2)
		if len(keyValue) != 2 {
			continue
		}

		key := strings.TrimSpace(keyValue[0])
		if key == "" {
			continue
		}

		attributes = append(attributes, LoadBalancerAttribute{
			Key:   key,
			Value: strings.TrimSpace(keyValue[1]),
		})
	}

	return attributes
}

func collectParentRefs(ingress networkingv1.Ingress, listeners []Listener) []ParentRef {
	parentRefs := make([]ParentRef, 0, len(listeners))

	for _, listener := range listeners {
		parentRefs = append(parentRefs, ParentRef{
			GatewayName: ingress.Name,
			SectionName: listener.Name,
			Namespace:   ingress.Namespace,
		})
	}

	return parentRefs
}

func collectListeners(ingress networkingv1.Ingress, hostnames []string) []Listener {
	configs := collectListenerConfigs(ingress)
	hasExplicitListenPorts := ingress.Annotations[ALBListenPortsAnnotation] != ""

	if len(hostnames) == 0 {
		listeners := make([]Listener, 0, len(configs))
		indexes := map[string]int{}

		for _, config := range configs {
			prefix := strings.ToLower(config.Protocol)
			listeners = append(listeners, Listener{
				Name:     listenerName(prefix, indexes[prefix]),
				Port:     config.Port,
				Protocol: config.Protocol,
			})
			indexes[prefix]++
		}

		return listeners
	}

	listeners := make([]Listener, 0, len(hostnames)*len(configs))
	indexes := map[string]int{}

	for _, hostname := range hostnames {
		for _, config := range configs {
			if !listenerConfigAppliesToHostname(ingress, hostname, config, hasExplicitListenPorts) {
				continue
			}

			prefix := strings.ToLower(config.Protocol)
			listeners = append(listeners, Listener{
				Name:     listenerName(prefix, indexes[prefix]),
				Port:     config.Port,
				Protocol: config.Protocol,
				Hostname: hostname,
			})
			indexes[prefix]++
		}
	}

	return listeners
}

func listenerName(prefix string, index int) string {
	if index == 0 {
		return prefix
	}

	return prefix + "-" + strconv.Itoa(index+1)
}

func collectHostnames(ingress networkingv1.Ingress) []string {
	hostnames := make([]string, 0, len(ingress.Spec.Rules))
	seen := make(map[string]struct{}, len(ingress.Spec.Rules))

	for _, rule := range ingress.Spec.Rules {
		if rule.Host == "" {
			continue
		}
		if _, ok := seen[rule.Host]; ok {
			continue
		}

		hostnames = append(hostnames, rule.Host)
		seen[rule.Host] = struct{}{}
	}

	return hostnames
}

func collectScheme(ingress networkingv1.Ingress) string {
	scheme := strings.TrimSpace(ingress.Annotations[ALBSchemeAnnotation])
	switch strings.ToLower(scheme) {
	case "internet-facing":
		return "internet-facing"
	case "internal":
		return "internal"
	default:
		return ""
	}
}

func collectListenerConfigs(ingress networkingv1.Ingress) []listenerConfig {
	configs, ok := parseListenPortsAnnotation(ingress.Annotations[ALBListenPortsAnnotation])
	if ok && len(configs) > 0 {
		return configs
	}

	return defaultListenerConfigs(ingress)
}

func parseListenPortsAnnotation(value string) ([]listenerConfig, bool) {
	if value == "" {
		return nil, false
	}

	var raw []map[string]int32
	if err := json.Unmarshal([]byte(value), &raw); err != nil {
		return nil, false
	}

	configs := make([]listenerConfig, 0, len(raw))
	for _, entry := range raw {
		for protocol, port := range entry {
			normalized := strings.ToUpper(protocol)
			if port <= 0 {
				continue
			}
			if normalized != "HTTP" && normalized != "HTTPS" {
				continue
			}

			configs = append(configs, listenerConfig{
				Protocol: normalized,
				Port:     port,
			})
		}
	}

	return configs, true
}

func collectTLSHosts(ingress networkingv1.Ingress) map[string]struct{} {
	tlsHosts := make(map[string]struct{})

	for _, tls := range ingress.Spec.TLS {
		for _, host := range tls.Hosts {
			if host == "" {
				continue
			}
			tlsHosts[host] = struct{}{}
		}
	}

	return tlsHosts
}

func defaultListenerConfigs(ingress networkingv1.Ingress) []listenerConfig {
	configs := []listenerConfig{
		{
			Protocol: "HTTP",
			Port:     80,
		},
	}

	if len(ingress.Spec.TLS) > 0 {
		configs = append(configs, listenerConfig{
			Protocol: "HTTPS",
			Port:     443,
		})
	}

	return configs
}

func listenerConfigAppliesToHostname(ingress networkingv1.Ingress, hostname string, config listenerConfig, hasExplicitListenPorts bool) bool {
	if config.Protocol != "HTTPS" {
		return true
	}

	if hasExplicitListenPorts {
		return true
	}

	tlsHosts := collectTLSHosts(ingress)
	return tlsAppliesToHostname(hostname, tlsHosts, len(ingress.Spec.TLS) > 0)
}

func tlsAppliesToHostname(hostname string, tlsHosts map[string]struct{}, hasTLSConfig bool) bool {
	if !hasTLSConfig {
		return false
	}

	if len(tlsHosts) == 0 {
		return true
	}

	_, ok := tlsHosts[hostname]
	return ok
}

func collectRouteRules(ingress networkingv1.Ingress) []HTTPRouteRule {
	rules := make([]HTTPRouteRule, 0)
	targetType := collectTargetType(ingress)

	for _, ingressRule := range ingress.Spec.Rules {
		if ingressRule.HTTP == nil {
			continue
		}

		for _, path := range ingressRule.HTTP.Paths {
			rule := HTTPRouteRule{
				Hostname:    ingressRule.Host,
				Path:        path.Path,
				PathType:    path.PathType,
				BackendRefs: collectBackendRefs(path.Backend, targetType),
			}
			rules = append(rules, rule)
		}
	}

	return rules
}

func collectTargetType(ingress networkingv1.Ingress) string {
	targetType := strings.TrimSpace(ingress.Annotations[ALBTargetTypeAnnotation])
	switch strings.ToLower(targetType) {
	case "instance":
		return "instance"
	case "ip":
		return "ip"
	default:
		return ""
	}
}

func collectBackendRefs(backend networkingv1.IngressBackend, targetType string) []BackendRef {
	if backend.Service == nil {
		return nil
	}

	return []BackendRef{
		{
			Name:       backend.Service.Name,
			PortNumber: backend.Service.Port.Number,
			PortName:   backend.Service.Port.Name,
			TargetType: targetType,
		},
	}
}
