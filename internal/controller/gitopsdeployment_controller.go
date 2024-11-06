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
	"github.com/dolevkle/gitops-deploy-operator/pkg/git"
	"github.com/dolevkle/gitops-deploy-operator/pkg/k8s"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	gitopsv1alpha1 "github.com/dolevkle/gitops-deploy-operator/api/v1alpha1"
)

// GitOpsDeploymentReconciler reconciles a GitOpsDeployment object
type GitOpsDeploymentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=gitops.example.com,resources=gitopsdeployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=gitops.example.com,resources=gitopsdeployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=gitops.example.com,resources=gitopsdeployments/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the GitOpsDeployment object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *GitOpsDeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the GitOpsDeployment instance
	var deployment gitopsv1alpha1.GitOpsDeployment
	if err := r.Get(ctx, req.NamespacedName, &deployment); err != nil {
		logger.Error(err, "Failed to get GitOpsDeployment")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Check if the object is being deleted
	//if !deployment.ObjectMeta.DeletionTimestamp.IsZero() {
	//	if utils.ContainsString(deployment.GetFinalizers(), gitOpsFinalizer) {
	//		// Run finalization logic
	//		if err := r.finalizeDeployment(ctx, &deployment, logger); err != nil {
	//			return ctrl.Result{}, err
	//		}
	//		// Remove finalizer
	//		deployment.SetFinalizers(utils.RemoveString(deployment.GetFinalizers(), gitOpsFinalizer))
	//		if err := r.Update(ctx, &deployment); err != nil {
	//			logger.Error(err, "Failed to remove finalizer")
	//			return ctrl.Result{}, err
	//		}
	//	}
	//	return ctrl.Result{}, nil
	//}

	// Add finalizer if not present
	//if !utils.ContainsString(deployment.GetFinalizers(), gitOpsFinalizer) {
	//	deployment.SetFinalizers(append(deployment.GetFinalizers(), gitOpsFinalizer))
	//	if err := r.Update(ctx, &deployment); err != nil {
	//		logger.Error(err, "Failed to add finalizer")
	//		return ctrl.Result{}, err
	//	}
	//}

	// Clone or pull the Git repository
	repoManager := git.NewRepoManager(deployment.Spec.RepoURL, deployment.Spec.Branch, deployment.Name)
	if err := repoManager.CloneOrPull(); err != nil {
		logger.Error(err, "Failed to clone or pull repository")
		r.updateStatus(ctx, &deployment, metav1.ConditionFalse, "CloneFailed", err.Error())
		return ctrl.Result{}, err
	}

	// Apply manifests
	k8sManager := k8s.NewK8sManager(r.Client, r.Scheme)
	manifestsPath := repoManager.GetManifestsPath(deployment.Spec.Path)
	if err := k8sManager.ApplyManifests(ctx, manifestsPath, logger); err != nil {
		logger.Error(err, "Failed to apply manifests")
		r.updateStatus(ctx, &deployment, metav1.ConditionFalse, "ApplyFailed", err.Error())
		return ctrl.Result{}, err
	}

	// Update status
	deployment.Status.Synced = true
	deployment.Status.LastSyncTime = metav1.Now()
	r.updateStatus(ctx, &deployment, metav1.ConditionTrue, "Reconciled", "Successfully applied manifests")
	if err := r.Status().Update(ctx, &deployment); err != nil {
		logger.Error(err, "Failed to update status")
		return ctrl.Result{}, err
	}

	// Schedule next reconciliation
	interval, err := time.ParseDuration(deployment.Spec.Interval)
	if err != nil {
		logger.Error(err, "Invalid interval format")
		return ctrl.Result{}, err
	}
	return ctrl.Result{RequeueAfter: interval}, nil
}

func (r *GitOpsDeploymentReconciler) finalizeDeployment(ctx context.Context, deployment *gitopsv1alpha1.GitOpsDeployment, logger logr.Logger) error {
	// Cleanup resources
	k8sManager := k8s.NewK8sManager(r.Client, r.Scheme)
	manifestsPath := git.GetRepoPath(deployment.Name, deployment.Spec.Path)
	if err := k8sManager.DeleteManifests(ctx, manifestsPath, logger); err != nil {
		logger.Error(err, "Failed to delete manifests during finalization")
		return err
	}
	// Optionally, remove cloned repository
	if err := git.DeleteRepo(deployment.Name); err != nil {
		logger.Error(err, "Failed to delete repository during finalization")
		return err
	}
	return nil
}

// Updates the status of the GitOpsDeployment resource.
func (r *GitOpsDeploymentReconciler) updateStatus(ctx context.Context, deployment *gitopsv1alpha1.GitOpsDeployment, status metav1.ConditionStatus, reason, message string) {
	condition := metav1.Condition{
		Type:               "Ready",
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: metav1.Now(),
	}
	deployment.Status.Conditions = []metav1.Condition{condition}
}

// SetupWithManager sets up the controller with the Manager.
func (r *GitOpsDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gitopsv1alpha1.GitOpsDeployment{}).
		Named("gitopsdeployment").
		Complete(r)
}
