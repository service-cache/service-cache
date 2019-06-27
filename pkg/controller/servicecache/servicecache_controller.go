package servicecache

import (
	"context"

	cachev1alpha1 "service-cache-operator/pkg/apis/cache/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_servicecache")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new ServiceCache Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileServiceCache{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("servicecache-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ServiceCache
	err = c.Watch(&source.Kind{Type: &cachev1alpha1.ServiceCache{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Services and requeue the owner ServiceCache
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &cachev1alpha1.ServiceCache{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileServiceCache implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileServiceCache{}

// ReconcileServiceCache reconciles a ServiceCache object
type ReconcileServiceCache struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a ServiceCache object and makes changes based on the state read
// and what is in the ServiceCache.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileServiceCache) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling ServiceCache")

	// Fetch the ServiceCache instance
	instance := &cachev1alpha1.ServiceCache{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Find the corresponding Service object
	svc := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: svc.Name, Namespace: svc.Namespace}, svc)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("No related Service found", "Service.Namespace", instance.Namespace, "Service.Name", instance.Name)
			svc.Labels["service-cache.github.io/author"] = "auto-generated"
			r.client.Add() // TODO: add the service for this service cache object
		}
		return reconcile.Result{}, err
	}

	// Add a label for the Service
	svc.Labels["service-cache.github.io/author"] = "auto-generated"
	// TODO: read the configuration from service cache object, and update the labels in service object
	r.client.Update(context.TODO(), svc)

	// Set ServiceCache instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, svc, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Service has been labelled - don't requeue
	reqLogger.Info("Found the Service", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
	return reconcile.Result{}, nil
}
