package k8sutils

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type KubernetesObject interface {
	metav1.Object
	runtime.Object
}

func ObjectInfo(obj KubernetesObject) string {
	return fmt.Sprintf("%s/%s", obj.GetObjectKind().GroupVersionKind().Kind, obj.GetName())
}
