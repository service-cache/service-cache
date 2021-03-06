package servicecache

import (
	"context"
	"strconv"
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

var log = logf.Log.WithName("controller_servicecache")

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
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileServiceCache) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	logger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	logger.Info("Reconciling ServiceCache")

	// Fetch the ServiceCache instance
	instance := &cachev1alpha1.ServiceCache{
		ObjectMeta: metav1.ObjectMeta {
			Name:      request.Name,
			Namespace: request.Namespace,
		},
	}
	err1 := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err1 != nil {
		if errors.IsNotFound(err1) {
			// remove the annotations on its related Service object.
			logger.Info("ServiceCache object is not found, so remove annotations from the related Service", "Service.Name", request.Name)
			r.removeAnnotationsFromService(instance.Namespace, instance.Name)
	
			// Request object not found, could have been deleted after reconcile request.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err1
	}

	if !validateServiceCache(instance) {
		logger.Info("The configuration in ServiceCache object is correct, so remove it")
		r.client.Delete(context.TODO(), instance)
		return reconcile.Result{}, nil
	}

	svc, err := r.findService(instance.Name, instance.Namespace)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("No related Service found, so delete the ServiceCache")
			// remove this servicecache object, since its corresponding service is not existent.
			r.client.Delete(context.TODO(), instance)
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	hasDiff := controller_utils.DiffServiceAndServiceCache(svc, instance)
	if !hasDiff {
		logger.Info("Configuration between Service and its ServiceCache has no difference")
		return reconcile.Result{}, nil
	}

	// read the configuration from service cache object, and update the annotations in service object
	r.syncServiceCacheToService(instance, svc)
	logger.Info("Configuration has been synced to Service from ServiceCache")

	// Set ServiceCache instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, svc, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Service has been labelled - don't requeue
	logger.Info("Service is now up to date")
	return reconcile.Result{}, nil
}

func (r *ReconcileServiceCache) syncServiceCacheToService(sc *cachev1alpha1.ServiceCache, svc *corev1.Service) error {
	svc.Annotations[controller_utils.KeyOfCacheableByDefault] = strconv.FormatBool(sc.Spec.CacheableByDefault)
	if sc.Spec.URLs != nil {
		var b strings.Builder
		b.WriteString("[")
		b.WriteString(strings.Join(sc.Spec.URLs, ","))
		b.WriteString("]")
		svc.Annotations[controller_utils.KeyOfCacheableUrls] = b.String()
	}
	r.client.Update(context.TODO(), svc)
  return nil
}

func (r *ReconcileServiceCache) removeAnnotationsFromService(svcName, svcNamespace string) error {
	svc, err := r.findService(svcName, svcNamespace)
	if err == nil {
		delete(svc.Annotations, controller_utils.KeyOfCacheableByDefault)
		delete(svc.Annotations, controller_utils.KeyOfCacheableUrls)
	}
	return err
}

func (r *ReconcileServiceCache) findService(svcName, svcNamespace string) (*corev1.Service, error) {
  svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta {
			Name:      svcName,
			Namespace: svcNamespace,
		},
	}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: svc.Name, Namespace: svc.Namespace}, svc)
	return svc, err
}

func validateServiceCache(sc *cachev1alpha1.ServiceCache) bool {
	//TODO: validate configurations in ServiceCache object
  return true
}