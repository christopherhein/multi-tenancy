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

package controllers

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	tenancyv1alpha1 "sigs.k8s.io/multi-tenancy/tenant/apis/tenancy/v1alpha1"
)

var _ = Describe("Tenant Controller", func() {

	const timeout = time.Second * 30
	const interval = time.Second * 1

	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	Context("Create owned resources", func() {
		It("Should create successfully", func() {

			instance := &tenancyv1alpha1.Tenant{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "tenant-1",
				},
				Spec: tenancyv1alpha1.TenantSpec{
					TenantAdminNamespaceName: "tenant-admin-1",
					RequireNamespacePrefix:   true,
					TenantAdmins: []rbacv1.Subject{
						rbacv1.Subject{
							Kind: "Group",
							Name: "foo",
						},
					},
				},
			}

			// Create
			Expect(k8sClient.Create(context.Background(), instance)).Should(Succeed())

			// Set
			key := types.NamespacedName{Name: instance.Name}
			nskey := types.NamespacedName{Name: instance.Spec.TenantAdminNamespaceName}
			crkey := types.NamespacedName{Name: fmt.Sprintf("%s-tenant-admin-role", instance.Name)}

			By("Expecting Namespace created")
			Eventually(func() bool {
				f := &corev1.Namespace{}
				err := k8sClient.Get(context.Background(), nskey, f)
				return !errors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By(fmt.Sprintf("Expecting CR RBAC Object"))
			Eventually(func() bool {
				f := &rbacv1.ClusterRole{}
				err := k8sClient.Get(context.Background(), crkey, f)
				return !errors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			instanceCopy := &tenancyv1alpha1.Tenant{}
			Expect(k8sClient.Get(context.Background(), key, instanceCopy)).Should(Succeed())

			instanceCopy.Spec.TenantAdmins = []rbacv1.Subject{
				rbacv1.Subject{
					Kind: "Group",
					Name: "foo",
				},
				rbacv1.Subject{
					Kind: "Group",
					Name: "Bar",
				},
			}
			Expect(k8sClient.Update(context.Background(), instanceCopy)).Should(Succeed())

			By("Labels propagating to the Namespace")
			Eventually(func() bool {
				f := &tenancyv1alpha1.Tenant{}
				_ = k8sClient.Get(context.Background(), key, f)
				return len(f.Spec.TenantAdmins) == 2
			}, timeout, interval).Should(BeTrue())

			// // Update
			// updated := &databricksv1alpha1.DbfsBlock{}
			// Expect(k8sClient.Get(context.Background(), key, updated)).Should(Succeed())

			// updated.Spec.Data = dataStr2
			// Expect(k8sClient.Update(context.Background(), updated)).Should(Succeed())

			// By("Expecting size to be 5500")
			// Eventually(func() int64 {
			// 	f := &databricksv1alpha1.DbfsBlock{}
			// 	_ = k8sClient.Get(context.Background(), key, f)
			// 	return f.Status.FileInfo.FileSize
			// }, timeout, interval).Should(Equal(int64(5500)))

			// // Delete
			// By("Expecting to delete successfully")
			// Eventually(func() error {
			// 	f := &databricksv1alpha1.DbfsBlock{}
			// 	_ = k8sClient.Get(context.Background(), key, f)
			// 	return k8sClient.Delete(context.Background(), f)
			// }, timeout, interval).Should(Succeed())

			// By("Expecting to delete finish")
			// Eventually(func() error {
			// 	f := &databricksv1alpha1.DbfsBlock{}
			// 	return k8sClient.Get(context.Background(), key, f)
			// }, timeout, interval).ShouldNot(Succeed())
		})
	})
})
