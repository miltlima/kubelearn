package utils

import corev1 "k8s.io/api/core/v1"

func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// FirstContainer returns the first container of a pod spec, or nil if it has none.
func FirstContainer(spec corev1.PodSpec) *corev1.Container {
	if len(spec.Containers) == 0 {
		return nil
	}
	return &spec.Containers[0]
}

// PtrEquals reports whether p is set and points to want.
func PtrEquals[T comparable](p *T, want T) bool {
	return p != nil && *p == want
}
