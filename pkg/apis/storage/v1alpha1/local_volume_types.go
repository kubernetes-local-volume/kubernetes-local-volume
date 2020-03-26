package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type LocalVolume struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// +optional
	Spec LocalVolumeSpec `json:"spec,omitempty"`
	// +optional
	Status LocalVolumeStatus `json:"status,omitempty"`
}

type LocalVolumeSpec struct {
}

type LocalVolumeStatus struct {
	// +optional
	TotalSize uint64 `json:"totalSize,omitempty"`
	// +optional
	FreeSize uint64 `json:"freeSize,omitempty"`
	// +optional
	PreAllocated map[string]string `json:"preAllocated,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type LocalVolumeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []LocalVolume `json:"items"`
}
