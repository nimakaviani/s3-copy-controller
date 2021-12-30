/*
Copyright 2021.

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
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	cloudobj "dev.nimak.link/s3-copy-controller/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Object controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		ObjName   = "test-obj"
		Namespace = "default"

		timeout  = time.Second * 30
		duration = time.Second * 30
		interval = time.Millisecond * 250
	)

	var (
		objLookupKey  = types.NamespacedName{Name: ObjName, Namespace: Namespace}
		createdObject *cloudobj.Object
	)

	AfterEach(func() {
		createdObject = &cloudobj.Object{}
		Eventually(func() bool {
			err := k8sClient.Get(ctx, objLookupKey, createdObject)
			return err == nil
		}, timeout, interval).Should(BeTrue())
		Expect(k8sClient.Delete(ctx, createdObject)).Should(Succeed())
	})

	Context("upon submitting an object", func() {
		BeforeEach(func() {
			By("By creating a new Object")
			ctx := context.Background()
			obj := &cloudobj.Object{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "s3.aws.dev.nimak.link/v1alpha1",
					Kind:       "Object",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      ObjName,
					Namespace: Namespace,
				},
				Spec: cloudobj.ObjectSpec{
					DeletionPolicy: "Delete",
					Target: cloudobj.ObjectTarget{
						Region: "us-west-2",
						Bucket: "test-bucket",
						Key:    "test.key",
					},
					Source: cloudobj.ObjectSource{
						Data: "test data",
					},
					Credentials: cloudobj.Credentials{
						Source: "Secret",
						SecretReference: cloudobj.SecretKeySelector{
							SecretReference: cloudobj.SecretReference{
								Namespace: "creds-ns",
								Name:      "creds-name",
							},
							Key: "creds-key",
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, obj)).Should(Succeed())
		})

		When("object becomes available in the cluster", func() {
			BeforeEach(func() {
				createdObject = &cloudobj.Object{}
				Eventually(func() bool {
					err := k8sClient.Get(ctx, objLookupKey, createdObject)
					return err == nil
				}, timeout, interval).Should(BeTrue())
			})

			It("should have the correct object created in the cluster", func() {
				Expect(createdObject.Spec.DeletionPolicy).Should(Equal("Delete"))
				Expect(createdObject.ObjectMeta.Name).Should(Equal("test-obj"))

				By("not having the secret available in the cluster")
				Consistently(func() (bool, error) {
					err := k8sClient.Get(ctx, objLookupKey, createdObject)
					if err != nil {
						return false, err
					}
					return createdObject.Status.Synced == "", nil
				}, timeout, interval, "should not have its status synced").Should(BeTrue())
			})
		})

	})

	Context("with secret present", func() {
		BeforeEach(func() {
			By(" creating secret")
			ctx := context.Background()
			secret := &corev1.Secret{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Secret",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "creds-name",
					Namespace: "default",
				},
				Type: corev1.SecretTypeOpaque,
				Data: map[string][]byte{
					"creds-key": []byte("c29tZS1kYXRh"),
				},
			}
			Expect(k8sClient.Create(ctx, secret)).Should(Succeed())

		})

		AfterEach(func() {
			createdSecret := &corev1.Secret{}
			secretLookupKey := types.NamespacedName{Name: "creds-name", Namespace: "default"}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, secretLookupKey, createdSecret)
				return err == nil
			}, timeout, interval).Should(BeTrue())
			Expect(k8sClient.Delete(ctx, createdSecret)).Should(Succeed())
		})

		It("shoudl create secret", func() {
			createdSecret := &corev1.Secret{}
			secretLookupKey := types.NamespacedName{Name: "creds-name", Namespace: "default"}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, secretLookupKey, createdSecret)
				return err == nil
			}, timeout, interval).Should(BeTrue())
		})

		When("creating the object", func() {
			BeforeEach(func() {
				By("By creating a new Object")
				ctx := context.Background()
				obj := &cloudobj.Object{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "s3.aws.dev.nimak.link/v1alpha1",
						Kind:       "Object",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      ObjName,
						Namespace: Namespace,
					},
					Spec: cloudobj.ObjectSpec{
						DeletionPolicy: "Delete",
						Target: cloudobj.ObjectTarget{
							Region: "us-west-2",
							Bucket: "test-bucket",
							Key:    "test.key",
						},
						Source: cloudobj.ObjectSource{
							Ref:  "local",
							Data: "test-data",
						},
						Credentials: cloudobj.Credentials{
							Source: "Secret",
							SecretReference: cloudobj.SecretKeySelector{
								SecretReference: cloudobj.SecretReference{
									Namespace: "default",
									Name:      "creds-name",
								},
								Key: "creds-key",
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, obj)).Should(Succeed())

				By("waiting for the object to become available")
				createdObject = &cloudobj.Object{}

				// We'll need to retry getting this newly created Object, given that creation may not immediately happen.
				Eventually(func() bool {
					err := k8sClient.Get(ctx, objLookupKey, createdObject)
					return err == nil
				}, timeout, interval).Should(BeTrue())
			})

			FIt("the controller should pull the StoreManager", func() {
				Eventually(fakeStoreManager.GetCallCount, timeout, interval).Should(Equal(1))
				cfg := fakeStoreManager.GetArgsForCall(0)
				Expect(cfg.Region).To(Equal("us-west-2"))
			})

			It("should try to add object data into object store", func() {
				Eventually(fakeObjectStore.StoreCallCount, timeout, interval).Should(Equal(1))
				Eventually(func() (string, error) {
					_, data, _ := fakeObjectStore.StoreArgsForCall(0)
					return string(data), nil
				}, timeout, interval).Should(Equal("test-data"))
			})
		})
	})

})
