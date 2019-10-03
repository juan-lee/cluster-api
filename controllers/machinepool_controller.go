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
	"strings"
	"sync"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha2"
	clusterv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"
	"sigs.k8s.io/cluster-api/controllers/external"
	"sigs.k8s.io/cluster-api/controllers/remote"
	capierrors "sigs.k8s.io/cluster-api/errors"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/patch"
)

var (
	errNilProviderID = errors.New("providerid is nil")
)

// +kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;patch
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io;bootstrap.cluster.x-k8s.io,resources=*,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machinepools;machinepools/status,verbs=get;list;watch;create;update;patch;delete

// MachinePoolReconciler reconciles a MachinePool object
type MachinePoolReconciler struct {
	client.Client
	Log logr.Logger

	controller       controller.Controller
	recorder         record.EventRecorder
	externalWatchers sync.Map
}

func (r *MachinePoolReconciler) SetupWithManager(mgr ctrl.Manager, options controller.Options) error {
	c, err := ctrl.NewControllerManagedBy(mgr).
		For(&clusterv1alpha2.MachinePool{}).
		WithOptions(options).
		Build(r)

	r.controller = c
	r.recorder = mgr.GetEventRecorderFor("machinepool-controller")
	return err
}

func (r *MachinePoolReconciler) Reconcile(req ctrl.Request) (_ ctrl.Result, reterr error) {
	ctx := context.Background()
	_ = r.Log.WithValues("machinepool", req.NamespacedName)

	mp := &clusterv1.MachinePool{}
	if err := r.Client.Get(ctx, req.NamespacedName, mp); err != nil {
		if apierrors.IsNotFound(err) {
			// Object not found, return. Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	// Initialize the patch helper.
	patchHelper, err := patch.NewHelper(mp, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}

	defer func() {
		// Always reconcile the Status.Phase field.
		r.reconcilePhase(ctx, mp)

		// Always attempt to Patch the MachinePool object and status after each reconciliation.
		if err := patchHelper.Patch(ctx, mp); err != nil {
			if reterr == nil {
				reterr = err
			}
		}
	}()

	// Cluster might be nil as some providers might not require a cluster object
	// for machine pool management.
	cluster, err := util.GetClusterFromMetadata(ctx, r.Client, mp.ObjectMeta)
	if errors.Cause(err) == util.ErrNoCluster {
		klog.V(2).Infof("MachinePool %q in namespace %q doesn't specify %q label, assuming nil cluster",
			mp.Name, mp.Namespace, clusterv1.MachineClusterLabelName)
	} else if err != nil {
		return ctrl.Result{}, errors.Wrapf(err, "failed to get cluster %q for machine pool %q in namespace %q",
			mp.Labels[clusterv1.MachineClusterLabelName], mp.Name, mp.Namespace)
	}

	// Handle deletion reconciliation loop.
	if !mp.ObjectMeta.DeletionTimestamp.IsZero() {
		// TODO(jpang): implement delete
		// return r.reconcileDelete(ctx, cluster, mp)
		return ctrl.Result{}, nil
	}

	// Handle normal reconciliation loop.
	return r.reconcile(ctx, cluster, mp)
}

func (r *MachinePoolReconciler) reconcile(ctx context.Context, cluster *clusterv1.Cluster, mp *clusterv1.MachinePool) (ctrl.Result, error) {
	// If the MachinePool belongs to a cluster, add an owner reference.
	if cluster != nil && r.shouldAdopt(mp) {
		mp.OwnerReferences = util.EnsureOwnerRef(mp.OwnerReferences, metav1.OwnerReference{
			APIVersion: cluster.APIVersion,
			Kind:       cluster.Kind,
			Name:       cluster.Name,
			UID:        cluster.UID,
		})
	}

	// If the MachinePool doesn't have a finalizer, add one.
	if !util.Contains(mp.Finalizers, clusterv1.MachinePoolFinalizer) {
		// TODO(jpang): implement delete
		// mp.Finalizers = append(mp.ObjectMeta.Finalizers, clusterv1.MachinePoolFinalizer)
	}

	// Call the inner reconciliation methods.
	reconciliationErrors := []error{
		r.reconcileBootstrap(ctx, mp),
		r.reconcileInfrastructure(ctx, mp),
		r.reconcileNodeRefs(ctx, cluster, mp),
	}

	// Parse the errors, making sure we record if there is a RequeueAfterError.
	res := ctrl.Result{}
	errs := []error{}
	for _, err := range reconciliationErrors {
		if requeueErr, ok := errors.Cause(err).(capierrors.HasRequeueAfterError); ok {
			// Only record and log the first RequeueAfterError.
			if !res.Requeue {
				res.Requeue = true
				res.RequeueAfter = requeueErr.GetRequeueAfter()
				klog.Infof("Reconciliation for MachinePool %q in namespace %q asked to requeue: %v", mp.Name, mp.Namespace, err)
			}
			continue
		}

		errs = append(errs, err)
	}
	return res, kerrors.NewAggregate(errs)
}

func (r *MachinePoolReconciler) reconcileDelete(ctx context.Context, cluster *clusterv1.Cluster, mp *clusterv1.MachinePool) (ctrl.Result, error) {
	if mp.Spec.ProviderID == nil {
		return ctrl.Result{}, errNilProviderID
	}
	klog.Infof("Deleting nodes in %q for machine pool %q", *mp.Spec.ProviderID, mp.Name)
	if err := r.deleteNodes(ctx, cluster, *mp.Spec.ProviderID); err != nil && !apierrors.IsNotFound(err) {
		klog.Errorf("Error deleting node %q for machine pool %q: %v", *mp.Spec.ProviderID, mp.Name, err)
		return ctrl.Result{}, err
	}

	if ok, err := r.reconcileDeleteExternal(ctx, mp); !ok || err != nil {
		// Return early and don't remove the finalizer if we got an error or
		// the external reconciliation deletion isn't ready.
		return ctrl.Result{}, err
	}

	mp.ObjectMeta.Finalizers = util.Filter(mp.ObjectMeta.Finalizers, clusterv1.MachinePoolFinalizer)
	return ctrl.Result{}, nil
}

func (r *MachinePoolReconciler) deleteNodes(ctx context.Context, cluster *clusterv1.Cluster, name string) error {
	if cluster == nil {
		// Try to retrieve the Node from the local cluster, if no Cluster reference is found.
		var nodes corev1.NodeList
		if err := r.Client.List(ctx, &nodes); err != nil {
			return err
		}
		for n := range nodes.Items {
			if strings.HasPrefix(nodes.Items[n].Spec.ProviderID, name) {
				err := r.Client.Delete(ctx, &nodes.Items[n])
				if err != nil {
					klog.Errorf("Error deleting node %q while deleting MachinePool with ProviderIDs %q/*, won't retry: %v",
						nodes.Items[n].Name, name, err)
					return nil
				}
			}
		}
		return nil
	}

	// Otherwise, proceed to get the remote cluster client and get the Node.
	remoteClient, err := remote.NewClusterClient(r.Client, cluster)
	if err != nil {
		klog.Errorf("Error creating a remote client for cluster %q while deleting MachinePool %q, won't retry: %v",
			cluster.Name, name, err)
		return nil
	}

	corev1Remote, err := remoteClient.CoreV1()
	if err != nil {
		klog.Errorf("Error creating a remote client for cluster %q while deleting MachinePool %q, won't retry: %v",
			cluster.Name, name, err)
		return nil
	}

	nodes, err := corev1Remote.Nodes().List(metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Error listing nodes for cluster %q while deleting MachinePool %q, won't retry: %v",
			cluster.Name, name, err)
		return nil
	}

	for n := range nodes.Items {
		if strings.HasPrefix(nodes.Items[n].Spec.ProviderID, name) {
			err := corev1Remote.Nodes().Delete(nodes.Items[n].Name, &metav1.DeleteOptions{})
			if err != nil {
				klog.Errorf("Error deleting node %q for cluster %q while deleting MachinePool with ProviderIDs %q/*, won't retry: %v",
					nodes.Items[n].Name, cluster.Name, name, err)
				return nil
			}
		}
	}
	return nil
}

// reconcileDeleteExternal tries to delete external references, returning true if it cannot find any.
func (r *MachinePoolReconciler) reconcileDeleteExternal(ctx context.Context, m *clusterv1.MachinePool) (bool, error) {
	objects := []*unstructured.Unstructured{}
	references := []*corev1.ObjectReference{
		m.Spec.Bootstrap.ConfigRef,
		&m.Spec.InfrastructureRef,
	}

	// Loop over the references and try to retrieve it with the client.
	for _, ref := range references {
		if ref == nil {
			continue
		}

		obj, err := external.Get(r.Client, ref, m.Namespace)
		if err != nil && !apierrors.IsNotFound(err) {
			return false, errors.Wrapf(err, "failed to get %s %q for MachinePool %q in namespace %q",
				ref.GroupVersionKind(), ref.Name, m.Name, m.Namespace)
		}
		if obj != nil {
			objects = append(objects, obj)
		}
	}

	// Issue a delete request for any object that has been found.
	for _, obj := range objects {
		if err := r.Client.Delete(ctx, obj); err != nil && !apierrors.IsNotFound(err) {
			return false, errors.Wrapf(err,
				"failed to delete %v %q for MachinePool %q in namespace %q",
				obj.GroupVersionKind(), obj.GetName(), m.Name, m.Namespace)
		}
	}

	// Return true if there are no more external objects.
	return len(objects) == 0, nil
}

func (r *MachinePoolReconciler) shouldAdopt(ms *clusterv1.MachinePool) bool {
	return !util.HasOwner(ms.OwnerReferences, clusterv1.GroupVersion.String(), []string{"MachinePool", "Cluster"})
}
