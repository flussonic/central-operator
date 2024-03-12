/*
Copyright 2024.

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

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	mediav1 "flussonic.com/central-operator/api/v1"
)

// CentralReconciler reconciles a Central object
type CentralReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=media.flussonic.com,resources=centrals,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=media.flussonic.com,resources=centrals/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=media.flussonic.com,resources=centrals/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Central object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.0/pkg/reconcile
func (r *CentralReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	central := &mediav1.Central{}
	err := r.Client.Get(ctx, req.NamespacedName, central)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	retry, err := r.deployCentral(ctx, central)
	if retry {
		return ctrl.Result{Requeue: true}, nil
	}
	if err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CentralReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mediav1.Central{}).
		Complete(r)
}

func (r *CentralReconciler) deployCentral(ctx context.Context, w *mediav1.Central) (bool, error) {
	labels := map[string]string{
		"app": w.Name,
	}

	svc1 := &corev1.Service{}
	err := r.Client.Get(ctx, types.NamespacedName{Name: w.Name, Namespace: w.Namespace}, svc1)
	if err != nil && errors.IsNotFound(err) {
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      w.Name,
				Namespace: w.Namespace,
			},
			Spec: corev1.ServiceSpec{
				Selector: labels,
				Type:     "ClusterIP",
				Ports: []corev1.ServicePort{{
					Name:       w.Name,
					Port:       80,
					TargetPort: intstr.FromInt(9019),
				}},
			},
		}
		_ = ctrl.SetControllerReference(w, svc, r.Scheme)
		if err = r.Client.Create(ctx, svc); err != nil {
			return false, err
		}
		return true, nil
	} else if err != nil {
		return false, err
	}

	replicas := int32(1)

	envs := []corev1.EnvVar{
		{
			Name:  "CENTRAL_API_KEY",
			Value: w.Spec.APIKey,
		},
		{
			Name:  "CENTRAL_DATABASE_URL",
			Value: w.Spec.Database,
		},
		{
			Name:  "CENTRAL_API_URL",
			Value: w.Spec.APIURL,
		},
		{
			Name:  "CENTRAL_LOG_LEVEL",
			Value: "debug",
		},
	}

	deployment := &appsv1.Deployment{}

	err = r.Client.Get(ctx, types.NamespacedName{Name: w.Name, Namespace: w.Namespace}, deployment)
	if err != nil && errors.IsNotFound(err) {
		spec := corev1.PodSpec{
			NodeSelector: w.Spec.WebNodeSelector,
			Containers: []corev1.Container{{
				Name:            w.Name,
				Image:           w.Spec.Image,
				ImagePullPolicy: "IfNotPresent",
				Env:             envs,
			}},
		}

		deploy := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      w.Name,
				Namespace: w.Namespace,
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: labels,
				},
				Replicas: &replicas,
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: labels,
					},
					Spec: spec,
				},
			},
		}
		_ = ctrl.SetControllerReference(w, deploy, r.Scheme)
		err = r.Client.Create(ctx, deploy)
		if err != nil {
			return false, err
		}
		return true, nil
	} else if err != nil {
		return false, err
	}

	deployment.Spec.Template.Spec.NodeSelector = w.Spec.WebNodeSelector
	deployment.Spec.Template.Spec.Containers[0].Image = w.Spec.Image
	deployment.Spec.Template.Spec.Containers[0].Env = envs
	deployment.Spec.Replicas = &replicas

	if err = r.Client.Update(ctx, deployment); err != nil {
		return false, err
	}

	return false, nil
}
