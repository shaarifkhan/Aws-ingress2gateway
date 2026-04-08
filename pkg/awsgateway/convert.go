package awsgateway

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	awsgatewayv1beta1 "sigs.k8s.io/aws-load-balancer-controller/apis/gateway/v1beta1"

	"aws-ingress2gateway/pkg/albir"
)

// Resources holds the typed AWS Gateway customization objects produced from the IR model.
type Resources struct {
	LoadBalancerConfigurations []awsgatewayv1beta1.LoadBalancerConfiguration
	TargetGroupConfigurations  []awsgatewayv1beta1.TargetGroupConfiguration
}

// ConvertModel converts the current IR model into typed AWS Gateway customization CRDs.
func ConvertModel(model albir.Model) Resources {
	resources := Resources{
		LoadBalancerConfigurations: make([]awsgatewayv1beta1.LoadBalancerConfiguration, 0, len(model.LoadBalancerConfigurations)),
		TargetGroupConfigurations:  make([]awsgatewayv1beta1.TargetGroupConfiguration, 0, len(model.TargetGroupConfigurations)),
	}

	for _, config := range model.LoadBalancerConfigurations {
		resources.LoadBalancerConfigurations = append(resources.LoadBalancerConfigurations, ConvertLoadBalancerConfiguration(config))
	}

	for _, config := range model.TargetGroupConfigurations {
		resources.TargetGroupConfigurations = append(resources.TargetGroupConfigurations, ConvertTargetGroupConfiguration(config))
	}

	return resources
}

// ConvertLoadBalancerConfiguration converts one IR load balancer config into the typed AWS CRD.
func ConvertLoadBalancerConfiguration(config albir.LoadBalancerConfiguration) awsgatewayv1beta1.LoadBalancerConfiguration {
	spec := awsgatewayv1beta1.LoadBalancerConfigurationSpec{}

	if config.LoadBalancerName != "" {
		loadBalancerName := config.LoadBalancerName
		spec.LoadBalancerName = &loadBalancerName
	}

	if scheme := convertLoadBalancerScheme(config.Scheme); scheme != nil {
		spec.Scheme = scheme
	}

	if len(config.LoadBalancerAttributes) > 0 {
		spec.LoadBalancerAttributes = convertLoadBalancerAttributes(config.LoadBalancerAttributes)
	}

	if config.WAFv2ACLARN != "" {
		spec.WAFv2 = &awsgatewayv1beta1.WAFv2Configuration{
			ACL: config.WAFv2ACLARN,
		}
	}

	listeners := convertListenerConfigurations(config.ListenerConfigurations)
	if len(listeners) > 0 {
		spec.ListenerConfigurations = &listeners
	}

	return awsgatewayv1beta1.LoadBalancerConfiguration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: awsgatewayv1beta1.GroupVersion.String(),
			Kind:       "LoadBalancerConfiguration",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.Name,
			Namespace: config.Namespace,
		},
		Spec: spec,
	}
}

// ConvertTargetGroupConfiguration converts one IR target group config into the typed AWS CRD.
func ConvertTargetGroupConfiguration(config albir.TargetGroupConfiguration) awsgatewayv1beta1.TargetGroupConfiguration {
	props := awsgatewayv1beta1.TargetGroupProps{}

	if targetType := convertTargetType(config.TargetType); targetType != nil {
		props.TargetType = targetType
	}

	if config.HealthCheckPath != "" {
		healthCheckPath := config.HealthCheckPath
		props.HealthCheckConfig = &awsgatewayv1beta1.HealthCheckConfiguration{
			HealthCheckPath: &healthCheckPath,
		}
	}

	return awsgatewayv1beta1.TargetGroupConfiguration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: awsgatewayv1beta1.GroupVersion.String(),
			Kind:       "TargetGroupConfiguration",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.Name,
			Namespace: config.Namespace,
		},
		Spec: awsgatewayv1beta1.TargetGroupConfigurationSpec{
			TargetReference: awsgatewayv1beta1.Reference{
				Name: config.ServiceName,
			},
			DefaultConfiguration: props,
		},
	}
}

func convertLoadBalancerScheme(scheme string) *awsgatewayv1beta1.LoadBalancerScheme {
	switch scheme {
	case "internal":
		value := awsgatewayv1beta1.LoadBalancerSchemeInternal
		return &value
	case "internet-facing":
		value := awsgatewayv1beta1.LoadBalancerSchemeInternetFacing
		return &value
	default:
		return nil
	}
}

func convertListenerConfigurations(configs []albir.LoadBalancerListenerConfiguration) []awsgatewayv1beta1.ListenerConfiguration {
	typed := make([]awsgatewayv1beta1.ListenerConfiguration, 0, len(configs))

	for _, config := range configs {
		listener := awsgatewayv1beta1.ListenerConfiguration{
			ProtocolPort: awsgatewayv1beta1.ProtocolPort(fmt.Sprintf("%s:%d", config.Protocol, config.Port)),
		}

		if len(config.Certificates) > 0 {
			defaultCertificate := config.Certificates[0]
			listener.DefaultCertificate = &defaultCertificate

			if len(config.Certificates) > 1 {
				certificates := make([]*string, 0, len(config.Certificates)-1)
				for _, certificate := range config.Certificates[1:] {
					certificate := certificate
					certificates = append(certificates, &certificate)
				}
				listener.Certificates = certificates
			}
		}

		if config.SSLPolicy != "" {
			sslPolicy := config.SSLPolicy
			listener.SslPolicy = &sslPolicy
		}

		typed = append(typed, listener)
	}

	return typed
}

func convertLoadBalancerAttributes(attributes []albir.LoadBalancerAttribute) []awsgatewayv1beta1.LoadBalancerAttribute {
	typed := make([]awsgatewayv1beta1.LoadBalancerAttribute, 0, len(attributes))

	for _, attribute := range attributes {
		typed = append(typed, awsgatewayv1beta1.LoadBalancerAttribute{
			Key:   attribute.Key,
			Value: attribute.Value,
		})
	}

	return typed
}

func convertTargetType(targetType string) *awsgatewayv1beta1.TargetType {
	switch targetType {
	case "instance":
		value := awsgatewayv1beta1.TargetTypeInstance
		return &value
	case "ip":
		value := awsgatewayv1beta1.TargetTypeIP
		return &value
	default:
		return nil
	}
}
