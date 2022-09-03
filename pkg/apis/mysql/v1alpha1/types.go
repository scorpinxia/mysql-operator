package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MySQL is a simple user-defined resource.
type MySQL struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MySQLSpec   `json:"spec"`
	Status MySQLStatus `json:"status"`
}

// MySQLSpec is the spec fo Example.
type MySQLSpec struct {
	Version string `json:"version"`
}

// MySQLStatus is the status of Example.
type MySQLStatus struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MySQLList is the list of Example resources.
type MySQLList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []MySQL `json:"items"`
}
