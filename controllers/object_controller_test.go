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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Object controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		ObjName   = "test-obj"
		Namespace = "default"

		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When updating Object Status", func() {
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

		It("should create the object in the cluser", func() {
			jobLookupKey := types.NamespacedName{Name: ObjName, Namespace: Namespace}
			createdObj := &cloudobj.Object{}

			// We'll need to retry getting this newly created CronJob, given that creation may not immediately happen.
			Eventually(func() bool {
				err := k8sClient.Get(ctx, jobLookupKey, createdObj)
				return err == nil
			}, timeout, interval).Should(BeTrue())
			Expect(createdObj.Spec.DeletionPolicy).Should(Equal("Delete"))
		})
	})
})
