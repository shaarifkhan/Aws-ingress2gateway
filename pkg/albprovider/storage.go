package albprovider

import (
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/types"
)

// Storage holds the Ingress objects that our AWS-specific pipeline cares about.
// For now, it stores only ALB Ingress resources.
type Storage struct {
	Ingresses map[types.NamespacedName]*networkingv1.Ingress
}

// NewStorage creates an empty storage value.
func NewStorage() *Storage {
	return &Storage{
		Ingresses: make(map[types.NamespacedName]*networkingv1.Ingress),
	}
}

// AddIngress stores one Ingress under its namespace/name key.
func (s *Storage) AddIngress(ingress networkingv1.Ingress) {
	key := types.NamespacedName{
		Namespace: ingress.Namespace,
		Name:      ingress.Name,
	}

	ingressCopy := ingress
	s.Ingresses[key] = &ingressCopy
}

// AddIngresses stores a slice of Ingress objects.
func (s *Storage) AddIngresses(ingresses []networkingv1.Ingress) {
	for _, ingress := range ingresses {
		s.AddIngress(ingress)
	}
}

// ListIngresses returns the stored Ingresses as values.
func (s *Storage) ListIngresses() []networkingv1.Ingress {
	ingresses := make([]networkingv1.Ingress, 0, len(s.Ingresses))

	for _, ingress := range s.Ingresses {
		ingresses = append(ingresses, *ingress)
	}

	return ingresses
}
