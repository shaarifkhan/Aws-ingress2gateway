package albreader

import (
	networkingv1 "k8s.io/api/networking/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
)

const ALBIngressClass = "alb"

// GetIngressClass returns the ingress class from the modern field first,
// then falls back to the legacy annotation if needed.
func GetIngressClass(ingress networkingv1.Ingress) string {
	if ingress.Spec.IngressClassName != nil && *ingress.Spec.IngressClassName != "" {
		return *ingress.Spec.IngressClassName
	}

	return ingress.Annotations[networkingv1beta1.AnnotationIngressClass]
}

// IsALBIngress reports whether the Ingress targets the AWS ALB controller.
func IsALBIngress(ingress networkingv1.Ingress) bool {
	return GetIngressClass(ingress) == ALBIngressClass
}

// FilterALBIngresses keeps only Ingresses that target the AWS ALB controller.
func FilterALBIngresses(ingresses []networkingv1.Ingress) []networkingv1.Ingress {
	filtered := make([]networkingv1.Ingress, 0, len(ingresses))

	for _, ingress := range ingresses {
		if IsALBIngress(ingress) {
			filtered = append(filtered, ingress)
		}
	}

	return filtered
}
