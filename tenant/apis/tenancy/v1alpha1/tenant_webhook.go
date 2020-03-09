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

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var tenantlog = logf.Log.WithName("tenant-resource")

// SetupWebhookWithManager will setup the Webhook manager
func (r *Tenant) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-tenancy-x-k8s-io-v1alpha1-tenant,mutating=false,failurePolicy=fail,groups=tenancy.x-k8s.io,resources=tenants,versions=v1alpha1,name=vtenant.kb.io

var _ webhook.Validator = &Tenant{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Tenant) ValidateCreate() error {
	tenantlog.Info("validate create", "name", r.Name)
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Tenant) ValidateUpdate(old runtime.Object) error {
	tenantlog.Info("validate update", "name", r.Name)
	allErrs := field.ErrorList{}
	oldobj := old.(*Tenant)

	if err := r.validateImmutableTenantAdminNamespaceName(oldobj); err != nil {
		allErrs = append(allErrs, err)
	}

	if err := r.validateImmutableRequireNamespacePrefix(oldobj); err != nil {
		allErrs = append(allErrs, err)
	}

	return errors.NewInvalid(
		schema.GroupKind{Group: GroupVersion.Group, Kind: "Tenant"},
		r.Name, allErrs)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Tenant) ValidateDelete() error {
	tenantlog.Info("validate delete", "name", r.Name)
	return nil
}
