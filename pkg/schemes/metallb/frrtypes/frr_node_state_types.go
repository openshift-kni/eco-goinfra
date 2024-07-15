package frrtypes

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FRRNodeStateSpec defines the desired state of FRRNodeState.
type FRRNodeStateSpec struct {
}

// FRRNodeStateStatus defines the observed state of FRRNodeState.
type FRRNodeStateStatus struct {
	// RunningConfig represents the current FRR running config, which is the configuration the FRR instance is
	// currently running with.
	RunningConfig string `json:"runningConfig,omitempty"`
	// LastConversionResult is the status of the last translation between the `FRRConfiguration`s resources and FRR's
	// configuration, contains "success" or an error.
	LastConversionResult string `json:"lastConversionResult,omitempty"`
	// LastReloadResult represents the status of the last configuration update operation by FRR, contains
	//  "success" or an error.
	LastReloadResult string `json:"lastReloadResult,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// FRRNodeState exposes the status of the FRR instance running on each node.
type FRRNodeState struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FRRNodeStateSpec   `json:"spec,omitempty"`
	Status FRRNodeStateStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FRRNodeStateList contains a list of FRRNodeStatus.
type FRRNodeStateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FRRNodeState `json:"items"`
}
