package e2eutil

import (
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func WaitForSecret(t *testing.T, kubeClient kubernetes.Interface, namespace, name string, retryInterval, timeout time.Duration) error {
	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		_, secretErr := kubeClient.CoreV1().Secrets(namespace).Get(name, metav1.GetOptions{})
		if secretErr != nil {
			if apierrors.IsNotFound(secretErr) {
				t.Logf("Waiting for availability of secret '%s'\n", name)
				return false, nil
			}
			return false, secretErr
		}

		t.Logf("Secret [%s] available\n", name)
		return true, nil
	})
	if err != nil {
		return err
	}
	return nil
}
