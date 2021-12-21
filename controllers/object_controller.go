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
&selimitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cloudobject "dev.nimak.link/s3-copy-controller/api/v1alpha1"
)

// ObjectReconciler reconciles a Object object
type ObjectReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=s3.aws.dev.nimak.link,resources=objects,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=s3.aws.dev.nimak.link,resources=objects/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=s3.aws.dev.nimak.link,resources=objects/finalizers,verbs=update
//+kubebuilder:rbac:resources=secrets,verbs=get
//+kubebuilder:rbac:resources=configmaps,verbs=get

func (r *ObjectReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// secret := &corev1.Secret{}
	var childObj cloudobject.Object
	if err := r.Get(ctx, req.NamespacedName, &childObj); err != nil {
		log.Error(err, "fetching secret failed")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	go r.process(ctx, &childObj)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ObjectReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cloudobject.Object{}).
		Complete(r)
}

func (r *ObjectReconciler) process(ctx context.Context, obj *cloudobject.Object) {
	log := log.FromContext(ctx)
	secretData, err := r.getSecret(ctx, obj)
	if err != nil {
		log.Error(err, "failed to retrieve secret")
	}

	cfg, err := useProviderSecret(ctx, secretData, DefaultProfile, obj.Spec.Target.Region)
	if err != nil {
		log.Error(err, "failed to create aws config")
	}

	objData, err := r.getContent(ctx, obj)
	if err != nil {
		log.Error(err, "failed to pull object content")
	}

	if err := NewS3ObjectStore(cfg).Store(ctx, objData, obj.Spec.Target); err != nil {
		log.Error(err, "failed to store to object store")
		return
	}

	log.Info("data sync succeeded", getReference(obj))
}

func (r *ObjectReconciler) getSecret(ctx context.Context, obj *cloudobject.Object) ([]byte, error) {
	creds := obj.Spec.Credentials
	if creds.Source != "" && creds.Source != "Secret" {
		return nil, errors.Errorf("wrong source %s", creds.Source)
	}

	var secret corev1.Secret
	secretRef := types.NamespacedName{Namespace: creds.SecretReference.Namespace, Name: creds.SecretReference.Name}
	if err := r.Get(ctx, secretRef, &secret); err != nil {
		return nil, errors.Errorf("unrecognized credentials %s:%s", creds.SecretReference.Namespace, creds.SecretReference.Name)
	}

	secretData, ok := secret.Data[creds.SecretReference.Key]
	if !ok {
		return nil, errors.Errorf("key not found %s", creds.SecretReference.Key)
	}
	return secretData, nil
}

func (r *ObjectReconciler) getContent(ctx context.Context, obj *cloudobject.Object) ([]byte, error) {
	src := obj.Spec.Source
	switch strings.ToLower(src.Ref) {
	case "", "local":
		if src.Data == "" {
			return nil, errors.New("data field required for a 'local' reference")
		}
		return []byte(src.Data), nil

	case "configmap":
		var cm corev1.ConfigMap
		dataRef := types.NamespacedName{Namespace: src.Namespace, Name: src.Name}
		if err := r.Get(ctx, dataRef, &cm); err != nil {
			return nil, errors.Errorf("unrecognized configmap %s:%s", src.Namespace, src.Name)
		}
		data, ok := cm.Data[src.Key]
		if !ok || src.Key == "" {
			return nil, errors.Errorf("key not found %s", src.Key)
		}
		return []byte(data), nil

	default:
		return nil, errors.Errorf("source invalid")
	}
}

func getReference(obj *cloudobject.Object) string {
	return fmt.Sprintf("%s -> %s:%s",
		obj.Name,
		obj.Spec.Target.Bucket, obj.Spec.Target.Key,
	)
}
