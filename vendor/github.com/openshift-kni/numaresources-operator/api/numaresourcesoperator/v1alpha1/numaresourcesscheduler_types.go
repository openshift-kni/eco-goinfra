/*
Copyright 2021.

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

// NUMAResourcesSchedulerSpec defines the desired state of NUMAResourcesScheduler
type NUMAResourcesSchedulerSpec struct {
	SchedulerImage string `json:"imageSpec"`
}

// NUMAResourcesSchedulerStatus defines the observed state of NUMAResourcesScheduler
type NUMAResourcesSchedulerStatus struct {
	Deployment NamespacedName `json:"deployment,omitempty"`
	// Conditions show the current state of the NUMAResourcesOperator Operator
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+genclient
//+genclient:nonNamespaced
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=numaressched,path=numaresourcesschedulers,scope=Cluster

// NUMAResourcesScheduler is the Schema for the numaresourcesschedulers API
type NUMAResourcesScheduler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NUMAResourcesSchedulerSpec   `json:"spec,omitempty"`
	Status NUMAResourcesSchedulerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NUMAResourcesSchedulerList contains a list of NUMAResourcesScheduler
type NUMAResourcesSchedulerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NUMAResourcesScheduler `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NUMAResourcesScheduler{}, &NUMAResourcesSchedulerList{})
}
