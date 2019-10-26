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

package v1alpha2

import (
	"errors"
	"fmt"

	apiconversion "k8s.io/apimachinery/pkg/conversion"
	"sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *Cluster) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha3.Cluster)

	return Convert_v1alpha2_Cluster_To_v1alpha3_Cluster(src, dst, nil)
}

// nolint
func (dst *Cluster) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha3.Cluster)

	return Convert_v1alpha3_Cluster_To_v1alpha2_Cluster(src, dst, nil)
}

func (src *ClusterList) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha3.ClusterList)

	return Convert_v1alpha2_ClusterList_To_v1alpha3_ClusterList(src, dst, nil)
}

// nolint
func (dst *ClusterList) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha3.ClusterList)

	return Convert_v1alpha3_ClusterList_To_v1alpha2_ClusterList(src, dst, nil)
}

func (src *Machine) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha3.Machine)

	return Convert_v1alpha2_Machine_To_v1alpha3_Machine(src, dst, nil)
}

// nolint
func (dst *Machine) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha3.Machine)

	return Convert_v1alpha3_Machine_To_v1alpha2_Machine(src, dst, nil)
}

func (src *MachineList) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha3.MachineList)

	return Convert_v1alpha2_MachineList_To_v1alpha3_MachineList(src, dst, nil)
}

// nolint
func (dst *MachineList) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha3.MachineList)

	return Convert_v1alpha3_MachineList_To_v1alpha2_MachineList(src, dst, nil)
}

func (src *MachineSet) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha3.MachineSet)

	return Convert_v1alpha2_MachineSet_To_v1alpha3_MachineSet(src, dst, nil)
}

// nolint
func (dst *MachineSet) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha3.MachineSet)

	return Convert_v1alpha3_MachineSet_To_v1alpha2_MachineSet(src, dst, nil)
}

func (src *MachineSetList) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha3.MachineSetList)

	return Convert_v1alpha2_MachineSetList_To_v1alpha3_MachineSetList(src, dst, nil)
}

// nolint
func (dst *MachineSetList) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha3.MachineSetList)

	return Convert_v1alpha3_MachineSetList_To_v1alpha2_MachineSetList(src, dst, nil)
}

func (src *MachineDeployment) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha3.MachineDeployment)

	return Convert_v1alpha2_MachineDeployment_To_v1alpha3_MachineDeployment(src, dst, nil)
}

// nolint
func (dst *MachineDeployment) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha3.MachineDeployment)

	return Convert_v1alpha3_MachineDeployment_To_v1alpha2_MachineDeployment(src, dst, nil)
}

func (src *MachineDeploymentList) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha3.MachineDeploymentList)

	return Convert_v1alpha2_MachineDeploymentList_To_v1alpha3_MachineDeploymentList(src, dst, nil)
}

// nolint
func (dst *MachineDeploymentList) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha3.MachineDeploymentList)

	return Convert_v1alpha3_MachineDeploymentList_To_v1alpha2_MachineDeploymentList(src, dst, nil)
}

func (src *MachinePool) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha3.MachinePool)

	return Convert_v1alpha2_MachinePool_To_v1alpha3_MachinePool(src, dst, nil)
}

// nolint
func (dst *MachinePool) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha3.MachinePool)

	return Convert_v1alpha3_MachinePool_To_v1alpha2_MachinePool(src, dst, nil)
}

func (src *MachinePoolList) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha3.MachinePoolList)

	return Convert_v1alpha2_MachinePoolList_To_v1alpha3_MachinePoolList(src, dst, nil)
}

// nolint
func (dst *MachinePoolList) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha3.MachinePoolList)

	return Convert_v1alpha3_MachinePoolList_To_v1alpha2_MachinePoolList(src, dst, nil)
}

func Convert_v1alpha2_MachineSpec_To_v1alpha3_MachineSpec(in *MachineSpec, out *v1alpha3.MachineSpec, s apiconversion.Scope) error {
	if err := autoConvert_v1alpha2_MachineSpec_To_v1alpha3_MachineSpec(in, out, s); err != nil {
		return err
	}

	// Discards unused ObjectMeta

	return nil
}

func Convert_v1alpha3_MachineDeploymentSpec_To_v1alpha2_MachineDeploymentSpec(in *v1alpha3.MachineDeploymentSpec, out *MachineDeploymentSpec, s apiconversion.Scope) error {
	return errors.New("cannot recover removed MachineDeploymentSpec Cluster Name")
}

func Convert_v1alpha3_MachineDeploymentStatus_To_v1alpha2_MachineDeploymentStatus(in *v1alpha3.MachineDeploymentStatus, out *MachineDeploymentStatus, s apiconversion.Scope) error {
	return fmt.Errorf("cannot recover removed MachineDeploymentStatus field Phase")
}

func Convert_v1alpha3_MachineSetSpec_To_v1alpha2_MachineSetSpec(in *v1alpha3.MachineSetSpec, out *MachineSetSpec, s apiconversion.Scope) error {
	return errors.New("cannot recover removed MachineSetSpec Cluster Name")
}

func Convert_v1alpha3_MachineSpec_To_v1alpha2_MachineSpec(in *v1alpha3.MachineSpec, out *MachineSpec, s apiconversion.Scope) error {
	return errors.New("cannot recover removed MachineSpec Cluster Name")
}

func Convert_v1alpha3_MachinePoolSpec_To_v1alpha2_MachinePoolSpec(in *v1alpha3.MachinePoolSpec, out *MachinePoolSpec, s apiconversion.Scope) error {
	return errors.New("cannot recover removed MachinePoolSpec Cluster Name")
}
