package reconcilers

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type BaseControllerReconciler struct {
	// client should be a split client that reads objects from
	// the cache and writes to the Kubernetes APIServer
	client client.Client
	// apiClientReader should be a client that directly reads objects
	// from the Kubernetes APIServer
	apiClientReader client.Reader
	//
	scheme *runtime.Scheme
}

func NewBaseControllerReconciler(client client.Client, apiClientReader client.Reader, scheme *runtime.Scheme) BaseControllerReconciler {
	return BaseControllerReconciler{
		client:          client,
		apiClientReader: apiClientReader,
		scheme:          scheme,
	}
}

// blank assignment to verify that BaseReconciler implements reconcile.Reconciler
var _ reconcile.Reconciler = &BaseControllerReconciler{}

func (r *BaseControllerReconciler) Reconcile(reconcile.Request) (reconcile.Result, error) {
	return reconcile.Result{}, nil
}

func (r *BaseControllerReconciler) Client() client.Client {
	return r.client
}

func (r *BaseControllerReconciler) APIClientReader() client.Reader {
	return r.apiClientReader
}

func (r *BaseControllerReconciler) Scheme() *runtime.Scheme {
	return r.scheme
}
