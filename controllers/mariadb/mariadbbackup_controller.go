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
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	mariadbv1alpha1 "github.com/ManojDhanorkar/mariadb-operator/apis/mariadb/v1alpha1"
)

var logger1 = logf.Log.WithName("controller_backup")
var defaultBackupConfig = NewDefaultBackupConfig()

const pvStorageName1 = "mariadb-bkp-pv-storage"

// const bkpPVClaimName = "mariadb-bkp-pv-claim"

// NewBackupCronJob Returns the CronJob object for the Database Backup
func NewBackupCronJob(bkp *mariadbv1alpha1.MariaDBBackup, db *mariadbv1alpha1.MariaDB, Scheme *runtime.Scheme) *batchv1.CronJob {

	bkpPVClaimName := GetMariadbBkpVolumeClaimName(bkp)
	dbPort := db.Spec.Port

	hostname := mariadbBkpServiceName(bkp) + "." + bkp.Namespace
	// currentTime := time.Now()
	//formatedDate := currentTime.Format("2006-01-02_15:04:05")
	// filename := "/var/lib/mysql/backup/backup_" + formatedDate + ".sql"
	filename := "/var/lib/mysql/backup_`date +%F_%T`.sql"
	backupCommand := "echo 'Starting DB Backup'  &&  " +
		"mysqldump -P " + fmt.Sprint(dbPort) + " -h '" + hostname +
		"' --lock-tables --all-databases > " + filename +
		"&& echo 'Completed DB Backup'"

	cron := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      bkp.Name,
			Namespace: bkp.Namespace,
			Labels:    MariaDBLabels1(db, "mariadb"),
		},
		Spec: batchv1.CronJobSpec{
			Schedule: bkp.Spec.Schedule,
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							ServiceAccountName: "mariadb-operator",
							Volumes: []corev1.Volume{
								{
									Name: pvStorageName1,
									VolumeSource: corev1.VolumeSource{
										PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
											ClaimName: bkpPVClaimName,
										},
									},
								},
							},
							Containers: []corev1.Container{
								{
									Name:    bkp.Name,
									Image:   db.Spec.Image,
									Command: []string{"/bin/sh", "-c"},
									Args:    []string{backupCommand},
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      pvStorageName1,
											MountPath: "/var/lib/mysql",
										},
									},
									Env: []corev1.EnvVar{
										{
											Name:  "MYSQL_PWD",
											Value: db.Spec.Rootpwd,
										},
										{
											Name:  "USER",
											Value: "root",
										},
									},
								},
							},
							RestartPolicy: "OnFailure",
						},
					},
				},
			},
		},
	}
	ctrl.SetControllerReference(bkp, cron, Scheme)
	return cron
}

func FetchCronJob(name, namespace string, Client client.Client) (*batchv1.CronJob, error) {
	logger1.Info("Fetching CronJob ...")
	cronJob := &batchv1.CronJob{}
	err := Client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, cronJob)
	return cronJob, err
}

func (r *MariaDBBackupReconciler) createCronJob(bkp *mariadbv1alpha1.MariaDBBackup, db *mariadbv1alpha1.MariaDB) error {
	if _, err := FetchCronJob(bkp.Name, bkp.Namespace, r.Client); err != nil {
		if err := r.Client.Create(context.TODO(), NewBackupCronJob(bkp, db, r.Scheme)); err != nil {
			return err
		}
	}
	return nil
}

func NewDbBackupPVC(bkp *mariadbv1alpha1.MariaDBBackup, v *mariadbv1alpha1.MariaDB, Scheme *runtime.Scheme) *corev1.PersistentVolumeClaim {
	logger1.Info("Creating new PVC for Database Backup")
	labels := MariaDBBkpLabels(bkp, "mariadb-backup")
	storageClassName := "standard"
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetMariadbBkpVolumeClaimName(bkp),
			Namespace: v.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &storageClassName,
			AccessModes:      []v1.PersistentVolumeAccessMode{v1.ReadWriteMany},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceName(corev1.ResourceStorage): resource.MustParse(bkp.Spec.BackupSize),
				},
			},
			VolumeName: GetMariadbBkpVolumeName(bkp),
		},
	}

	logger1.Info("PVC created for Database Backup ")
	ctrl.SetControllerReference(bkp, pvc, Scheme)
	return pvc
}

func FetchPVCByNameAndNS1(name, namespace string, client client.Client) (*corev1.PersistentVolumeClaim, error) {
	// reqLogger := logger1.WithValues("PVC Name", name, "PVC Namespace", namespace)
	logger1.Info("Fetching Persistent Volume Claim")

	pvc := &corev1.PersistentVolumeClaim{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, pvc)
	return pvc, err
}

func GetMariadbBkpVolumeClaimName(bkp *mariadbv1alpha1.MariaDBBackup) string {
	return bkp.Name + "-pv-claim"
}

func (r *MariaDBBackupReconciler) createBackupPVC(bkp *mariadbv1alpha1.MariaDBBackup, db *mariadbv1alpha1.MariaDB) error {
	pvcName := GetMariadbBkpVolumeClaimName(bkp)
	pvc, err := FetchPVCByNameAndNS1(pvcName, bkp.Namespace, r.Client)
	if err != nil || pvc == nil {
		pvc := NewDbBackupPVC(bkp, db, r.Scheme)
		if err := r.Client.Create(context.TODO(), pvc); err != nil {
			return err
		}
	}
	r.bkpPVC = pvc
	return nil
}

func NewDbBackupPV(bkp *mariadbv1alpha1.MariaDBBackup, v *mariadbv1alpha1.MariaDB, Scheme *runtime.Scheme) *corev1.PersistentVolume {
	logger1.Info("Creating new PV for Database Backup")
	labels := MariaDBBkpLabels(bkp, "mariadb-backup")
	pv := &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetMariadbBkpVolumeName(bkp),
			Namespace: v.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PersistentVolumeSpec{
			StorageClassName: "standard",
			Capacity: corev1.ResourceList{
				corev1.ResourceName(corev1.ResourceStorage): resource.MustParse(bkp.Spec.BackupSize),
			},
			AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteMany},
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: bkp.Spec.BackupPath},
			},
		},
	}

	logger1.Info("PV created for Database Backup ")
	ctrl.SetControllerReference(bkp, pv, Scheme)
	return pv
}

func FetchPVByName1(name string, Client client.Client) (*corev1.PersistentVolume, error) {

	logger1.Info("Fetching Persistent Volume")

	pv := &corev1.PersistentVolume{}
	err := Client.Get(context.TODO(), types.NamespacedName{Name: name}, pv)
	return pv, err
}

func GetMariadbBkpVolumeName(bkp *mariadbv1alpha1.MariaDBBackup) string {
	return bkp.Name + "-" + bkp.Namespace + "-pv"
}

func (r *MariaDBBackupReconciler) createBackupPV(bkp *mariadbv1alpha1.MariaDBBackup, db *mariadbv1alpha1.MariaDB) error {
	pvName := GetMariadbBkpVolumeName(bkp)
	pv, err := FetchPVByName1(pvName, r.Client)
	if err != nil || pv == nil {
		pv := NewDbBackupPV(bkp, db, r.Scheme)
		if err := r.Client.Create(context.TODO(), pv); err != nil {
			return err
		}
	}
	r.bkpPV = pv
	return nil
}

const dbBakupServicePort = 3306
const dbBakupServiceTargetPort = 3306

func mariadbBkpServiceName(bkp *mariadbv1alpha1.MariaDBBackup) string {
	return bkp.Name + "-service"
}

// NewDbBackupService Create a new service object for Database Backup
func NewDbBackupService(bkp *mariadbv1alpha1.MariaDBBackup, v *mariadbv1alpha1.MariaDB, scheme *runtime.Scheme) *corev1.Service {
	labels := MariaDBLabels1(v, "mariadb-backup")
	selectorLabels := MariaDBLabels1(v, "mariadb")

	s := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mariadbBkpServiceName(bkp),
			Namespace: v.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: selectorLabels,
			Ports: []corev1.ServicePort{{
				Protocol:   corev1.ProtocolTCP,
				Port:       dbBakupServicePort,
				TargetPort: intstr.FromInt(dbBakupServiceTargetPort),
			}},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	ctrl.SetControllerReference(v, s, scheme)
	return s
}

func buildDatabaseBackupCriteria(bkp *mariadbv1alpha1.MariaDBBackup, db *mariadbv1alpha1.MariaDB) *client.ListOptions {
	labelSelector := labels.SelectorFromSet(MariaDBLabels1(db, "mariadb-backup"))
	listOps := &client.ListOptions{Namespace: bkp.Namespace, LabelSelector: labelSelector}
	return listOps
}

func FetchDatabaseBackupService(bkp *mariadbv1alpha1.MariaDBBackup, db *mariadbv1alpha1.MariaDB, Client client.Client) (*corev1.Service, error) {
	logger1.Info("Fetching Database Backup Service ...")
	listOps := buildDatabaseBackupCriteria(bkp, db)
	bkpServiceList := &corev1.ServiceList{}
	err := Client.List(context.TODO(), bkpServiceList, listOps)
	if err != nil {
		return nil, err
	}

	if len(bkpServiceList.Items) == 0 {
		return nil, err
	}

	srv := bkpServiceList.Items[0]
	return &srv, nil
}

func (r *MariaDBBackupReconciler) getDatabaseBackupService(bkp *mariadbv1alpha1.MariaDBBackup, db *mariadbv1alpha1.MariaDB) error {
	dbService, err := FetchDatabaseBackupService(bkp, db, r.Client)
	if err != nil || dbService == nil {
		if err := r.Client.Create(context.TODO(), NewDbBackupService(bkp, db, r.Scheme)); err != nil {
			return err
		}
	}
	//r.dbService = dbService
	return nil
}

func FetchDatabaseService(bkp *mariadbv1alpha1.MariaDBBackup, db *mariadbv1alpha1.MariaDB, Client client.Client) (*corev1.Service, error) {
	logger1.Info("Fetching Database Service ...")
	listOps := buildDatabaseCriteria(bkp, db)
	dbServiceList := &corev1.ServiceList{}
	err := Client.List(context.TODO(), dbServiceList, listOps)
	if err != nil {
		return nil, err
	}

	if len(dbServiceList.Items) == 0 {
		return nil, err
	}

	srv := dbServiceList.Items[0]
	return &srv, nil
}

func (r *MariaDBBackupReconciler) getDatabaseService(bkp *mariadbv1alpha1.MariaDBBackup, db *mariadbv1alpha1.MariaDB) error {
	dbService, err := FetchDatabaseService(bkp, db, r.Client)
	if err != nil || dbService == nil {
		r.dbService = nil
		err := fmt.Errorf("Unable to find the Database Service")
		return err
	}
	r.dbService = dbService
	return nil
}

func buildDatabaseCriteria(bkp *mariadbv1alpha1.MariaDBBackup, db *mariadbv1alpha1.MariaDB) *client.ListOptions {
	labelSelector := labels.SelectorFromSet(MariaDBLabels1(db, "mariadb"))
	listOps := &client.ListOptions{Namespace: db.Namespace, LabelSelector: labelSelector}
	return listOps
}

func FetchDatabasePod(bkp *mariadbv1alpha1.MariaDBBackup, db *mariadbv1alpha1.MariaDB, Client client.Client) (*corev1.Pod, error) {
	logger1.Info("Fetching Database Pod ...")
	listOps := buildDatabaseCriteria(bkp, db)
	dbPodList := &corev1.PodList{}
	err := Client.List(context.TODO(), dbPodList, listOps)
	if err != nil {
		return nil, err
	}

	if len(dbPodList.Items) == 0 {
		return nil, err
	}

	pod := dbPodList.Items[0]
	return &pod, nil
}

func (r *MariaDBBackupReconciler) getDatabasePod(bkp *mariadbv1alpha1.MariaDBBackup, db *mariadbv1alpha1.MariaDB) error {
	dbPod, err := FetchDatabasePod(bkp, db, r.Client)
	if err != nil || dbPod == nil {
		r.dbPod = nil
		err := fmt.Errorf("Unable to find the Database Pod")
		return err
	}
	r.dbPod = dbPod
	return nil
}

func FetchDatabaseCR(name, namespace string, client client.Client) (*mariadbv1alpha1.MariaDB, error) {
	logger1.Info("Fetching Database CR ...")
	db := &mariadbv1alpha1.MariaDB{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, db)
	return db, err
}

func (r *MariaDBBackupReconciler) createResources(bkp *mariadbv1alpha1.MariaDBBackup, req ctrl.Request) error {
	logger1.Info("Creating secondary Backup resources ...")

	// Check if the database instance was created
	db, err := FetchDatabaseCR("mariadb", req.Namespace, r.Client)
	if err != nil {
		logger1.Error(err, "Failed to fetch Database instance/cr")
		return err
	}

	// Get the Database Pod created by the Database Controller
	if err := r.getDatabasePod(bkp, db); err != nil {
		logger1.Error(err, "Failed to get a Database pod")
		return err
	}

	// Get the Database Service created by the Database Controller
	if err := r.getDatabaseService(bkp, db); err != nil {
		logger1.Error(err, "Failed to get a Database service")
		return err
	}

	// Get the Database Backup Service created by the Backup Controller
	if err := r.getDatabaseBackupService(bkp, db); err != nil {
		logger1.Error(err, "Failed to get a Database Backup service")
		return err
	}

	// Check if the PV is created, if not create one
	if err := r.createBackupPV(bkp, db); err != nil {
		logger1.Error(err, "Failed to create the Persistent Volume for MariaDB Backup")
		return err
	}

	// Check if the PVC is created, if not create one
	if err := r.createBackupPVC(bkp, db); err != nil {
		logger1.Error(err, "Failed to create the Persistent Volume Claim for MariaDB Backup")
		return err
	}

	// Check if the cronJob is created, if not create one
	if err := r.createCronJob(bkp, db); err != nil {
		logger1.Error(err, "Failed to create the CronJob")
		return err
	}

	return nil
}

const (
	schedule   = "0 0 * * *"
	backupPath = "/mnt/backup"
)

type DefaultBackupConfig struct {
	Schedule   string `json:"schedule"`
	BackupPath string `json:"backupPath"`
}

func NewDefaultBackupConfig() *DefaultBackupConfig {
	return &DefaultBackupConfig{
		Schedule:   schedule,
		BackupPath: backupPath,
	}
}

// AddBackupMandatorySpecs will add the specs which are mandatory for Backup CR in the case them
// not be applied
func AddBackupMandatorySpecs(bkp *mariadbv1alpha1.MariaDBBackup) {

	/*
	 Backup Container
	*/

	if bkp.Spec.Schedule == "" {
		bkp.Spec.Schedule = defaultBackupConfig.Schedule
	}

	if bkp.Spec.BackupPath == "" {
		bkp.Spec.BackupPath = defaultBackupConfig.BackupPath
	}

}

func FetchBackupCR(name, namespace string, Client client.Client) (*mariadbv1alpha1.MariaDBBackup, error) {
	logger1.Info("Fetching Backup CR ...")
	bkp := &mariadbv1alpha1.MariaDBBackup{}
	err := Client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, bkp)
	return bkp, err
}

// MariaDBBackupReconciler reconciles a MariaDBBackup object
type MariaDBBackupReconciler struct {
	Client    client.Client
	Scheme    *runtime.Scheme
	dbPod     *corev1.Pod
	dbService *corev1.Service
	bkpPV     *corev1.PersistentVolume
	bkpPVC    *corev1.PersistentVolumeClaim
}

//+kubebuilder:rbac:groups=mariadb.xyzcompany.com,resources=mariadbbackups,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mariadb.xyzcompany.com,resources=mariadbbackups/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mariadb.xyzcompany.com,resources=mariadbbackups/finalizers,verbs=update
//+kubebuilder:rbac:groups=batch,resources=cronjobs;jobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=*,resources=pods;jobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=*,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=*,resources=persistentvolumes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=*,resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MariaDBBackup object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *MariaDBBackupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// _ = logger1.FromContext(ctx)

	logger1.Info("Reconciling Backup")

	bkp, err := FetchBackupCR(req.Name, req.Namespace, r.Client)
	//instance := &mariadbv1alpha1.Backup{}
	//err := r.Client.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile req.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			logger1.Info("Backup resource not found. Ignoring since object must be deleted.")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the req.
		logger1.Error(err, "Failed to get Backup.")
		return ctrl.Result{}, err
	}

	// Add const values for mandatory specs
	logger1.Info("Adding backup mandatory specs")
	AddBackupMandatorySpecs(bkp)

	// Create mandatory objects for the Backup
	if err := r.createResources(bkp, req); err != nil {
		logger1.Error(err, "Failed to create and update the secondary resource required for the Backup CR")
		return ctrl.Result{}, err
	}

	logger1.Info("Stop Reconciling Backup ...")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MariaDBBackupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mariadbv1alpha1.MariaDBBackup{}).
		Owns(&batchv1.CronJob{}).
		Complete(r)
}

func MariaDBLabels1(v *mariadbv1alpha1.MariaDB, tier string) map[string]string {
	return map[string]string{
		"app":        "MariaDB",
		"MariaDB_cr": v.Name,
		"tier":       tier,
	}
}

func MariaDBBkpLabels(v *mariadbv1alpha1.MariaDBBackup, tier string) map[string]string {
	return map[string]string{
		"app":        "MariaDB-Backup",
		"MariaDB_cr": v.Name,
		"tier":       tier,
	}
}
