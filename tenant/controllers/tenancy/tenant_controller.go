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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/multi-tenancy/tenant/controllers/tenancy/tenant"

	tenancyv1alpha1 "sigs.k8s.io/multi-tenancy/tenant/apis/tenancy/v1alpha1"
)

var (
	jobOwnerKey = ".metadata.controller"
	apiGVStr    = tenancyv1alpha1.GroupVersion.String()
)

// TenantReconciler reconciles a Tenant object
type TenantReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=tenancy.x-k8s.io,resources=tenants,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tenancy.x-k8s.io,resources=tenants/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch;create;update
// +kubebuilder:rbac:groups=core,resources=events,verbs=get;create
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles,verbs=get;list;create;update;patch
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings,verbs=get;list;create;update;patch
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles,verbs=get;list;create;update;patch
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;create;update;patch

// Reconcile will read the state of Tenant CRs and will create the namespaces and cluster roles/bindings.
func (r *TenantReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("tenant", req.NamespacedName)

	// Fetch the Tenant instance
	instance := &tenancyv1alpha1.Tenant{}
	// Tenant is a cluster scoped CR, we should clear the namespace field in request
	req.NamespacedName.Namespace = ""
	if err := r.Get(ctx, req.NamespacedName, instance); err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	// Create tenantAdminNamespace
	if instance.Spec.TenantAdminNamespaceName != "" {
		tenantAdminNs, err := tenant.ConstructNamespaceFromTenant(instance, r.Scheme)
		if err != nil {
			return ctrl.Result{}, err
		}

		if err := r.clientApplyObjects(ctx, []runtime.Object{tenantAdminNs}); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Create RBACs for tenantAdmins.
	if instance.Spec.TenantAdmins != nil {
		objects, err := tenant.ConstructRBACObjects(instance, r.Scheme)
		if err != nil {
			return ctrl.Result{}, err
		}

		if err := r.clientApplyObjects(ctx, objects); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager will initialize the manager for Tenants
func (r *TenantReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(&corev1.Namespace{}, jobOwnerKey, func(rawObj runtime.Object) []string {
		job := rawObj.(*corev1.Namespace)
		owner := metav1.GetControllerOf(job)
		if owner == nil {
			return nil
		}
		if owner.APIVersion != apiGVStr || owner.Kind != "Tenant" {
			return nil
		}

		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&tenancyv1alpha1.Tenant{}).
		Owns(&corev1.Namespace{}).
		Complete(r)
}

// clientApply create if not existing, update otherwise
func (r *TenantReconciler) clientApplyObjects(ctx context.Context, objects []runtime.Object) error {
	var err error
	for _, object := range objects {
		if err = r.Create(ctx, object); err != nil {
			if errors.IsAlreadyExists(err) {
				r.Recorder.Event(object, corev1.EventTypeNormal, "AlreadyExists", "Object exists, updating instead")
				if err = r.Update(ctx, object); err != nil {
					return err
				}
			}
			if err != nil && !errors.IsAlreadyExists(err) {
				r.Recorder.Event(object, corev1.EventTypeWarning, "UnknownError", "Object failed creating")
			}
		}
	}
	return err
}
