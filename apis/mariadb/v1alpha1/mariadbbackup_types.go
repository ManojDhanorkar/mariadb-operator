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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MariaDBBackupSpec defines the desired state of MariaDBBackup
type MariaDBBackupSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of MariaDBBackup. Edit mariadbbackup_types.go to remove/update
	BackupPath string `json:"backupPath"`
	BackupSize string `json:"backupSize"`
	Schedule   string `json:"schedule,omitempty"`
}

// MariaDBBackupStatus defines the observed state of MariaDBBackup
type MariaDBBackupStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	BackupStatus string `json:"backupStatus"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:path=mariadbbackups,scope=Namespaced

// MariaDBBackup is the Schema for the mariadbbackups API
type MariaDBBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MariaDBBackupSpec   `json:"spec,omitempty"`
	Status MariaDBBackupStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MariaDBBackupList contains a list of MariaDBBackup
type MariaDBBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MariaDBBackup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MariaDBBackup{}, &MariaDBBackupList{})
}
