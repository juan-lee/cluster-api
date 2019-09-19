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

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	clusterv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"
)

// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machinepools,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machinepools/status,verbs=get;update;patch

// MachinePoolReconciler reconciles a MachinePool object
type MachinePoolReconciler struct {
	client.Client
	Log logr.Logger

	recorder record.EventRecorder
}

func (r *MachinePoolReconciler) SetupWithManager(mgr ctrl.Manager, options controller.Options) error {
	err := ctrl.NewControllerManagedBy(mgr).
		For(&clusterv1alpha2.MachinePool{}).
		WithOptions(options).
		Complete(r)

	r.recorder = mgr.GetEventRecorderFor("machinepool-controller")
	return err
}

func (r *MachinePoolReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("machinepool", req.NamespacedName)

	return ctrl.Result{}, nil
}
