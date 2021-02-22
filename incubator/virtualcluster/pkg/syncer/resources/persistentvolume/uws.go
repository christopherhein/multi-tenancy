/*
Copyright 2020 The Kubernetes Authors.

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

package persistentvolume

import (
	"context"
	"fmt"

	pkgerr "github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"

	"sigs.k8s.io/multi-tenancy/incubator/virtualcluster/pkg/syncer/constants"
	"sigs.k8s.io/multi-tenancy/incubator/virtualcluster/pkg/syncer/conversion"
)

// StartUWS starts the upward syncer
// and blocks until an empty struct is sent to the stop channel.
func (c *controller) StartUWS(stopCh <-chan struct{}) error {
	if !cache.WaitForCacheSync(stopCh, c.pvSynced, c.pvcSynced) {
		return fmt.Errorf("failed to wait for caches to sync persistentvolume")
	}
	return c.UpwardController.Start(stopCh)
}

func (c *controller) BackPopulate(key string) error {
	ctx := context.Background()
	pPV, err := c.pvLister.Get(key)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	if pPV.Spec.ClaimRef == nil {
		return nil
	}

	pPVC, err := c.pvcLister.PersistentVolumeClaims(pPV.Spec.ClaimRef.Namespace).Get(pPV.Spec.ClaimRef.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			// Bound PVC is gone, we cannot find the tenant who owns the pv. Checker will fix any possible race.
			return nil
		}
		return err
	}

	clusterName, vNamespace := conversion.GetVirtualOwner(pPVC)
	if clusterName == "" {
		// Bound PVC does not belong to any tenant.
		return nil
	}

	tenantClient, err := c.MultiClusterController.GetClusterClient(clusterName)
	if err != nil {
		return pkgerr.Wrapf(err, "failed to create client from cluster %s config", clusterName)
	}

	vPV := &v1.PersistentVolume{}
	objectKey := types.NamespacedName{Namespace: "", Name: key}
	if err := c.MultiClusterController.Get(ctx, clusterName, objectKey, vPV); err != nil {
		if errors.IsNotFound(err) {
			// Create a new pv with bound claim in tenant master

			vPVC := &v1.PersistentVolumeClaim{}
			objectKey := types.NamespacedName{Namespace: vNamespace, Name: pPVC.Name}
			if err := tenantClient.Get(ctx, objectKey, vPVC); err != nil {
				// If corresponding pvc does not exist in tenant, we'll let checker fix any possible race.
				klog.Errorf("Cannot find the bound pvc %s/%s in tenant cluster %s for pv %v", vNamespace, pPVC.Name, clusterName, pPV)
				return nil
			}
			vcName, vcNS, _, err := c.MultiClusterController.GetOwnerInfo(clusterName)
			if err != nil {
				return err
			}
			vPV = conversion.BuildVirtualPersistentVolume(clusterName, vcNS, vcName, pPV, vPVC)
			if err = tenantClient.Create(ctx, vPV); err != nil {
				return err
			}
			return nil
		}
		return err
	}

	if vPV.Annotations[constants.LabelUID] != string(pPV.UID) {
		return fmt.Errorf("vPV %s in cluster %s delegated UID is different from pPV.", vPV.Name, clusterName)
	}

	// We only update PV.Spec, PV.Status is managed by tenant/super pv binder controller independently.
	updatedPVSpec := conversion.Equality(c.Config, nil).CheckPVSpecEquality(&pPV.Spec, &vPV.Spec)
	if updatedPVSpec != nil {
		newPV := vPV.DeepCopy()
		newPV.Spec = *updatedPVSpec
		if err := tenantClient.Update(ctx, newPV); err != nil {
			return err
		}
	}
	return nil
}
