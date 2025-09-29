/*
Copyright 2025.

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

package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	nginxv1 "github.com/Tivoro/go-kubernetes-operator/api/v1"
)

// NginxReconciler reconciles a Nginx object
type NginxReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=nginx.testing,resources=nginxes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=nginx.testing,resources=nginxes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=nginx.testing,resources=nginxes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Nginx object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *NginxReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := logf.FromContext(ctx)
	l.WithValues("Nginx", req.NamespacedName)

	var nginx nginxv1.Nginx
	err := r.Get(ctx, req.NamespacedName, &nginx)
	if err != nil {
		l.Error(err, "Failed to get nginx spec")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// New pod
	pod := &corev1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Name:      nginx.Name + "-pod",
			Namespace: req.Namespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "nginx",
					Image: nginx.Spec.Image,
				},
			},
		},
	}

	found := &corev1.Pod{}
	err = r.Get(ctx, req.NamespacedName, found)
	if err != nil && errors.IsNotFound(err) {
		l.Info("Creating new nginx pod %s/%s", pod.Namespace, pod.Name)
		err := r.Create(ctx, pod)
		if err != nil {
			l.Error(err, "Failed to create pod")
			return reconcile.Result{}, err
		}
	} else if err != nil {
		l.Error(err, "Pod already exists")
		return reconcile.Result{}, err
	}


	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NginxReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&nginxv1.Nginx{}).
		Named("nginx").
		Complete(r)
}
