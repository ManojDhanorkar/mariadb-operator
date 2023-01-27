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

	"aws-vm/go/pkg/mod/github.com/go-logr/logr@v1.2.0"
	"aws-vm/go/pkg/mod/k8s.io/apimachinery@v0.25.0/pkg/api/errors"
	"aws-vm/go/pkg/mod/k8s.io/apimachinery@v0.25.0/pkg/types"
	"aws-vm/go/pkg/mod/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"github.com/persistentsys/mariadb-operator/pkg/resource"

	"github.com/persistentsys/mariadb-operator/pkg/service"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	corev1 "k8s.io/api/core/v1"

	appsv1 "k8s.io/api/apps/v1"

	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"

	//"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	mariadbv1alpha1 "github.com/ManojDhanorkar/mariadb-operator/apis/mariadb/v1alpha1"
)

var ctx context.Context
var log = ctrllog.FromContext(ctx)

const mariadbPort = 80
const pvStorageName = "mariadb-pv-storage"
const pvClaimName = "mariadb-pv-claim"

func Labels(v *mariadbv1alpha1.MariaDB, tier string) map[string]string {
	return map[string]string{
		"app":        "MariaDB",
		"MariaDB_cr": v.Name,
		"tier":       tier,
	}
}

// func MariaDBBkpLabels(v *mariadbv1alpha1.MariaDB, tier string) map[string]string {
// 	return map[string]string{
// 		"app":        "MariaDB-Backup",
// 		"MariaDB_cr": v.Name,
// 		"tier":       tier,
// 	}
// }

// func MonitorLabels(v *mariadbv1alpha1.MariaDB, tier string) map[string]string {
// 	return map[string]string{
// 		"app":        "MariaDB-Monitor",
// 		"Monitor_cr": v.Name,
// 		"tier":       tier,
// 	}
// }

// func ServiceMonitorLabels(v *mariadbv1alpha1.MariaDB, tier string) map[string]string {
// 	return map[string]string{
// 		"app":        "ServiceMonitor",
// 		"Monitor_cr": v.Name,
// 		"tier":       tier,
// 	}
// }

func mysqlAuthName() string {
	return "mysql-auth"
}

// MariaDBReconciler reconciles a MariaDB object
type MariaDBReconciler struct {
	Client client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

//+kubebuilder:rbac:groups=mariadb.xyzcompany.com,resources=mariadbs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mariadb.xyzcompany.com,resources=mariadbs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mariadb.xyzcompany.com,resources=mariadbs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MariaDB object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *MariaDBReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// _ = log.FromContext(ctx)

	// TODO(user): your logic here

	// reqLogger.Info("Reconciling MariaDB")

	reqLogger := log.WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	//log := r.Log.WithValues("AWSVMScheduler", req.NamespacedName)
	reqLogger.Info("Reconciling MariaDB")

	log.Info("Reconciling MariaDB")

	// Fetch the MariaDB instance
	instance := &mariadbv1alpha1.MariaDB{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	var result *ctrl.Result

	result, err = r.ensureSecret(req, instance, r.mariadbAuthSecret(instance))
	if result != nil {
		return *result, err
	}

	result, err = r.ensureDeployment(req, instance, r.mariadbDeployment(instance))
	if result != nil {
		return *result, err
	}

	result, err = r.ensureService(req, instance, r.mariadbService(instance))
	if result != nil {
		return *result, err
	}

	result, err = r.ensurePV(req, instance)
	if result != nil {
		return *result, err
	}

	result, err = r.ensurePVC(req, instance)
	if result != nil {
		return *result, err
	}

	err = r.updateMariadbStatus(instance)
	if err != nil {
		// Requeue the request if the status could not be updated
		return ctrl.Result{}, err
	}

	// Everything went fine, don't requeue
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MariaDBReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mariadbv1alpha1.MariaDB{}).
		Complete(r)
}

// ensuredep
func mariadbDeploymentName(v *mariadbv1alpha1.MariaDB) string {
	return v.Name + "-server"
}

func (r *MariaDBReconciler) mariadbDeployment(v *mariadbv1alpha1.MariaDB) *appsv1.Deployment {
	labels := Labels(v, "mariadb")
	size := v.Spec.Size
	image := v.Spec.Image

	dbname := v.Spec.Database
	rootpwd := v.Spec.Rootpwd

	userSecret := &corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: mysqlAuthName()},
			Key:                  "username",
		},
	}

	passwordSecret := &corev1.EnvVarSource{
		SecretKeyRef: &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: mysqlAuthName()},
			Key:                  "password",
		},
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mariadbDeploymentName(v),
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
					Volumes: []corev1.Volume{
						{
							Name: pvStorageName,
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: pvClaimName,
								},
							},
						},
					},
					Containers: []corev1.Container{{
						Image:           image,
						ImagePullPolicy: corev1.PullAlways,
						Name:            "mariadb-service",
						Ports: []corev1.ContainerPort{{
							ContainerPort: mariadbPort,
							Name:          "mariadb",
						}},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      pvStorageName,
								MountPath: "/var/lib/mysql",
							},
						},
						Env: []corev1.EnvVar{
							{
								Name:  "MYSQL_ROOT_PASSWORD",
								Value: rootpwd,
							},
							{
								Name:  "MYSQL_DATABASE",
								Value: dbname,
							},
							{
								Name:      "MYSQL_USER",
								ValueFrom: userSecret,
							},
							{
								Name:      "MYSQL_PASSWORD",
								ValueFrom: passwordSecret,
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

func (r *MariaDBReconciler) ensureDeployment(req ctrl.Request,
	instance *mariadbv1alpha1.MariaDB,
	dep *appsv1.Deployment,
) (*ctrl.Result, error) {

	// See if deployment already exists and create if it doesn't
	found := &appsv1.Deployment{}
	err := r.Client.Get(ctx, types.NamespacedName{
		Name:      dep.Name,
		Namespace: instance.Namespace,
	}, found)
	if err != nil && errors.IsNotFound(err) {

		// Create the deployment
		log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Client.Create(ctx, dep)

		if err != nil {
			// Deployment failed
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return &reconcile.Result{}, err
		} else {
			// Deployment was successful
			return nil, nil
		}
	} else if err != nil {
		// Error that isn't due to the deployment not existing
		log.Error(err, "Failed to get Deployment")
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
		err = r.Client.Update(ctx, dep)
		if err != nil {
			log.Error(err, "Failed to update Deployment.", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return &ctrl.Result{}, err
		}
		log.Info("Updated Deployment image. ")
	}

	return nil, nil
}

// ensure service
func mariadbServiceName(v *mariadbv1alpha1.MariaDB) string {
	return v.Name + "-service"
}

func (r *MariaDBReconciler) mariadbService(v *mariadbv1alpha1.MariaDB) *corev1.Service {
	labels := Labels(v, "mariadb")

	s := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mariadbServiceName(v),
			Namespace: v.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{{
				Protocol:   corev1.ProtocolTCP,
				Port:       mariadbPort,
				TargetPort: intstr.FromInt(3306),
				NodePort:   v.Spec.Port,
			}},
			Type: corev1.ServiceTypeNodePort,
		},
	}

	ctrl.SetControllerReference(v, s, r.Scheme)
	return s
}

func (r *MariaDBReconciler) ensureService(req ctrl.Request,
	instance *mariadbv1alpha1.MariaDB,
	s *corev1.Service,
) (*ctrl.Result, error) {
	found := &corev1.Service{}
	err := r.Client.Get(ctx, types.NamespacedName{
		Name:      s.Name,
		Namespace: instance.Namespace,
	}, found)
	if err != nil && errors.IsNotFound(err) {

		// Create the service
		log.Info("Creating a new Service", "Service.Namespace", s.Namespace, "Service.Name", s.Name)
		err = r.Client.Create(ctx, s)

		if err != nil {
			// Creation failed
			log.Error(err, "Failed to create new Service", "Service.Namespace", s.Namespace, "Service.Name", s.Name)
			return &ctrl.Result{}, err
		} else {
			// Creation was successful
			return nil, nil
		}
	} else if err != nil {
		// Error that isn't due to the service not existing
		log.Error(err, "Failed to get Service")
		return &ctrl.Result{}, err
	}

	return nil, nil
}

func (r *MariaDBReconciler) ensureSecret(req ctrl.Request,
	instance *mariadbv1alpha1.MariaDB,
	s *corev1.Secret,
) (*ctrl.Result, error) {
	found := &corev1.Secret{}
	err := r.Client.Get(ctx, types.NamespacedName{
		Name:      s.Name,
		Namespace: instance.Namespace,
	}, found)
	if err != nil && errors.IsNotFound(err) {
		// Create the secret
		log.Info("Creating a new secret", "Secret.Namespace", s.Namespace, "Secret.Name", s.Name)
		err = r.Client.Create(ctx, s)

		if err != nil {
			// Creation failed
			log.Error(err, "Failed to create new Secret", "Secret.Namespace", s.Namespace, "Secret.Name", s.Name)
			return &ctrl.Result{}, err
		} else {
			// Creation was successful
			return nil, nil
		}
	} else if err != nil {
		// Error that isn't due to the secret not existing
		log.Error(err, "Failed to get Secret")
		return &ctrl.Result{}, err
	}

	return nil, nil
}

// ensurePV - Ensure that PV is present. If not, create one
func (r *MariaDBReconciler) ensurePV(req ctrl.Request,
	instance *mariadbv1alpha1.MariaDB,
) (*ctrl.Result, error) {
	pvName := resource.GetMariadbVolumeName(instance)
	_, err := service.FetchPVByName(pvName, r.Client)

	if err != nil && errors.IsNotFound(err) {
		// Create Persistent Volume
		log.Info("Creating a new PV", "PV.Name", pvName)

		pv := resource.NewMariaDbPV(instance, r.Scheme)
		err := r.Client.Create(context.TODO(), pv)
		if err != nil {
			// Creation failed
			log.Error(err, "Failed to create new PV", "PV.Name", pvName)
			return &ctrl.Result{}, err
		} else {
			// Creation was successful
			return nil, nil
		}
	} else if err != nil {
		// Error that isn't due to the service not existing
		log.Error(err, "Failed to get PV")
		return &ctrl.Result{}, err
	}
	return nil, nil
}

// ensurePVC - Ensure that PVC is present. If not, create one
func (r *MariaDBReconciler) ensurePVC(req ctrl.Request,
	instance *mariadbv1alpha1.MariaDB,
) (*ctrl.Result, error) {
	pvcName := resource.GetMariadbVolumeClaimName(instance)
	_, err := service.FetchPVCByNameAndNS(pvcName, instance.Namespace, r.Client)

	if err != nil && errors.IsNotFound(err) {
		// Create Persistent Volume Claim
		log.Info("Creating a new PVC", "PVC.Name", pvcName)

		pvc := resource.NewMariaDbPVC(instance, r.Scheme)
		err := r.Client.Create(context.TODO(), pvc)
		if err != nil {
			// Creation failed
			log.Error(err, "Failed to create new PVC", "PV.Name", pvcName, "PVC.Namespace", instance.Namespace)
			return &ctrl.Result{}, err
		} else {
			// Creation was successful
			return nil, nil
		}
	} else if err != nil {
		// Error that isn't due to the service not existing
		log.Error(err, "Failed to get PVC")
		return &ctrl.Result{}, err
	}
	return nil, nil
}

// updatestatus
func (r *MariaDBReconciler) updateMariadbStatus(v *mariadbv1alpha1.MariaDB) error {
	//v.Status.BackendImage = mariadbImage
	err := r.Client.Status().Update(ctx, v)
	return err
}

// auth secret

func (r *MariaDBReconciler) mariadbAuthSecret(v *mariadbv1alpha1.MariaDB) *corev1.Secret {

	username := v.Spec.Username
	password := v.Spec.Password

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mysqlAuthName(),
			Namespace: v.Namespace,
		},
		Type: "Opaque",
		Data: map[string][]byte{
			"username": []byte(username),
			"password": []byte(password),
		},
	}
	ctrl.SetControllerReference(v, secret, r.Scheme)
	return secret
}
