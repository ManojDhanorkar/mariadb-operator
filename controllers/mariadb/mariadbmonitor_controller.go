/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package mariadb

import (
	"context"

	grafanav1alpha1 "github.com/grafana-operator/grafana-operator/v4/api/integreatly/v1alpha1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	mariadbv1alpha1 "github.com/ManojDhanorkar/mariadb-operator/apis/mariadb/v1alpha1"
)

const DashboardJSON = `{
	"annotations":{
	   "list":[
		  {
			 "builtIn":1,
			 "datasource":"-- Grafana --",
			 "enable":true,
			 "hide":true,
			 "iconColor":"rgba(0, 211, 255, 1)",
			 "name":"Annotations & Alerts",
			 "type":"dashboard"
		  }
	   ]
	},
	"editable":true,
	"gnetId":null,
	"graphTooltip":0,
	"id":25,
	"links":[
 
	],
	"panels":[
	   {
		  "aliasColors":{
 
		  },
		  "bars":false,
		  "dashLength":10,
		  "dashes":false,
		  "datasource":"prometheus",
		  "fill":1,
		  "fillGradient":0,
		  "gridPos":{
			 "h":9,
			 "w":12,
			 "x":0,
			 "y":0
		  },
		  "hiddenSeries":false,
		  "id":2,
		  "legend":{
			 "avg":false,
			 "current":false,
			 "max":false,
			 "min":false,
			 "show":true,
			 "total":false,
			 "values":false
		  },
		  "lines":true,
		  "linewidth":1,
		  "nullPointMode":"null",
		  "options":{
			 "dataLinks":[
 
			 ]
		  },
		  "percentage":false,
		  "pointradius":2,
		  "points":false,
		  "renderer":"flot",
		  "seriesOverrides":[
 
		  ],
		  "spaceLength":10,
		  "stack":false,
		  "steppedLine":false,
		  "targets":[
			 {
				"expr":"mysql_global_status_threads_connected",
				"legendFormat":"",
				"refId":"A"
			 }
		  ],
		  "thresholds":[
 
		  ],
		  "timeFrom":null,
		  "timeRegions":[
 
		  ],
		  "timeShift":null,
		  "title":"Panel Title",
		  "tooltip":{
			 "shared":true,
			 "sort":0,
			 "value_type":"individual"
		  },
		  "type":"graph",
		  "xaxis":{
			 "buckets":null,
			 "mode":"time",
			 "name":null,
			 "show":true,
			 "values":[
 
			 ]
		  },
		  "yaxes":[
			 {
				"format":"short",
				"label":null,
				"logBase":1,
				"max":null,
				"min":null,
				"show":true
			 },
			 {
				"format":"short",
				"label":null,
				"logBase":1,
				"max":null,
				"min":null,
				"show":true
			 }
		  ],
		  "yaxis":{
			 "align":false,
			 "alignLevel":null
		  }
	   }
	],
	"schemaVersion":22,
	"style":"dark",
	"tags":[
 
	],
	"templating":{
	   "list":[
 
	   ]
	},
	"time":{
	   "from":"now-6h",
	   "to":"now"
	},
	"timepicker":{
	   "refresh_intervals":[
		  "5s",
		  "10s",
		  "30s",
		  "1m",
		  "5m",
		  "15m",
		  "30m",
		  "1h",
		  "2h",
		  "1d"
	   ]
	},
	"timezone":"",
	"title":"MariaDBDashboard",
	"uid":"IEVGlUgMk",
	"version":2
 }
 `

var logger2 = logf.Log.WithName("controller_monitor")

const monitorPort = 9104
const monitorPortName = "monitor"
const monitorApp = "monitor-app"

func (r *MariaDBMonitorReconciler) monitorGrafanaDashboard(v *mariadbv1alpha1.MariaDBMonitor) *grafanav1alpha1.GrafanaDashboard {

	labels := ServiceMonitorLabels(v, monitorApp)

	s := &grafanav1alpha1.GrafanaDashboard{
		ObjectMeta: v12.ObjectMeta{
			Name:      "GrafanaDashboard",
			Namespace: v.Namespace,
			Labels:    labels,
		},
		Spec: grafanav1alpha1.GrafanaDashboardSpec{
			Json: DashboardJSON,
			// Name: "mariadb.json",
			Plugins: []grafanav1alpha1.GrafanaPlugin{
				{
					Name:    "grafana-piechart-panel",
					Version: "1.5.0",
				},
			},
		},
	}

	ctrl.SetControllerReference(v, s, r.Scheme)
	return s
}

func (r *MariaDBMonitorReconciler) ensureGrafanaDashboard(req ctrl.Request,
	instance *mariadbv1alpha1.MariaDBMonitor,
	s *grafanav1alpha1.GrafanaDashboard,
) (*ctrl.Result, error) {

	found := &grafanav1alpha1.GrafanaDashboard{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{
		Name:      s.Name,
		Namespace: instance.Namespace,
	}, found)
	if err != nil && errors.IsNotFound(err) {

		// Create the GrafanaDashboard
		logger2.Info("Creating a new GrafanaDashboard", "GrafanaDashboard.Namespace", s.Namespace, "GrafanaDashboard.Name", s.Name)
		err = r.Client.Create(context.TODO(), s)

		if err != nil {
			// Creation failed
			logger2.Error(err, "Failed to create new GrafanaDashboard", "GrafanaDashboard.Namespace", s.Namespace, "GrafanaDashboard.Name", s.Name)
			return &ctrl.Result{}, err
		} else {
			// Creation was successful
			return nil, nil
		}
	} else if err != nil {
		// Error that isn't due to the service not existing
		logger2.Error(err, "Failed to get GrafanaDashboard")
		return &ctrl.Result{}, err
	}

	return nil, nil
}

func monitorServiceMonitorName(v *mariadbv1alpha1.MariaDBMonitor) string {
	return v.Name + "-serviceMonitor"
}

func (r *MariaDBMonitorReconciler) monitorServiceMonitor(v *mariadbv1alpha1.MariaDBMonitor) *monitoringv1.ServiceMonitor {
	labels := ServiceMonitorLabels(v, monitorApp)

	s := &monitoringv1.ServiceMonitor{

		ObjectMeta: v12.ObjectMeta{
			Name:      monitorServiceMonitorName(v),
			Namespace: v.Namespace,
			Labels:    labels,
		},
		Spec: monitoringv1.ServiceMonitorSpec{

			Endpoints: []monitoringv1.Endpoint{{
				Path: "/metrics",
				Port: monitorPortName,
			}},
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"tier": monitorApp,
				},
			},
		},
	}

	ctrl.SetControllerReference(v, s, r.Scheme)
	return s
}

func (r *MariaDBMonitorReconciler) ensureServiceMonitor(req ctrl.Request,
	instance *mariadbv1alpha1.MariaDBMonitor,
	s *monitoringv1.ServiceMonitor,
) (*ctrl.Result, error) {
	found := &monitoringv1.ServiceMonitor{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{
		Name:      s.Name,
		Namespace: instance.Namespace,
	}, found)
	if err != nil && errors.IsNotFound(err) {

		// Create the service
		logger2.Info("Creating a new ServiceMonitor", "ServiceMonitor.Namespace", s.Namespace, "ServiceMonitor.Name", s.Name)
		err = r.Client.Create(context.TODO(), s)

		if err != nil {
			// Creation failed
			logger2.Error(err, "Failed to create new ServiceMonitor", "ServiceMonitor.Namespace", s.Namespace, "ServiceMonitor.Name", s.Name)
			return &ctrl.Result{}, err
		} else {
			// Creation was successful
			return nil, nil
		}
	} else if err != nil {
		// Error that isn't due to the service not existing
		logger2.Error(err, "Failed to get ServiceMonitor")
		return &ctrl.Result{}, err
	}

	return nil, nil
}

func (r *MariaDBMonitorReconciler) updateMonitorStatus(v *mariadbv1alpha1.MariaDBMonitor) error {
	err := r.Client.Status().Update(context.TODO(), v)
	return err
}

func monitorServiceName(v *mariadbv1alpha1.MariaDBMonitor) string {
	return v.Name + "-service"
}

func (r *MariaDBMonitorReconciler) monitorService(v *mariadbv1alpha1.MariaDBMonitor) *corev1.Service {
	labels := MonitorLabels(v, monitorApp)

	s := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      monitorServiceName(v),
			Namespace: v.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{{
				Protocol:   corev1.ProtocolTCP,
				Port:       monitorPort,
				TargetPort: intstr.FromInt(9104),
				Name:       monitorPortName,
			}},
			Type: corev1.ServiceTypeNodePort,
		},
	}

	ctrl.SetControllerReference(v, s, r.Scheme)
	return s
}

func (r *MariaDBMonitorReconciler) ensureService(req ctrl.Request,
	instance *mariadbv1alpha1.MariaDBMonitor,
	s *corev1.Service,
) (*ctrl.Result, error) {
	found := &corev1.Service{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{
		Name:      s.Name,
		Namespace: instance.Namespace,
	}, found)
	if err != nil && errors.IsNotFound(err) {

		// Create the service
		logger2.Info("Creating a new Service", "Service.Namespace", s.Namespace, "Service.Name", s.Name)
		err = r.Client.Create(context.TODO(), s)

		if err != nil {
			// Creation failed
			logger2.Error(err, "Failed to create new Service", "Service.Namespace", s.Namespace, "Service.Name", s.Name)
			return &ctrl.Result{}, err
		} else {
			// Creation was successful
			return nil, nil
		}
	} else if err != nil {
		// Error that isn't due to the service not existing
		logger2.Error(err, "Failed to get Service")
		return &ctrl.Result{}, err
	}

	return nil, nil
}

func monitorDeploymentName(v *mariadbv1alpha1.MariaDBMonitor) string {
	return v.Name + "-deployment"
}

func (r *MariaDBMonitorReconciler) monitorDeployment(v *mariadbv1alpha1.MariaDBMonitor) *appsv1.Deployment {

	labels := MonitorLabels(v, monitorApp)

	size := v.Spec.Size
	image := v.Spec.Image
	dataSourceName := v.Spec.DataSourceName

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      monitorDeploymentName(v),
			Namespace: v.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &size,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:           image,
						ImagePullPolicy: corev1.PullAlways,
						Name:            monitorApp,
						Ports: []corev1.ContainerPort{{
							ContainerPort: monitorPort,
							Name:          monitorPortName,
						}},
						Env: []corev1.EnvVar{
							{
								Name:  "DATA_SOURCE_NAME",
								Value: dataSourceName,
							},
						},
					}},
				},
			},
		},
	}

	ctrl.SetControllerReference(v, dep, r.Scheme)
	return dep
}

func (r *MariaDBMonitorReconciler) ensureDeployment(req ctrl.Request,
	instance *mariadbv1alpha1.MariaDBMonitor,
	dep *appsv1.Deployment,
) (*ctrl.Result, error) {

	// See if deployment already exists and create if it doesn't
	found := &appsv1.Deployment{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{
		Name:      dep.Name,
		Namespace: instance.Namespace,
	}, found)
	if err != nil && errors.IsNotFound(err) {

		// Create the deployment
		logger2.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Client.Create(context.TODO(), dep)

		if err != nil {
			// Deployment failed
			logger2.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return &ctrl.Result{}, err
		} else {
			// Deployment was successful
			return nil, nil
		}
	} else if err != nil {
		// Error that isn't due to the deployment not existing
		logger2.Error(err, "Failed to get Deployment")
		return &ctrl.Result{}, err
	}

	// Check for any updates for redeployment
	applyChange := false

	// Ensure the deployment size is same as the spec
	size := instance.Spec.Size
	if *dep.Spec.Replicas != size {
		dep.Spec.Replicas = &size
		applyChange = true
	}

	// Ensure image name is correct, update image if required
	image := instance.Spec.Image
	var currentImage string = ""

	if found.Spec.Template.Spec.Containers != nil {
		currentImage = found.Spec.Template.Spec.Containers[0].Image
	}

	if image != currentImage {
		dep.Spec.Template.Spec.Containers[0].Image = image
		applyChange = true
	}

	if applyChange {
		err = r.Client.Update(context.TODO(), dep)
		if err != nil {
			logger2.Error(err, "Failed to update Deployment.", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return &ctrl.Result{}, err
		}
		logger2.Info("Updated Deployment image. ")
	}

	return nil, nil
}

// MariaDBMonitorReconciler reconciles a MariaDBMonitor object
type MariaDBMonitorReconciler struct {
	Client client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=mariadb.xyzcompany.com,resources=mariadbmonitors,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mariadb.xyzcompany.com,resources=mariadbmonitors/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mariadb.xyzcompany.com,resources=mariadbmonitors/finalizers,verbs=update
//+kubebuilder:rbac:groups=*,resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=*,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=*,resources=services,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MariaDBMonitor object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *MariaDBMonitorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// _ = logger2.FromContext(ctx)

	logger2.Info("Reconciling Monitor")

	instance := &mariadbv1alpha1.MariaDBMonitor{}

	err := r.Client.Get(context.TODO(), req.NamespacedName, instance)

	if err != nil {

		if errors.IsNotFound(err) {

			// Request object not found, could have been deleted after reconcile request.

			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.

			// Return and don't requeue

			return ctrl.Result{}, err

		}

		return ctrl.Result{}, nil

	}

	var result *ctrl.Result

	result, err = r.ensureDeployment(req, instance, r.monitorDeployment(instance))

	if result != nil {

		return *result, err

	}

	result, err = r.ensureService(req, instance, r.monitorService(instance))

	if result != nil {

		return *result, err

	}
	err = r.updateMonitorStatus(instance)

	if err != nil {

		// Requeue the request if the status could not be updated

		return ctrl.Result{}, err

	}

	// Everything went fine, don't requeue

	return ctrl.Result{}, nil

}

// SetupWithManager sets up the controller with the Manager.
func (r *MariaDBMonitorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mariadbv1alpha1.MariaDBMonitor{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}

func MonitorLabels(v *mariadbv1alpha1.MariaDBMonitor, tier string) map[string]string {
	return map[string]string{
		"app":        "MariaDB-Monitor",
		"Monitor_cr": v.Name,
		"tier":       tier,
	}
}

func ServiceMonitorLabels(v *mariadbv1alpha1.MariaDBMonitor, tier string) map[string]string {
	return map[string]string{
		"app":        "ServiceMonitor",
		"Monitor_cr": v.Name,
		"tier":       tier,
	}
}
