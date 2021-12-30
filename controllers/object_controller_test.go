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
		ObjName    = "test-obj"
		SecretName = "creds-name"
		Namespace  = "default"

		timeout  = time.Second * 30
		duration = time.Second * 30
		interval = time.Millisecond * 250
	)

	var (
		objLookupKey    = types.NamespacedName{Name: ObjName, Namespace: Namespace}
		secretLookupKey = types.NamespacedName{Name: SecretName, Namespace: Namespace}
		createdObject   *cloudobj.Object
	)

	Context("with secret present", func() {
		AfterEach(func() {
			// delete object
			obj := &cloudobj.Object{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, objLookupKey, obj)
				return err == nil
			}, timeout, interval).Should(BeTrue())
			Expect(k8sClient.Delete(ctx, obj)).Should(Succeed())
			Eventually(func() bool {
				err := k8sClient.Get(ctx, objLookupKey, obj)
				return err == nil
			}, timeout, interval).Should(BeFalse())

			// delete secret
			secret := &corev1.Secret{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, secretLookupKey, secret)
				return err == nil
			}, timeout, interval).Should(BeTrue())
			Expect(k8sClient.Delete(ctx, secret)).Should(Succeed())
			Eventually(func() bool {
				err := k8sClient.Get(ctx, secretLookupKey, secret)
				return err == nil
			}, timeout, interval).Should(BeFalse())
		})

		It("should successfully try to store the object", func() {
			By("having the correct secret present")
			ctx := context.Background()
			secret := &corev1.Secret{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Secret",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      SecretName,
					Namespace: Namespace,
				},
				Type: corev1.SecretTypeOpaque,
				Data: map[string][]byte{
					"creds-key": []byte("c29tZS1kYXRh"),
				},
			}
			Expect(k8sClient.Create(ctx, secret)).Should(Succeed())

			By("having the new Object defined")
			ctx = context.Background()
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

			By("submitting the new object")
			Expect(k8sClient.Create(ctx, obj)).Should(Succeed())

			By("waiting for the object to become available")
			createdObject = &cloudobj.Object{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, objLookupKey, createdObject)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			By("makes the call to StoreManager")
			Eventually(fakeStoreManager.GetCallCount, timeout, interval).Should(BeNumerically(">", 0))
			cfg := fakeStoreManager.GetArgsForCall(0)
			Expect(cfg.Region).To(Equal("us-west-2"))

			By("uses ObjectStore to save content")
			Eventually(fakeObjectStore.StoreCallCount, timeout, interval).Should(BeNumerically(">", 0))
			Eventually(func() (string, error) {
				_, data, _ := fakeObjectStore.StoreArgsForCall(0)
				return string(data), nil
			}, timeout, interval).Should(Equal("test-data"))

			By("retrieving the object after storing the data")
			updatedObject := &cloudobj.Object{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, objLookupKey, updatedObject)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			By("object status should reflect updates")
			Expect(updatedObject.Status.Synced).To(BeTrue())
			Expect(updatedObject.Status.Reference).Should(Equal("s3://test-bucket/test.key"))
		})

		It("should successfully try to delete the object", func() {
			By("having the correct secret present")
			ctx := context.Background()
			secret := &corev1.Secret{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Secret",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      SecretName,
					Namespace: Namespace,
				},
				Type: corev1.SecretTypeOpaque,
				Data: map[string][]byte{
					"creds-key": []byte("c29tZS1kYXRh"),
				},
			}
			Expect(k8sClient.Create(ctx, secret)).Should(Succeed())

			By("having the new Object defined")
			ctx = context.Background()
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

			By("submitting the new object")
			Expect(k8sClient.Create(ctx, obj)).Should(Succeed())

			By("waiting for the object to become available")
			createdObject = &cloudobj.Object{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, objLookupKey, createdObject)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			By("deleting the object")
			Expect(k8sClient.Delete(ctx, obj)).Should(Succeed())

			By("makes the call to StoreManager")
			Eventually(fakeStoreManager.GetCallCount, timeout, interval).Should(BeNumerically(">", 0))
			cfg := fakeStoreManager.GetArgsForCall(0)
			Expect(cfg.Region).To(Equal("us-west-2"))

			By("uses ObjectStore to save content")
			Eventually(fakeObjectStore.DeleteCallCount, timeout, interval).Should(BeNumerically(">", 0))
			Eventually(func() bool {
				_, target := fakeObjectStore.DeleteArgsForCall(0)
				return target.Bucket == "test-bucket"
			}, timeout, interval).Should(BeTrue())
		})
	})

	Context("without secret present", func() {
		AfterEach(func() {
			// delete object
			obj := &cloudobj.Object{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, objLookupKey, obj)
				return err == nil
			}, timeout, interval).Should(BeTrue())
			Expect(k8sClient.Delete(ctx, obj)).Should(Succeed())
		})

		It("should successfully try to store the object", func() {
			By("having the new Object defined")
			ctx = context.Background()
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

			By("submitting the new object")
			Expect(k8sClient.Create(ctx, obj)).Should(Succeed())

			By("waiting for the object to become available")
			createdObject = &cloudobj.Object{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, objLookupKey, createdObject)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			By("object status should reflect updates")
			Expect(createdObject.Status.Synced).To(BeFalse())
			Expect(createdObject.Status.Reference).Should(BeEmpty())
		})
	})
})
