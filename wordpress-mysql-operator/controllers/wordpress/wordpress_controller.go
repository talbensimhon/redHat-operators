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

package wordpress

import (
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"context"

	"github.com/go-logr/logr"
	"github.com/prometheus/common/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	wordpressv1alpha1 "github.com/example-inc/wordpress-mysql-operator/apis/wordpress/v1alpha1"
)

// WordpressReconciler reconciles a Wordpress object
type WordpressReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=wordpress.example.com,resources=wordpresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=wordpress.example.com,resources=wordpresses/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=wordpress.example.com,resources=wordpresses/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Wordpress object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *WordpressReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("wordpress", req.NamespacedName)

	// Fetch the wordpress instance
	wordpress := &wordpressv1alpha1.Wordpress{}
	err := r.Get(ctx, req.NamespacedName, wordpress)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Wordpress resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Wordpress")
		return ctrl.Result{}, err
	}

	// Check if the deployment already exists, if not create a new one
	found := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: wordpress.Name, Namespace: wordpress.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.deploymentForWordpress(wordpress)
		log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Create(ctx, dep)
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		log.Info("Deployment created successfully", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)

		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	// Ensure the sqlPassword same as the spec
	password := wordpress.Spec.Password
	log.Debug("managed to read pass", password)

	// Update the Wordpress status with the pod names
	// List the pods for this wordpress's deployment
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(wordpress.Namespace),
		client.MatchingLabels(labelsForWordpress()),
	}
	if err = r.List(ctx, podList, listOpts...); err != nil {
		log.Error(err, "Failed to list pods", "Wordpress.Namespace", wordpress.Namespace, "Wordpress.Name", wordpress.Name)
		return ctrl.Result{}, err
	}
	podNames := getPodNames(podList.Items)

	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, wordpress.Status.Nodes) {
		wordpress.Status.Nodes = podNames
		err := r.Status().Update(ctx, wordpress)
		if err != nil {
			log.Error(err, "Failed to update Wordpress status")
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *WordpressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&wordpressv1alpha1.Wordpress{}).
		Complete(r)
}

// labelsForWordpress returns the labels for selecting the resources
// belonging to the given wordpress CR name.
func labelsForWordpress() map[string]string {
	return map[string]string{"app": "wordpress", "tier": "frontend"}
}

// deploymentForWordpress returns a wordpress Deployment object
func (r *WordpressReconciler) deploymentForWordpress(m *wordpressv1alpha1.Wordpress) *appsv1.Deployment {
	ls := labelsForWordpress()
	password := m.Spec.Password

	dep := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
			Labels:    ls,
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: "wordpress:4.8-apache",
						Name:  m.Name,
						Env: []corev1.EnvVar{{
							Name:  "WORDPRESS_DB_HOST",
							Value: "wordpress - mysql",
						},
							{Name: "WORDPRESS_DB_PASSWORD",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: "sqlRootPassword",
										},
										Key: password,
									},
								},
							}}, //env
						Ports: []corev1.ContainerPort{{
							HostPort: 80,
							Name:     m.Name,
						}}, //ports
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "wordpress-persistent-storage",
							MountPath: "/var/www/html",
						}}, //volumeMount

					}}, //containers
					Volumes: []corev1.Volume{{
						Name: "wordpress-persistent-storage",
						VolumeSource: corev1.VolumeSource{
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: "wp-pv-claim",
							},
						},
					}},
				}, //pod`spec
			}, //pod Template

		}, //Deploy spec

	}

	ctrl.SetControllerReference(m, dep, r.Scheme)
	return dep
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}
