package service

import (
	"context"
	"strings"

	cachev1alpha1 "service-cache-operator/pkg/apis/cache/v1alpha1"
	controller_utils "service-cache-operator/pkg/controller/utils"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_service")

// Add creates a new Service Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileService{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("service-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Service
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Pods and requeue the owner Service
	err = c.Watch(&source.Kind{Type: &cachev1alpha1.ServiceCache{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &corev1.Service{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileService implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileService{}

// ReconcileService reconciles a Service object
type ReconcileService struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Service object and makes changes based on the state read
// and what is in the Service.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileService) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Service")

	// Fetch the Service instance
	instance := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta {
			Name:      request.Name,
			Namespace: request.Namespace,
		},
	}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	serviceCache, errOfServiceCache := r.findServiceCache(instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			reqLogger.Info("The Service is not found, perhaps it's deleted already.")
			if errOfServiceCache == nil {
				reqLogger.Info("The Service is not found, but found its related ServiceCache object so delete it.")
				r.client.Delete(context.TODO(), serviceCache)
			}
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reqLogger.Error(err, "The Service cannot be read")
		return reconcile.Result{}, err
	}

	// if service is not annotated, then skip; Furthermore, if the ServiceCache object for the service is found, remove it.
	if !isAnnotatedByServiceCache(instance) {
		if errOfServiceCache == nil && serviceCache != nil {
			reqLogger.Info("Service is not annotated but found its ServiceCache object, so remove this ServiceCache object",
			  "ServiceCache.Namespace", serviceCache.Namespace, "ServiceCache.Name", serviceCache.Name)
			r.client.Delete(context.TODO(), serviceCache)
		}
		// The service is not annotated by service cache annotations, so return and don't requeue
		reqLogger.Info("Skip reconcile: Service is not annotated so it's not our target")
		return reconcile.Result{}, nil
	}

	// get or create the Service Cache object
	if errOfServiceCache != nil {
		if errors.IsNotFound(errOfServiceCache) {
			reqLogger.Error(errOfServiceCache, "Failed to find the related ServiceCache object: so create one")
			serviceCache, errOfServiceCache = r.createServiceCache(instance)
			if errOfServiceCache != nil {
				return reconcile.Result{}, errOfServiceCache
			}
		} else {
			return reconcile.Result{}, errOfServiceCache
		}
	}

	hasDiff := controller_utils.DiffServiceAndServiceCache(instance, serviceCache)
	if !hasDiff {
		reqLogger.Info("Skip reconcile: Configuration between Service and its ServiceCache has no difference", "Service.Namespace", instance.Namespace, "Service.Name", instance.Name)
		return reconcile.Result{}, nil
	}

	// update service cache based on service's configuration
	r.syncServiceToServiceCache(instance, serviceCache)

	// Set Service instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, serviceCache, r.scheme); err != nil {
		reqLogger.Error(err, "Failed to call SetControllerReference()", "Service.Namespace", instance.Namespace, "Service.Name", instance.Name)
		return reconcile.Result{}, err
	}

	// Service Cache object is up to date now, so don't requeue
	reqLogger.Info("Skip reconcile: ServiceCache is now up to date", "ServiceCache.Namespace", serviceCache.Namespace, "ServiceCache.Name", serviceCache.Name)
	return reconcile.Result{}, nil
}

// findServiceCache returns a ServiceCache object or nil
func (r *ReconcileService) findServiceCache(svc *corev1.Service) (*cachev1alpha1.ServiceCache, error) {
	logger := log.WithValues("ServiceCache.Namespace", svc.Namespace, "ServiceCache.Name", svc.Name)
	found := &cachev1alpha1.ServiceCache{
		ObjectMeta: metav1.ObjectMeta {
			Name:      svc.Name,
			Namespace: svc.Namespace,
		},
	}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: svc.Name, Namespace: svc.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		logger.Error(err, "ServiceCache object is not found")
		return nil, err
	}
	return found, err
}

// createServiceCache returns a ServiceCache object
func (r *ReconcileService) createServiceCache(svc *corev1.Service) (*cachev1alpha1.ServiceCache, error) {
	logger := log.WithValues("ServiceCache.Namespace", svc.Namespace, "ServiceCache.Name", svc.Name)
	sc := &cachev1alpha1.ServiceCache {
		ObjectMeta: metav1.ObjectMeta {
			Name:      svc.Name,
			Namespace: svc.Namespace,
		},
		Spec: cachev1alpha1.ServiceCacheSpec {
			CacheableByDefault: false,
			URLs: nil,
		},
	}
	logger.Info("Creating a new ServiceCache object")
	err := r.client.Create(context.TODO(), sc)

	return sc, err
}

func (r *ReconcileService) syncServiceToServiceCache(svc *corev1.Service, serviceCache *cachev1alpha1.ServiceCache) error {
	serviceCache.Spec.CacheableByDefault = (svc.Annotations["service-cache.github.io/default"] == "true")
	urls := svc.Annotations["service-cache.github.io/URLs"]
	urls = strings.TrimSuffix(strings.TrimPrefix(urls, "["), "]")
	serviceCache.Spec.URLs = strings.Split(urls, ",")
	logger := log.WithValues("ServiceCache.Namespace", svc.Namespace, "ServiceCache.Name", svc.Name)
	logger.Info("Syncing to servicecache", "CacheableByDefault", serviceCache.Spec.CacheableByDefault, "URLs", serviceCache.Spec.URLs)
	r.client.Update(context.TODO(), serviceCache)
	return nil
}

func isAnnotatedByServiceCache(svc *corev1.Service) bool {
	for k := range svc.Annotations {
		if strings.HasPrefix(k, "service-cache.github.io/") {
			return true
		}
	}
	return false
}
