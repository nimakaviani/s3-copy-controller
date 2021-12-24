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
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cloudobject "dev.nimak.link/s3-copy-controller/api/v1alpha1"
	awshelper "dev.nimak.link/s3-copy-controller/controllers/aws"
)

// ObjectReconciler reconciles a Object object
type ObjectReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

const (
	ObjectFinalizer = "s3.aws.dev.nimak.link/finalizer"
	Failed          = "Failed"
	Synced          = "Synced"
	Removed         = "Removed"

	// switch elements
	Delete    = "delete"
	Retain    = "retain"
	Local     = "local"
	ConfigMap = "configmap"
	Empty     = ""
)

type Action int

const (
	StoreAction = iota
	DeleteAction
)

//+kubebuilder:rbac:groups=s3.aws.dev.nimak.link,resources=objects,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=s3.aws.dev.nimak.link,resources=objects/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=s3.aws.dev.nimak.link,resources=objects/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch
//+kubebuilder:rbac:resources=configmaps,verbs=get
//+kubebuilder:rbac:resources=secrets,verbs=get

func (r *ObjectReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var obj cloudobject.Object
	if err := r.Get(ctx, req.NamespacedName, &obj); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if obj.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(&obj, ObjectFinalizer) {
			controllerutil.AddFinalizer(&obj, ObjectFinalizer)
			if err := r.Update(ctx, &obj); err != nil {
				return ctrl.Result{}, err
			}
		}
		// process object creation / update
		if err := r.process(ctx, &obj, StoreAction); err != nil {
			return ctrl.Result{}, err
		}
	} else {
		if controllerutil.ContainsFinalizer(&obj, ObjectFinalizer) {
			if err := r.deleteExternalResources(ctx, &obj); err != nil {
				return ctrl.Result{}, err
			}

			controllerutil.RemoveFinalizer(&obj, ObjectFinalizer)
			if err := r.Update(ctx, &obj); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	// stop reconciliation after processing resource
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ObjectReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cloudobject.Object{}).
		Complete(r)
}

func (r *ObjectReconciler) process(ctx context.Context, obj *cloudobject.Object, action Action) (controllerError error) {
	var (
		// we use err to capture non-controller errors and
		// handle them separately for external operations
		// via the deferred function `processError`
		err error

		secretData []byte
		objData    []byte
		cfg        *aws.Config
	)

	log := log.FromContext(ctx)
	log.Info("processing resource", "key", client.ObjectKeyFromObject(obj), "action", action)

	defer func() {
		controllerError = r.processError(ctx, obj, action, &err)
	}()

	if secretData, err = r.getSecret(ctx, obj); err != nil {
		return
	}

	if cfg, err = awshelper.UseProviderSecret(ctx, secretData, awshelper.DefaultProfile, obj.Spec.Target.Region); err != nil {
		return
	}

	if objData, err = r.getContent(ctx, obj); err != nil {
		return
	}

	objectStore := awshelper.NewS3ObjectStore(cfg)
	switch action {
	case StoreAction:
		if err = objectStore.Store(ctx, objData, obj.Spec.Target); err != nil {
			return
		}

		obj.Status.Synced = strconv.FormatBool(true)
		obj.Status.Reference = fmt.Sprintf("s3://%s/%s", obj.Spec.Target.Bucket, obj.Spec.Target.Key)
		if controllerError = r.Status().Update(ctx, obj); controllerError != nil {
			return
		}

		r.Recorder.Event(obj, corev1.EventTypeNormal, Synced, fmt.Sprintf("object reference: %s", getReference(obj)))
		log.Info("successfully synced resource", "key", getReference(obj))

	case DeleteAction:
		switch strings.ToLower(obj.Spec.DeletionPolicy) {
		case Delete:
			if err = objectStore.Delete(ctx, obj.Spec.Target); err != nil {
				return
			}
			log.Info("successfully deleted resource", "key", getReference(obj))
		case Retain:
			log.Info("retaining the object in the object store")
			// do nothing
		default:
			err = errors.Errorf("invalid deletionPolicy %s", obj.Spec.DeletionPolicy)
		}
	}

	return nil
}

func (r *ObjectReconciler) processError(ctx context.Context, obj *cloudobject.Object, action Action, processingError *error) error {
	pe := *processingError
	if pe == nil {
		return nil
	}

	log := log.FromContext(ctx)
	log.Error(pe, "failed to sync resource")

	obj.Status.Synced = strconv.FormatBool(false)
	obj.Status.Reference = ""
	r.Recorder.Event(obj, corev1.EventTypeWarning, Failed, pe.Error())
	if err := r.Status().Update(ctx, obj); err != nil {
		return err
	}

	// fail reconciler and prevent resource deletion
	// if along the way deleting the remote object fails
	if action == DeleteAction {
		return pe
	}

	return nil
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
	case Local, Empty:
		if src.Data == "" {
			return nil, errors.New("data field required for a 'local' reference")
		}
		return []byte(src.Data), nil

	case ConfigMap:
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

func (r *ObjectReconciler) deleteExternalResources(ctx context.Context, obj *cloudobject.Object) error {
	return r.process(ctx, obj, DeleteAction)
}
