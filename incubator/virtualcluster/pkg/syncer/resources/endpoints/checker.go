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

package endpoints

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"

	"sigs.k8s.io/multi-tenancy/incubator/virtualcluster/pkg/syncer/conversion"
	"sigs.k8s.io/multi-tenancy/incubator/virtualcluster/pkg/syncer/metrics"
)

var numMissingEndPoints uint64
var numMissMatchedEndPoints uint64

func (c *controller) StartPatrol(stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()

	if !cache.WaitForCacheSync(stopCh, c.endpointsSynced) {
		return fmt.Errorf("failed to wait for caches to sync before starting Endpoint checker")
	}
	c.Patroller.Start(stopCh)
	return nil
}

// PatrollerDo checks to see if Endpoints in super master informer cache and tenant master
// keep consistency.
// Note that eps are managed by tenant/super ep controller separately. The checker will not do GC but only report diff.
func (c *controller) PatrollerDo() {
	ctx := context.Background()
	clusterNames := c.MultiClusterController.GetClusterNames()
	if len(clusterNames) == 0 {
		klog.Infof("tenant masters has no clusters, give up period checker")
		return
	}

	numMissingEndPoints = 0
	numMissMatchedEndPoints = 0
	wg := sync.WaitGroup{}

	for _, clusterName := range clusterNames {
		wg.Add(1)
		go func(clusterName string) {
			defer wg.Done()
			c.checkEndPointsOfTenantCluster(ctx, clusterName)
		}(clusterName)
	}
	wg.Wait()
	metrics.CheckerMissMatchStats.WithLabelValues("MissingEndPoints").Set(float64(numMissingEndPoints))
	metrics.CheckerMissMatchStats.WithLabelValues("MissMatchedEndPoints").Set(float64(numMissMatchedEndPoints))
}

// checkEndPointsOfTenantCluster checks to see if endpoints controller in tenant and super master working consistently.
func (c *controller) checkEndPointsOfTenantCluster(ctx context.Context, clusterName string) {
	epList := &v1.EndpointsList{}
	if err := c.MultiClusterController.List(ctx, clusterName, epList); err != nil {
		klog.Errorf("error listing endpoints from cluster %s informer cache: %v", clusterName, err)
		return
	}
	klog.V(4).Infof("check endpoints consistency in cluster %s", clusterName)
	for _, vEp := range epList.Items {
		targetNamespace := conversion.ToSuperMasterNamespace(clusterName, vEp.Namespace)
		pEp, err := c.endpointsLister.Endpoints(targetNamespace).Get(vEp.Name)
		if errors.IsNotFound(err) {
			// pEp not found and vEp still exists, report the inconsistent ep controller behavior
			klog.Errorf("Cannot find pEp %v/%v in super master", targetNamespace, vEp.Name)
			atomic.AddUint64(&numMissingEndPoints, 1)
			continue
		}
		if err != nil {
			klog.Errorf("error getting pEp %s/%s from super master cache: %v", targetNamespace, vEp.Name, err)
			continue
		}
		updated := conversion.Equality(c.Config, nil).CheckEndpointsEquality(pEp, &vEp)
		if updated != nil {
			atomic.AddUint64(&numMissMatchedEndPoints, 1)
			klog.Warningf("Endpoint %v/%v diff in super&tenant master", targetNamespace, vEp.Name)
		}
	}
}
