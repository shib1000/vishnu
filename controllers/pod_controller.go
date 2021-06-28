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

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PodReconciler reconciles a Pod object
type PodReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

const (
	addPodNameLabelAnnotation = "vishnu.k8s/add-pod-name-label"
	podNameLabel              = "vishnu.k8s/pod-name"
)

// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;update;patch;

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Pod object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *PodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("pod", req.NamespacedName)

	log := r.Log.WithValues("pod", req.NamespacedName)

	/*
	   Step 0: Fetch the Pod from the Kubernetes API.
	*/

	var pod corev1.Pod
	if err := r.Get(ctx, req.NamespacedName, &pod); err != nil {
		if apierrors.IsNotFound(err) {
			// Handling the case of pod delete
			return ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch Pod")
		return ctrl.Result{}, err
	}

	labelShouldBePresent := pod.Annotations[addPodNameLabelAnnotation] == "true"
	labelIsPresent := pod.Labels[podNameLabel] == pod.Name

	if labelShouldBePresent == labelIsPresent {
		// The desired state and actual state of the Pod are the same.
		// No further action is required by the operator at this moment.
		log.Info("no update required")
		return ctrl.Result{}, nil
	} else if labelShouldBePresent {
		if pod.Labels == nil {
			pod.Labels = make(map[string]string)
		}
		pod.Labels[podNameLabel] = pod.Name
		log.Info("Added podname to label")
	} else {
		delete(pod.Labels, pod.Name)
	}

	err := r.Update(ctx, &pod)
	if err != nil {
		if apierrors.IsConflict(err) {
			return ctrl.Result{Requeue: true}, err
		}
		if apierrors.IsNotFound(err) {
			return ctrl.Result{Requeue: true}, err
		}

		log.Error(err, "unable to update Pod")

		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// Uncomment the following line adding a pointer to an instance of the controlled resource as an argument
		For(&corev1.Pod{}).
		Complete(r)
}
