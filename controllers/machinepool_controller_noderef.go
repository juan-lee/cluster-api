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
	"time"

	"github.com/pkg/errors"
	apicorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/klog"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha2"
	"sigs.k8s.io/cluster-api/controllers/noderefutil"
	"sigs.k8s.io/cluster-api/controllers/remote"
	capierrors "sigs.k8s.io/cluster-api/errors"
)

func (r *MachinePoolReconciler) reconcileNodeRefs(ctx context.Context, cluster *clusterv1.Cluster, machinePool *clusterv1.MachinePool) error {
	// Check that the MachinePool hasn't been deleted or in the process.
	if !machinePool.DeletionTimestamp.IsZero() {
		return nil
	}

	errs := []error{}
	for n := range machinePool.Spec.Instances {
		if err := r.reconcileNodeRef(ctx, cluster, machinePool, &machinePool.Spec.Instances[n], &machinePool.Status.Instances[n]); err != nil {
			errs = append(errs, err)
		}
	}
	return kerrors.NewAggregate(errs)
}

func (r *MachinePoolReconciler) reconcileNodeRef(ctx context.Context, cluster *clusterv1.Cluster, machinePool *clusterv1.MachinePool, spec *clusterv1.Instance, status *clusterv1.InstanceStatus) error {
	// Check that the MachinePool doesn't already have a NodeRef.
	if status.NodeRef != nil {
		return nil
	}

	// Check that Cluster isn't nil.
	if cluster == nil {
		klog.V(2).Infof("MachinePool %q in namespace %q doesn't have a linked cluster, won't assign NodeRef", machinePool.Name, machinePool.Namespace)
		return nil
	}

	// Check that the MachinePool has a valid ProviderID.
	if spec.ProviderID == nil || *spec.ProviderID == "" {
		klog.Warningf("MachinePool %q in namespace %q doesn't have a valid ProviderID yet", machinePool.Name, machinePool.Namespace)
		return nil
	}

	providerID, err := noderefutil.NewProviderID(*spec.ProviderID)
	if err != nil {
		return err
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
	nodeRef, err := r.getNodeReference(corev1Client, providerID)
	if err != nil {
		if err == ErrNodeNotFound {
			return errors.Wrapf(&capierrors.RequeueAfterError{RequeueAfter: 10 * time.Second},
				"cannot assign NodeRef to MachinePool %q in namespace %q, no matching Node", machinePool.Name, machinePool.Namespace)
		}
		klog.Errorf("Failed to assign NodeRef to MachinePool %q in namespace %q: %v", machinePool.Name, machinePool.Namespace, err)
		r.recorder.Event(machinePool, apicorev1.EventTypeWarning, "FailedSetNodeRef", err.Error())
		return err
	}

	// Set the MachinePool NodeRef.
	status.NodeRef = nodeRef
	klog.Infof("Set MachinePool's (%q in namespace %q) NodeRef to %q", machinePool.Name, machinePool.Namespace, status.NodeRef.Name)
	r.recorder.Event(machinePool, apicorev1.EventTypeNormal, "SuccessfulSetNodeRef", status.NodeRef.Name)
	return nil
}

func (r *MachinePoolReconciler) getNodeReference(client corev1.NodesGetter, providerID *noderefutil.ProviderID) (*apicorev1.ObjectReference, error) {
	listOpt := metav1.ListOptions{}

	for {
		nodeList, err := client.Nodes().List(listOpt)
		if err != nil {
			return nil, err
		}

		for _, node := range nodeList.Items {
			nodeProviderID, err := noderefutil.NewProviderID(node.Spec.ProviderID)
			if err != nil {
				klog.V(3).Infof("Failed to parse ProviderID for Node %q: %v", node.Name, err)
				continue
			}

			if providerID.Equals(nodeProviderID) {
				return &apicorev1.ObjectReference{
					Kind:       node.Kind,
					APIVersion: node.APIVersion,
					Name:       node.Name,
					UID:        node.UID,
				}, nil
			}
		}

		listOpt.Continue = nodeList.Continue
		if listOpt.Continue == "" {
			break
		}
	}

	return nil, ErrNodeNotFound
}
