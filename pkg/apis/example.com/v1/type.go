package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// BarSpec defines the desired state of Bar
type BarSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS -- desired state of cluster
	DeploymentName string `json:"deploymentName"`
	Image          string `json:"image"`
	Replicas       *int32 `json:"replicas"`
}

// BarStatus defines the observed state of Bar.
// It should always be reconstructable from the state of the cluster and/or outside world.
type BarStatus struct {
	// INSERT ADDITIONAL STATUS FIELDS -- observed state of cluster
}

// 下面这个一定不能少，少了的话不能生成 lister 和 informer
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Bar is the Schema for the bars API
// +k8s:openapi-gen=true
type Bar struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BarSpec   `json:"spec,omitempty"`
	Status BarStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BarList contains a list of Bar
type BarList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Bar `json:"items"`
}
