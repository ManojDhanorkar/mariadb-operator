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

// MariaDBMonitorSpec defines the desired state of MariaDBMonitor
type MariaDBMonitorSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of MariaDBMonitor. Edit mariadbmonitor_types.go to remove/update
	Size int32 `json:"size"`

	// Database source name
	DataSourceName string `json:"dataSourceName"`

	// Image name with version
	Image string `json:"image"`
}

// MariaDBMonitorStatus defines the observed state of MariaDBMonitor
type MariaDBMonitorStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	MonitorStatus string `json:"monitorStatus"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:path=mariadbmonitors,scope=Namespaced

// MariaDBMonitor is the Schema for the mariadbmonitors API
type MariaDBMonitor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MariaDBMonitorSpec   `json:"spec,omitempty"`
	Status MariaDBMonitorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MariaDBMonitorList contains a list of MariaDBMonitor
type MariaDBMonitorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MariaDBMonitor `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MariaDBMonitor{}, &MariaDBMonitorList{})
}
