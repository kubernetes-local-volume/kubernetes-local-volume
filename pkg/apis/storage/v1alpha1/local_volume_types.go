package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type NodeLocalVolumeStorage struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// +optional
	Spec NodeLocalVolumeStorageSpec `json:"spec,omitempty"`
	// +optional
	Status NodeLocalVolumeStorageStatus `json:"status,omitempty"`
}

type NodeLocalVolumeStorageSpec struct {
}

type NodeLocalVolumeStorageStatus struct {
	// +optional
	TotalSize int64 `json:"totalSize,omitempty"`
	// +optional
	FreeSize int64 `json:"freeSize,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type NodeLocalVolumeStorageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []NodeLocalVolumeStorage `json:"items"`
}
