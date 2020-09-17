package k8sutils

import (
	v1 "k8s.io/api/core/v1"
)

// CmpResources returns true if the resource requirements a is equal to b,
func CmpResources(a, b *v1.ResourceRequirements) bool {
	return CmpResourceList(&a.Limits, &b.Limits) && CmpResourceList(&a.Requests, &b.Requests)
}

// CmpResourceList returns true if the resourceList a is equal to b,
func CmpResourceList(a, b *v1.ResourceList) bool {
	return a.Cpu().Cmp(*b.Cpu()) == 0 &&
		a.Memory().Cmp(*b.Memory()) == 0 &&
		b.Pods().Cmp(*b.Pods()) == 0 &&
		b.StorageEphemeral().Cmp(*b.StorageEphemeral()) == 0
}
