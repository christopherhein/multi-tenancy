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
	"fmt"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

func (r *Tenant) validateImmutableTenantAdminNamespaceName(old *Tenant) *field.Error {
	if r.Spec.TenantAdminNamespaceName != old.Spec.TenantAdminNamespaceName {
		return field.Forbidden(field.NewPath("spec").Child("tenantAdminNamespaceName"), fmt.Sprintf("cannot modify the tenantAdminNamespaceName field in spec after initial creation (attempting to change from %s to %s)", old.Spec.TenantAdminNamespaceName, r.Spec.TenantAdminNamespaceName))
	}
	return nil
}

func (r *Tenant) validateImmutableRequireNamespacePrefix(old *Tenant) *field.Error {
	if r.Spec.RequireNamespacePrefix != old.Spec.RequireNamespacePrefix {
		return field.Forbidden(field.NewPath("spec").Child("requireNamespacePrefix"), fmt.Sprintf("cannot modify the requireNamespacePrefix field in spec after initial creation (attempting to change from %v to %v)", old.Spec.RequireNamespacePrefix, r.Spec.RequireNamespacePrefix))
	}
	return nil
}
