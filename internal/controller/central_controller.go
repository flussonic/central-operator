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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	mediav1alpha1 "flussonic.com/central/operator/api/v1alpha1"
)

// CentralReconciler reconciles a Central object
type CentralReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

type CentralStreamers struct {
	Streamers []CentralStreamer `json:"streamers"`
}

type CentralStreamer struct {
	Hostname          string `json:"hostname"`
	APIUrl            string `json:"api_url,omitempty"`
	PrivatePayloadUrl string `json:"private_payload_url,omitempty"`
	PublicPayloadUrl  string `json:"public_payload_url,omitempty"`
	ClusterKey        string `json:"cluster_key,omitempty"`
}

// +kubebuilder:rbac:groups=media.flussonic.com,resources=centrals,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=media.flussonic.com,resources=centrals/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=media.flussonic.com,resources=centrals/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Central object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *CentralReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)

	log.Info("central reconciler: reconcile started...")

	spec := &mediav1alpha1.Central{}
	if err := r.Client.Get(ctx, req.NamespacedName, spec); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "central reconciler get central error")
		return ctrl.Result{}, err
	}

	result, err := r.deployCentral(ctx, spec)
	if err != nil || result.Requeue {
		return result, err
	}

	result, err = r.provisionStreamers(ctx, spec)
	if err != nil || result.Requeue {
		if err != nil {
			log.Error(err, "central reconciler provision streamers error")
		}
		return result, err
	}
	// FIXME: что делать, попав сюда? Ведь тут смысл оператора теряется, поды больше не будут опрашиваться
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CentralReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mediav1alpha1.Central{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}

func CreateCoreEnvs(s *mediav1alpha1.Central) []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name:  "CENTRAL_API_KEY",
			Value: s.Spec.APIKey,
		},
		{
			Name:  "CENTRAL_DATABASE_URL",
			Value: s.Spec.Database,
		},
		{
			Name:  "CENTRAL_CLUSTER_NODE_CONFIG_PROVISION_ENABLED",
			Value: "false",
		},
		{
			Name:  "CENTRAL_API_URL",
			Value: s.Spec.APIURL,
		},
		{
			Name:  "CENTRAL_LOG_REQUESTS",
			Value: strconv.FormatBool(s.Spec.LogRequests),
		},
		{
			Name:  "CENTRAL_LOG_LEVEL",
			Value: s.Spec.LogLevel,
		},
		{
			Name: "CENTRAL_HTTP_PORT",
			// FIXME: это поле не нужно пробрасывать из настроек. Достаточно выставить порт 80 и успокоиться
			Value: strconv.FormatInt(80, 10),
		},
		{
			Name:  "CENTRAL_DYNAMIC_STREAMS_AUTH_TOKEN",
			Value: s.Spec.DynamicStreamsAuthToken,
		},
		{
			Name:  "CENTRAL_EDIT_AUTH",
			Value: s.Spec.EditAuth,
		},
	}
}

func CreateProvisionerEnvs(s *mediav1alpha1.Central) []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name:  "CLUSTER_KEY",
			Value: s.Spec.ProvisionerClusterKey,
		},
		{
			Name:  "API_KEY",
			Value: s.Spec.APIKey,
		},
		{
			Name:  "SELECTOR",
			Value: s.Spec.ProvisionerSelector,
		},
		{
			Name:  "CENTRAL_API",
			Value: s.Spec.APIURL,
		},
	}
}

func (r *CentralReconciler) deployCentral(ctx context.Context, w *mediav1alpha1.Central) (ctrl.Result, error) {
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
				Type:     corev1.ServiceTypeClusterIP,
				Ports: []corev1.ServicePort{{
					Name:       w.Name,
					Port:       80,
					TargetPort: intstr.FromInt(80),
				}},
			},
		}
		_ = ctrl.SetControllerReference(w, svc, r.Scheme)
		if err = r.Client.Create(ctx, svc); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	}

	replicas := int32(1)

	envs := CreateCoreEnvs(w)

	deployment := &appsv1.Deployment{}

	err = r.Client.Get(ctx, types.NamespacedName{Name: w.Name, Namespace: w.Namespace}, deployment)
	if err != nil && errors.IsNotFound(err) {
		spec := corev1.PodSpec{
			NodeSelector: w.Spec.NodeSelector,
			Containers: []corev1.Container{{
				Name:            w.Name,
				Image:           w.Spec.Image,
				ImagePullPolicy: corev1.PullIfNotPresent,
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
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	}

	deployment.Spec.Template.Spec.NodeSelector = w.Spec.NodeSelector
	deployment.Spec.Template.Spec.Containers[0].Image = w.Spec.Image
	deployment.Spec.Template.Spec.Containers[0].Env = envs
	deployment.Spec.Replicas = &replicas

	if err = r.Client.Update(ctx, deployment); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *CentralReconciler) provisionStreamers(
	ctx context.Context,
	s *mediav1alpha1.Central,
) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)

	pods := &corev1.PodList{}
	if err := r.Client.List(ctx, pods, &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(labels.Set{
			"app": s.Spec.ProvisionerSelector,
		}),
		Namespace: s.Namespace,
	}); err != nil && !errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	existingStreamers := make(map[string]bool, len(pods.Items))
	for _, pod := range pods.Items {
		// FIXME: Вот это всё надо делать настраиваемым снаружи, не хардкодом
		streamer := CentralStreamer{
			Hostname:          pod.Spec.NodeName,
			APIUrl:            "http://" + pod.Status.PodIP + ":81",
			PrivatePayloadUrl: "http://" + pod.Status.PodIP + ":81",
			PublicPayloadUrl:  "http://" + pod.Status.HostIP,
			ClusterKey:        s.Spec.ProvisionerClusterKey,
		}

		payload, err := json.Marshal(streamer)
		if err != nil {
			fmt.Println("marshall err", err.Error())
			continue
		}

		// FIXME: APIURL и APIKey надо сделать автоматически создаваемыми этим оператором
		request, err := http.NewRequest(
			http.MethodPut,
			s.Spec.APIURL+"/central/api/v3/streamers/"+streamer.Hostname,
			bytes.NewBuffer(payload),
		)
		if err != nil {
			fmt.Println("new rq err", err.Error())
			continue
		}

		request.Header.Add("Content-Type", "application/json")
		request.Header.Add("Authorization", "Bearer "+s.Spec.APIKey)

		client := http.Client{}

		response, err := client.Do(request)
		if err != nil {
			fmt.Println("do err", err.Error())
			continue
		}
		_ = response.Body.Close()

		existingStreamers[streamer.Hostname] = true
	}

	centralStreamers, err := LoadCentralStreamers(
		s.Spec.APIURL+"/central/api/v3",
		s.Spec.APIKey,
	)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("load central streamers failed:%s", err.Error())
	}

	if centralStreamers == nil || centralStreamers.Streamers == nil {
		return ctrl.Result{}, nil
	}

	for _, streamer := range centralStreamers.Streamers {
		if ok := existingStreamers[streamer.Hostname]; !ok {
			if err := DeleteCentralStreamer(
				s.Spec.APIURL+"/central/api/v3",
				s.Spec.APIKey,
				streamer.Hostname,
			); err != nil {
				log.Error(err, "failed to delete streamer")
				continue
			}
		}
	}

	return ctrl.Result{Requeue: true, RequeueAfter: time.Second * 5}, nil
}

func LoadCentralStreamers(
	centralAPIURL string,
	centralAPIKey string,
) (*CentralStreamers, error) {
	var centralStreamers CentralStreamers

	request, err := http.NewRequest(
		http.MethodGet,
		centralAPIURL+"/streamers",
		http.NoBody,
	)

	if err != nil {
		return nil, err
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Authorization", "Bearer "+centralAPIKey)

	client := http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = response.Body.Close()
	}()

	buf, _ := io.ReadAll(response.Body)
	if err = json.Unmarshal(buf, &centralStreamers); err != nil {
		return nil, fmt.Errorf("error:%s, body:%s", err.Error(), string(buf))
	}

	if err = json.Unmarshal(buf, &centralStreamers); err != nil {
		return nil, fmt.Errorf("error:%s, body:%s", err.Error(), string(buf))
	}
	return &centralStreamers, nil
}

func DeleteCentralStreamer(
	centralAPIURL string,
	centralAPIKey string,
	streamerHostname string,
) error {
	request, err := http.NewRequest(
		http.MethodDelete,
		centralAPIURL+"/streamers/"+streamerHostname,
		http.NoBody,
	)

	if err != nil {
		return err
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Authorization", "Bearer "+centralAPIKey)

	client := http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return err
	}

	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode == http.StatusNoContent {
		return nil
	}
	return fmt.Errorf("status code is not 204, actual:%d", response.StatusCode)
}
