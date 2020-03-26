package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type NodeInfo struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// +optional
	Spec NodeInfoSpec `json:"spec,omitempty"`
	// +optional
	Status NodeInfoStatus `json:"status,omitempty"`
}

type NodeInfoSpec struct {
}

type NodeInfoStatus struct {
	// +optional
	TotalSize uint64 `json:"totalSize,omitempty"`
	// +optional
	UsedSize uint64 `json:"usedSize,omitempty"`
	// +optional
	PreAllocated map[string]string `json:"preAllocated,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type NodeInfoList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []NodeInfo `json:"items"`
}
