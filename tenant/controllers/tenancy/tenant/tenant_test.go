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
	"testing"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	tenancyv1alpha1 "sigs.k8s.io/multi-tenancy/tenant/apis/tenancy/v1alpha1"
)

func init() {
	_ = scheme.AddToScheme(scheme.Scheme)
	_ = tenancyv1alpha1.AddToScheme(scheme.Scheme)
}

func constructTenant(taNS string) *tenancyv1alpha1.Tenant {
	if taNS == "" {
		taNS = "tenant-admin-1"
	}
	return &tenancyv1alpha1.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "tenant-1",
			Labels: map[string]string{
				"test-label": "exists",
			},
		},
		Spec: tenancyv1alpha1.TenantSpec{
			TenantAdminNamespaceName: taNS,
			RequireNamespacePrefix:   true,
			TenantAdmins: []rbacv1.Subject{
				rbacv1.Subject{
					Kind: "Group",
					Name: "foo",
				},
			},
		},
	}
}

func constructNamespaceList(instance *tenancyv1alpha1.Tenant) *corev1.NamespaceList {
	namespace, _ := ConstructNamespaceFromTenant(instance, scheme.Scheme)
	return &corev1.NamespaceList{Items: []corev1.Namespace{*namespace}}
}

func TestConstructNamespaceFromTenant(t *testing.T) {
	instance := constructTenant("")

	type args struct {
		instance  *tenancyv1alpha1.Tenant
		labelName string
	}
	tests := []struct {
		name      string
		args      args
		wantName  string
		wantOwner string
		wantLabel string
		wantErr   bool
	}{
		{"TestConstructNamespaceFromTenant", args{instance, "test-label"}, "tenant-admin-1", instance.GetName(), "exists", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConstructNamespaceFromTenant(tt.args.instance, scheme.Scheme)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConstructNamespaceFromTenant() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.GetName() != tt.wantName {
				t.Errorf("ConstructNamespaceFromTenant().GetName() = %v, want %v", got.GetName(), tt.wantName)
			}

			if got.GetOwnerReferences()[0].Name != tt.wantOwner {
				t.Errorf("ConstructNamespaceFromTenant().GetOwnerRefs()[0] = %v, want %v", got.GetOwnerReferences()[0].Name, tt.wantOwner)
			}

			if got.GetLabels()[tt.args.labelName] != tt.wantLabel {
				t.Errorf("ConstructNamespaceFromTenant().GetLabels()[] = %v, want %v", got.GetLabels()[tt.args.labelName], tt.wantLabel)
			}
		})
	}
}

func TestConstructRBACObjects(t *testing.T) {
	instance := constructTenant("")
	type args struct {
		instance *tenancyv1alpha1.Tenant
		scheme   *runtime.Scheme
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{"TestGettingAllRBACResources", args{instance, scheme.Scheme}, 4, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConstructRBACObjects(tt.args.instance, tt.args.scheme)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConstructRBACObjects() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.want {
				t.Errorf("ConstructRBACObjects() = %v, want %v", len(got), tt.want)
			}
		})
	}
}
