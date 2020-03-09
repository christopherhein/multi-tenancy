/*
Copyright 2020 The Kubernetes authors.

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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	tenancyv1alpha1 "sigs.k8s.io/multi-tenancy/tenant/apis/tenancy/v1alpha1"
)

// TenantNamespaceReconciler reconciles a TenantNamespace object
type TenantNamespaceReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=tenancy.x-k8s.io,resources=tenantnamespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tenancy.x-k8s.io,resources=tenantnamespaces/status,verbs=get;update;patch

// Reconcile will read the state of TenantNamespace CRs and will create the ownership
func (r *TenantNamespaceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("tenantnamespace", req.NamespacedName)

	// your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager will initialize the manager for TenantNamespaces
func (r *TenantNamespaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tenancyv1alpha1.TenantNamespace{}).
		Complete(r)
}
