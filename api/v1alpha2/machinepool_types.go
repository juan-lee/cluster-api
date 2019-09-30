/*
Copyright The Kubernetes Authors.

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

package v1alpha2

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	capierrors "sigs.k8s.io/cluster-api/errors"
)

const (
	// MachinePoolFinalizer is set on PrepareForCreate callback.
	MachinePoolFinalizer = "machinepool.cluster.x-k8s.io"
)

// ANCHOR: MachinePoolSpec

// MachinePoolSpec defines the desired state of MachinePool
type MachinePoolSpec struct {
	// ObjectMeta will autopopulate the Node created. Use this to
	// indicate what labels, annotations, name prefix, etc., should be used
	// when creating the Node.
	// +optional
	ObjectMeta `json:"metadata,omitempty"`

	// Bootstrap is a reference to a local struct which encapsulates
	// fields to configure the MachinePool’s bootstrapping mechanism.
	Bootstrap Bootstrap `json:"bootstrap"`

	// InfrastructureRef is a required reference to a custom resource
	// offered by an infrastructure provider.
	InfrastructureRef corev1.ObjectReference `json:"infrastructureRef"`

	// Version defines the desired Kubernetes version.
	// This field is meant to be optionally used by bootstrap providers.
	// +optional
	Version *string `json:"version,omitempty"`

	// ProviderID is the identification ID of the machine provided by the provider.
	// This field must match the provider ID as seen on the node object corresponding to this machine.
	// This field is required by higher level consumers of cluster-api. Example use case is cluster autoscaler
	// with cluster-api as provider. Clean-up logic in the autoscaler compares machines to nodes to find out
	// machines at provider which could not get registered as Kubernetes nodes. With cluster-api as a
	// generic out-of-tree provider for autoscaler, this field is required by autoscaler to be
	// able to have a provider view of the list of machines. Another list of nodes is queried from the k8s apiserver
	// and then a comparison is done to find out unregistered machines and are marked for delete.
	// This field will be set by the actuators and consumed by higher level entities like autoscaler that will
	// be interfacing with cluster-api as generic provider.
	// +optional
	ProviderID *string `json:"providerID,omitempty"`
}

// ANCHOR_END: MachinePoolSpec

// ANCHOR: MachinePoolStatus

// MachinePoolStatus defines the observed state of MachinePool
type MachinePoolStatus struct {
	// Replicas is the most recently observed number of replicas.
	Replicas int `json:"replicas"`

	// ErrorReason indicates that there is a problem reconciling the state, and
	// will be set to a token value suitable for programmatic interpretation.
	// +optional
	ErrorReason *capierrors.MachinePoolStatusError `json:"errorReason,omitempty"`

	// ErrorMessage indicates that there is a problem reconciling the state,
	// and will be set to a descriptive error message.
	// +optional
	ErrorMessage *string `json:"errorMessage,omitempty"`

	// Phase represents the current phase of cluster actuation.
	// E.g. Pending, Running, Terminating, Failed etc.
	// +optional
	Phase string `json:"phase,omitempty"`

	// InfrastructureReady is the state of the infrastructure provider.
	// +optional
	InfrastructureReady bool `json:"infrastructureReady"`

	// BootstrapReady is the state of the bootstrap provider.
	// +optional
	BootstrapReady bool `json:"bootstrapReady"`
}

// ANCHOR_END: MachinePoolStatus

// SetTypedPhase sets the Phase field to the string representation of MachinePoolPhase.
func (m *MachinePoolStatus) SetTypedPhase(p MachinePoolPhase) {
	m.Phase = string(p)
}

// GetTypedPhase attempts to parse the Phase field and return
// the typed MachinePoolPhase representation as described in `machinepool_phase_types.go`.
func (m *MachinePoolStatus) GetTypedPhase() MachinePoolPhase {
	switch phase := MachinePoolPhase(m.Phase); phase {
	case
		MachinePoolPhasePending,
		MachinePoolPhaseProvisioning,
		MachinePoolPhaseProvisioned,
		MachinePoolPhaseDeleting,
		MachinePoolPhaseFailed:
		return phase
	default:
		return MachinePoolPhaseUnknown
	}
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=machinepools,shortName=mp,scope=Namespaced,categories=cluster-api
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// MachinePool is the Schema for the machinepools API
type MachinePool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MachinePoolSpec   `json:"spec,omitempty"`
	Status MachinePoolStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MachinePoolList contains a list of MachinePool
type MachinePoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MachinePool `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MachinePool{}, &MachinePoolList{})
}
