/*
Copyright 2019 The Kubernetes Authors.

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

package controllers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	apicorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/klog"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha2"
	"sigs.k8s.io/cluster-api/controllers/remote"
	capierrors "sigs.k8s.io/cluster-api/errors"
)

func (r *MachinePoolReconciler) reconcileNodeRefs(ctx context.Context, cluster *clusterv1.Cluster, machinepool *clusterv1.MachinePool) error {
	// Check that the MachinePool hasn't been deleted or in the process.
	if !machinepool.DeletionTimestamp.IsZero() {
		return nil
	}

	// Check that the MachinePool doesn't already have a NodeRef.
	// TODO(jpang): check for the correct number of available replicas
	// if machinepool.Status.NodeRefs != nil {
	// 	return nil
	// }

	// Check that Cluster isn't nil.
	if cluster == nil {
		klog.V(2).Infof("MachinePool %q in namespace %q doesn't have a linked cluster, won't assign NodeRef", machinepool.Name, machinepool.Namespace)
		return nil
	}

	// Check that the MachinePool has a valid ProviderID.
	if machinepool.Spec.ProviderID == nil || *machinepool.Spec.ProviderID == "" {
		klog.Warningf("MachinePool %q in namespace %q doesn't have a valid ProviderID yet", machinepool.Name, machinepool.Namespace)
		return nil
	}

	clusterClient, err := remote.NewClusterClient(r.Client, cluster)
	if err != nil {
		return err
	}

	corev1Client, err := clusterClient.CoreV1()
	if err != nil {
		return err
	}

	// Get the Node reference.
	nodeRefs, err := r.getNodeReferences(corev1Client, *machinepool.Spec.ProviderID)
	if err != nil {
		if err == ErrNodeNotFound {
			return errors.Wrapf(&capierrors.RequeueAfterError{RequeueAfter: 10 * time.Second},
				"cannot assign NodeRefs to MachinePool %q in namespace %q, no matching Nodes", machinepool.Name, machinepool.Namespace)
		}
		klog.Errorf("Failed to assign NodeRefs to MachinePool %q in namespace %q: %v", machinepool.Name, machinepool.Namespace, err)
		r.recorder.Event(machinepool, apicorev1.EventTypeWarning, "FailedSetNodeRef", err.Error())
		return err
	}

	if machinepool.Status.Replicas == 0 || len(nodeRefs) != machinepool.Status.Replicas {
		return errors.Wrapf(&capierrors.RequeueAfterError{RequeueAfter: 10 * time.Second},
			"Replicas != NodeRefs [%q != %q] for MachinePool %q in namespace %q", len(nodeRefs), machinepool.Status.Replicas, machinepool.Name, machinepool.Namespace)
	}

	machinepool.Status.NodeRefs = nodeRefs
	klog.Infof("Set MachinePool's (%q in namespace %q) NodeRefs to %q", machinepool.Name, machinepool.Namespace, machinepool.Status.NodeRefs)
	r.recorder.Event(machinepool, apicorev1.EventTypeNormal, "SuccessfulSetNodeRef", fmt.Sprintf("%+v", machinepool.Status.NodeRefs))
	return nil
}

// nolint
func (r *MachinePoolReconciler) getNodeReferences(client corev1.NodesGetter, providerID string) ([]apicorev1.ObjectReference, error) {
	listOpt := metav1.ListOptions{}
	var nodeRefs []apicorev1.ObjectReference

	for {
		nodeList, err := client.Nodes().List(listOpt)
		if err != nil {
			return nil, err
		}

		for _, node := range nodeList.Items {
			if strings.HasPrefix(node.Spec.ProviderID, providerID) {
				nodeRefs = append(nodeRefs, apicorev1.ObjectReference{
					Kind:       node.Kind,
					APIVersion: node.APIVersion,
					Name:       node.Name,
					UID:        node.UID,
				})
			}
		}

		listOpt.Continue = nodeList.Continue
		if listOpt.Continue == "" {
			break
		}
	}

	if nodeRefs == nil {
		return nil, ErrNodeNotFound
	}
	return nodeRefs, nil
}
