package apicast

import (
	"context"
	"encoding/json"

	"github.com/3scale/apicast-operator/version"

	apicast "github.com/3scale/apicast-operator/pkg/apicast"
	appsv1alpha1 "github.com/3scale/apicast-operator/pkg/apis/apps/v1alpha1"
	"github.com/3scale/apicast-operator/pkg/reconcilers"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_apicast")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new APIcast Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	reconciler, err := newReconciler(mgr)
	if err != nil {
		return err
	}
	return add(mgr, reconciler)
}

// We create an Client Reader that directly queries the API server
// without going to the Cache provided by the Manager's Client because
// there are some resources that do not implement Watch (like ImageStreamTag)
// and the Manager's Client always tries to use the Cache when reading
func newAPIClientReader(mgr manager.Manager) (client.Client, error) {
	return client.New(mgr.GetConfig(), client.Options{Mapper: mgr.GetRESTMapper(), Scheme: mgr.GetScheme()})
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) (reconcile.Reconciler, error) {

	apiClientReader, err := newAPIClientReader(mgr)
	if err != nil {
		return nil, err
	}

	b := reconcilers.NewBaseReconciler(mgr.GetClient(), apiClientReader, mgr.GetScheme(), log)
	return &ReconcileAPIcast{
		BaseControllerReconciler: reconcilers.NewBaseControllerReconciler(b),
	}, nil
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("apicast-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource APIcast
	err = c.Watch(&source.Kind{Type: &appsv1alpha1.APIcast{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes in secondary resources
	err = c.Watch(&source.Kind{Type: &v1.Secret{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appsv1alpha1.APIcast{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appsv1alpha1.APIcast{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &v1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appsv1alpha1.APIcast{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &extensions.Ingress{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appsv1alpha1.APIcast{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileAPIcast implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileAPIcast{}

// ReconcileAPIcast reconciles a APIcast object
type ReconcileAPIcast struct {
	reconcilers.BaseControllerReconciler
}

const (
	APIcastOperatorVersionAnnotation = "apicast.apps.3scale.net/operator-version"
)

// Reconcile reads that state of the cluster for a APIcast object and makes changes based on the state read
// and what is in the APIcast.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileAPIcast) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling APIcast")

	instance, err := r.getAPIcast(request)
	if err != nil {
		if errors.IsNotFound(err) {
			r.Logger().Info("APIcast not found")
			return reconcile.Result{}, nil
		}
		r.Logger().Error(err, "Error getting APIcast")
		return reconcile.Result{}, err
	}

	if reqLogger.V(1).Enabled() {
		jsonData, err := json.MarshalIndent(instance, "", "  ")
		if err != nil {
			return reconcile.Result{}, err
		}
		reqLogger.V(1).Info(string(jsonData))
	}

	if instance.ObjectMeta.Annotations == nil || instance.ObjectMeta.Annotations[APIcastOperatorVersionAnnotation] == "" {
		r.Logger().Info("APIcast operator version not set in annotations. Setting it...")
		if instance.ObjectMeta.Annotations == nil {
			instance.ObjectMeta.Annotations = map[string]string{}
		}
		err = r.updateAPIcastOperatorVersionInAnnotations(instance)
		if err != nil {
			return reconcile.Result{}, err
		}
		r.Logger().Info("APIcast operator version in annotations set. Requeuing request...")
		return reconcile.Result{Requeue: true}, err
	}

	if instance.ObjectMeta.Annotations[APIcastOperatorVersionAnnotation] != version.Version {
		r.Logger().Info("APIcast operator version in annotations does not match expected version. Applying upgrade procedure...")
		upgradeReconcileResult, err := r.upgradeAPIcast()
		if err != nil {
			r.Logger().Error(err, "Error upgrading APIcast")
			return reconcile.Result{}, err
		}
		if upgradeReconcileResult.Requeue {
			return upgradeReconcileResult, nil
		}
		r.Logger().Info("APIcast upgrade procedure applied")
		r.Logger().Info("Setting APIcast operator version in annotations...")
		err = r.updateAPIcastOperatorVersionInAnnotations(instance)
		if err != nil {
			return reconcile.Result{}, err
		}
		r.Logger().Info("APIcast operator version in annotations set. Requeuing request...")
		return reconcile.Result{Requeue: true}, nil
	}

	logicReconciler := NewAPIcastLogicReconciler(r.BaseReconciler, instance)
	result, err := logicReconciler.Reconcile()
	if err != nil {
		// Ignore conflicts, resource might just be outdated.
		if errors.IsConflict(err) {
			reqLogger.Info("Resource update conflict error. Requeuing...", "error", err)
			return reconcile.Result{Requeue: true}, nil
		}

		r.Logger().Error(err, "Main reconciler")
		return result, err
	}
	if result.Requeue {
		r.Logger().Info("Requeuing request...")
		return result, nil
	}
	r.Logger().Info("APIcast logic reconciled")

	result, err = r.updateStatus(instance, &logicReconciler)
	if err != nil {
		// Ignore conflicts, resource might just be outdated.
		if errors.IsConflict(err) {
			reqLogger.Info("Resource update conflict error. Requeuing...", "error", err)
			return reconcile.Result{Requeue: true}, nil
		}

		r.Logger().Error(err, "Status reconciler")
		return result, err
	}
	if result.Requeue {
		r.Logger().Info("Requeuing request...")
		return result, nil
	}
	r.Logger().Info("APIcast status reconciled")
	return reconcile.Result{}, nil
}

func (r *ReconcileAPIcast) updateStatus(instance *appsv1alpha1.APIcast, reconciler *APIcastLogicReconciler) (reconcile.Result, error) {
	apicastFactory, err := apicast.Factory(instance, r.Client())
	if err != nil {
		return reconcile.Result{}, err
	}

	desiredDeployment := apicastFactory.Deployment()
	apicastDeployment := &appsv1.Deployment{}
	err = r.Client().Get(context.TODO(), types.NamespacedName{Name: desiredDeployment.Name, Namespace: desiredDeployment.Namespace}, apicastDeployment)
	if err != nil && !errors.IsNotFound(err) {
		return reconcile.Result{}, err
	}

	deployedImage := apicastDeployment.Spec.Template.Spec.Containers[0].Image
	if instance.Status.Image != deployedImage {
		instance.Status.Image = deployedImage
		err = r.Client().Status().Update(context.TODO(), instance)
		if err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileAPIcast) upgradeAPIcast() (reconcile.Result, error) {
	return reconcile.Result{}, nil
}

func (r *ReconcileAPIcast) getAPIcast(request reconcile.Request) (*appsv1alpha1.APIcast, error) {
	instance := appsv1alpha1.APIcast{}
	err := r.Client().Get(context.TODO(), request.NamespacedName, &instance)
	return &instance, err
}

func (r *ReconcileAPIcast) updateAPIcastOperatorVersionInAnnotations(instance *appsv1alpha1.APIcast) error {
	instance.Annotations[APIcastOperatorVersionAnnotation] = version.Version
	err := r.Client().Update(context.TODO(), instance)
	if err != nil {
		r.Logger().Error(err, "Error setting APIcast operator version in annotations")
	}
	return err
}
