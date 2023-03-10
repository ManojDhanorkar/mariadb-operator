package monitor

import (
	"context"

	mariadbv1alpha1 "github.com/persistentsys/mariadb-operator/pkg/apis/mariadb/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_monitor")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Monitor Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileMonitor{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("monitor-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Monitor
	err = c.Watch(&source.Kind{Type: &mariadbv1alpha1.Monitor{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Monitor
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &mariadbv1alpha1.Monitor{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileMonitor implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileMonitor{}

// ReconcileMonitor reconciles a Monitor object
type ReconcileMonitor struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Monitor object and makes changes based on the state read
// and what is in the Monitor.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileMonitor) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Monitor")

	// Fetch the Monitor instance
	instance := &mariadbv1alpha1.Monitor{}
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

	var result *reconcile.Result

	result, err = r.ensureDeployment(request, instance, r.monitorDeployment(instance))
	if result != nil {
		return *result, err
	}

	result, err = r.ensureService(request, instance, r.monitorService(instance))
	if result != nil {
		return *result, err
	}

	result, err = r.ensureServiceMonitor(request, instance, r.monitorServiceMonitor(instance))
	if result != nil {
		return *result, err
	}

	// TODO:  Create grafana dashboard
	/*result, err = r.ensureGrafanaDashboard(request, instance, r.monitorGrafanaDashboard(instance))
	if result != nil {
		return *result, err
	}*/

	err = r.updateMonitorStatus(instance)
	if err != nil {
		// Requeue the request if the status could not be updated
		return reconcile.Result{}, err
	}

	// Everything went fine, don't requeue
	return reconcile.Result{}, nil
}
