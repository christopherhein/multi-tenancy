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

package tenant

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	tenancyv1alpha1 "sigs.k8s.io/multi-tenancy/tenant/apis/tenancy/v1alpha1"
)

// ConstructNamespaceFromTenant will return a new constructed namespace w/ owner references
func ConstructNamespaceFromTenant(instance *tenancyv1alpha1.Tenant, scheme *runtime.Scheme) (*corev1.Namespace, error) {
	ns := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        instance.Spec.TenantAdminNamespaceName,
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		},
	}

	for k, v := range instance.GetLabels() {
		ns.Labels[k] = v
	}

	if err := ctrl.SetControllerReference(instance, ns, scheme); err != nil {
		return nil, err
	}

	return ns, nil
}

// ConstructRBACObjects will return the RBAC objects
func ConstructRBACObjects(instance *tenancyv1alpha1.Tenant, scheme *runtime.Scheme) ([]runtime.Object, error) {
	objects := []runtime.Object{}

	crName := fmt.Sprintf("%s-tenant-admin-role", instance.Name)
	cr := &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			APIVersion: rbacv1.SchemeGroupVersion.String(),
			Kind:       "ClusterRole",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: crName,
		},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:         []string{"get", "list", "watch", "update", "patch", "delete"},
				APIGroups:     []string{tenancyv1alpha1.GroupVersion.Group},
				Resources:     []string{"tenants"},
				ResourceNames: []string{instance.Name},
			},
			{
				Verbs:         []string{"get", "list", "watch"},
				APIGroups:     []string{""},
				Resources:     []string{"namespaces"},
				ResourceNames: []string{instance.Spec.TenantAdminNamespaceName},
			},
		},
	}
	if err := ctrl.SetControllerReference(instance, cr, scheme); err != nil {
		return objects, err
	}
	objects = append(objects, cr)

	crbinding := &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: rbacv1.SchemeGroupVersion.String(),
			Kind:       "ClusterRoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-tenant-admins-rolebinding", instance.Name),
		},
		Subjects: instance.Spec.TenantAdmins,
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "ClusterRole",
			Name:     crName,
		},
	}
	if err := ctrl.SetControllerReference(instance, crbinding, scheme); err != nil {
		return objects, err
	}
	objects = append(objects, crbinding)

	role := &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			APIVersion: rbacv1.SchemeGroupVersion.String(),
			Kind:       "Role",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tenant-admin-role",
			Namespace: instance.Spec.TenantAdminNamespaceName,
		},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
				APIGroups: []string{tenancyv1alpha1.GroupVersion.Group},
				Resources: []string{"tenantnamespaces"},
			},
		},
	}
	if err := ctrl.SetControllerReference(instance, role, scheme); err != nil {
		return objects, err
	}
	objects = append(objects, role)

	rbinding := &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: rbacv1.SchemeGroupVersion.String(),
			Kind:       "RoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tenant-admins-rolebinding",
			Namespace: instance.Spec.TenantAdminNamespaceName,
		},
		Subjects: instance.Spec.TenantAdmins,
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "Role",
			Name:     "tenant-admin-role",
		},
	}
	if err := ctrl.SetControllerReference(instance, rbinding, scheme); err != nil {
		return objects, err
	}
	if len(instance.Spec.TenantAdmins) > 0 {
		objects = append(objects, rbinding)
	}

	return objects, nil
}
