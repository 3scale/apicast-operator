package reconcilers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/3scale/apicast-operator/pkg/k8sutils"
)

// MutateFn is a function which mutates the existing object into it's desired state.
type MutateFn func(existing, desired k8sutils.KubernetesObject) (bool, error)

func CreateOnlyMutator(existing, desired k8sutils.KubernetesObject) (bool, error) {
	return false, nil
}

type BaseReconciler struct {
	// client should be a split client that reads objects from
	// the cache and writes to the Kubernetes APIServer
	client client.Client
	// apiClientReader should be a client that directly reads objects
	// from the Kubernetes APIServer
	apiClientReader client.Reader
	scheme          *runtime.Scheme
	logger          logr.Logger
}

func NewBaseReconciler(client client.Client, apiClientReader client.Reader, scheme *runtime.Scheme, logger logr.Logger) BaseReconciler {
	return BaseReconciler{
		client:          client,
		apiClientReader: apiClientReader,
		scheme:          scheme,
		logger:          logger,
	}
}

func (b *BaseReconciler) Client() client.Client {
	return b.client
}

func (b *BaseReconciler) APIClientReader() client.Reader {
	return b.apiClientReader
}

func (b *BaseReconciler) Scheme() *runtime.Scheme {
	return b.scheme
}

func (b *BaseReconciler) Logger() logr.Logger {
	return b.logger
}

func (b *BaseReconciler) ReconcileResource(desired k8sutils.KubernetesObject, mutateFn MutateFn) error {
	existing, err := k8sutils.CopyKubernetesObject(desired)
	if err != nil {
		return err
	}

	key, err := client.ObjectKeyFromObject(existing)
	if err != nil {
		return err
	}

	if err = b.Client().Get(context.TODO(), key, existing); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		// Not found
		if !k8sutils.IsObjectTaggedToDelete(desired) {
			return b.createResource(desired)
		}

		// Marked for deletion and not found. Nothing to do.
		return nil
	}

	// item found successfully
	if k8sutils.IsObjectTaggedToDelete(desired) {
		return b.deleteResource(desired)
	}

	update, err := mutateFn(existing, desired)
	if err != nil {
		return err
	}

	if update {
		return b.updateResource(existing)
	}

	return nil
}

func (b *BaseReconciler) createResource(obj k8sutils.KubernetesObject) error {
	b.Logger().Info(fmt.Sprintf("Created object %s", k8sutils.ObjectInfo(obj)))
	return b.Client().Create(context.TODO(), obj)
}

func (b *BaseReconciler) updateResource(obj k8sutils.KubernetesObject) error {
	b.Logger().Info(fmt.Sprintf("Updated object %s", k8sutils.ObjectInfo(obj)))
	return b.Client().Update(context.TODO(), obj)
}

func (b *BaseReconciler) deleteResource(obj k8sutils.KubernetesObject) error {
	b.Logger().Info(fmt.Sprintf("Delete object %s", k8sutils.ObjectInfo(obj)))
	return b.Client().Delete(context.TODO(), obj)
}
