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

// MariaDBSpec defines the desired state of MariaDB
type MariaDBSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of MariaDB. Edit mariadb_types.go to remove/update
	Size int32 `json:"size"`

	// Database additional user details (base64 encoded)
	Username string `json:"username"`

	// Database additional user password (base64 encoded)
	Password string `json:"password"`

	// New Database name
	Database string `json:"database"`

	// Root user password
	Rootpwd string `json:"rootpwd"`

	// Image name with version
	Image string `json:"image"`

	// Database storage Path
	DataStoragePath string `json:"dataStoragePath"`

	// Database storage Size (Ex. 1Gi, 100Mi)
	DataStorageSize string `json:"dataStorageSize"`

	// Port number exposed for Database service
	Port int32 `json:"port"`
}

// MariaDBStatus defines the observed state of MariaDB
type MariaDBStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Nodes []string `json:"nodes,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MariaDB is the Schema for the mariadbs API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=mariadbs,scope=Namespaced
//+kubebuilder:object:root=true

// MariaDB is the Schema for the mariadbs API
type MariaDB struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MariaDBSpec   `json:"spec,omitempty"`
	Status MariaDBStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// MariaDBList contains a list of MariaDB
type MariaDBList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MariaDB `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MariaDB{}, &MariaDBList{})
}
