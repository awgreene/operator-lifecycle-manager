package operators

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	operatorsv1 "github.com/operator-framework/api/pkg/operators/v1"
	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/lib/ownerutil"
)

// OperatorReconciler reconciles a Operator object.
type OperatorConditionReconciler struct {
	client.Client
	log logr.Logger
}

// +kubebuilder:rbac:groups=operators.coreos.com,resources=operatorconditions,verbs=get;list;update;patch;delete
// +kubebuilder:rbac:groups=operators.coreos.com,resources=operatorconditions/status,verbs=update;patch

// SetupWithManager adds the OperatorCondition Reconciler reconciler to the given controller manager.
func (r *OperatorConditionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorsv1.OperatorCondition{}).
		Complete(r)
}

// NewOperatorReconciler constructs and returns an OperatorReconciler.
// As a side effect, the given scheme has operator discovery types added to it.
func NewOperatorConditionReconciler(cli client.Client, log logr.Logger, scheme *runtime.Scheme) (*OperatorConditionReconciler, error) {
	// Add watched types to scheme.
	if err := AddToScheme(scheme); err != nil {
		return nil, err
	}

	return &OperatorConditionReconciler{
		Client: cli,
		log:    log,
	}, nil
}

const (
	OperatorConditionEnvVarKey = "OPERATOR_CONDITION_NAME"
)

// Implement reconcile.Reconciler so the controller can reconcile objects
var _ reconcile.Reconciler = &OperatorConditionReconciler{}

func (r *OperatorConditionReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	// Set up a convenient log object so we don't have to type request over and over again
	log := r.log.WithValues("request", req)
	log.V(1).Info("reconciling operatorcondition")

	operatorCondition := &operatorsv1.OperatorCondition{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, operatorCondition)
	if err != nil {
		log.V(1).Info("Unable to find operatorcondition", "error", err)
		return ctrl.Result{}, err
	}

	ownerReference := ownerutil.GetOwnerByKind(operatorCondition, operatorsv1alpha1.ClusterServiceVersionKind)
	if ownerReference == nil {
		return ctrl.Result{}, fmt.Errorf("No CSV Owner specified")
	}

	err = r.ensureOperatorConditionRole(*operatorCondition, *ownerReference)
	if err != nil {
		log.V(1).Info("Error reconciling operatorcondition", "error", err)
		return ctrl.Result{Requeue: true}, err
	}

	err = r.ensureDeploymentEnvVars(*operatorCondition, *ownerReference)
	if err != nil {
		log.V(1).Info("Error reconciling operatorcondition", "error", err)
		return ctrl.Result{Requeue: true}, err
	}

	return ctrl.Result{}, nil
}

func (r *OperatorConditionReconciler) ensureOperatorConditionRole(operatorCondition operatorsv1.OperatorCondition, ownerReference metav1.OwnerReference) error {
	r.log.V(1).Info("Creating the RBAC for the operatorCondition", "operatorConditions", operatorCondition)

	// get all Service Accounts
	serviceAccountList := corev1.ServiceAccountList{}
	err := r.Client.List(context.TODO(), &serviceAccountList, client.InNamespace(operatorCondition.GetNamespace()))
	if err != nil {
		return err
	}

	subjects := []rbacv1.Subject{}
	for _, serviceAccount := range filterServiceAccounts(serviceAccountList.Items, hasOwnerFunc(ownerReference)) {
		subjects = append(subjects, rbacv1.Subject{
			Kind:     rbacv1.ServiceAccountKind,
			Name:     serviceAccount.GetName(),
			APIGroup: "",
		})
	}

	role := rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:            operatorCondition.GetName(),
			Namespace:       operatorCondition.GetNamespace(),
			OwnerReferences: []metav1.OwnerReference{ownerReference},
		},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:         []string{"get"},
				APIGroups:     []string{"operators.coreos.com"},
				Resources:     []string{"operatorconditions"},
				ResourceNames: []string{operatorCondition.GetName()},
			},
			{
				Verbs:         []string{"get,update,patch"},
				APIGroups:     []string{"operators.coreos.com"},
				Resources:     []string{"operatorconditions/status"},
				ResourceNames: []string{operatorCondition.GetName()},
			},
		},
	}
	err = r.Client.Create(context.TODO(), &role)
	if err != nil {
		if !k8serrors.IsAlreadyExists(err) {

			return err
		}
		existingRole := rbacv1.Role{}
		err := r.Client.Get(context.TODO(), client.ObjectKey{Name: role.GetName(), Namespace: role.GetNamespace()}, &existingRole)
		if err != nil {
			r.log.V(1).Info("Alex: error getting role")
			return err
		}
		existingRole.OwnerReferences = role.OwnerReferences
		existingRole.Rules = role.Rules
		err = r.Client.Update(context.TODO(), &existingRole)
		r.log.V(1).Info("Alex: error updating role")
		if err != nil {
			r.log.V(1).Info("Alex: error updating rolebinding 3", "error", err)
			return err
		}
	}

	roleBinding := rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:            operatorCondition.GetName(),
			Namespace:       operatorCondition.GetNamespace(),
			OwnerReferences: []metav1.OwnerReference{ownerReference},
		},
		Subjects: subjects,
		RoleRef: rbacv1.RoleRef{
			Kind:     "Role",
			Name:     role.GetName(),
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
	err = r.Client.Create(context.TODO(), &roleBinding)
	if err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return err
		}
		existingRoleBinding := rbacv1.RoleBinding{}
		err := r.Client.Get(context.TODO(), client.ObjectKey{Name: roleBinding.GetName(), Namespace: roleBinding.GetNamespace()}, &existingRoleBinding)
		if err != nil {
			return err
		}
		existingRoleBinding.OwnerReferences = roleBinding.OwnerReferences
		existingRoleBinding.Subjects = roleBinding.Subjects
		existingRoleBinding.RoleRef = roleBinding.RoleRef
		err = r.Client.Update(context.TODO(), &existingRoleBinding)
		if err != nil {
			return err
		}
	}

	return nil
}

func hasOwnerFunc(ownerReference metav1.OwnerReference) func(object metav1.Object) bool {
	return func(object metav1.Object) bool {
		if ownerutil.HasOwnerReference(object, ownerReference) {
			return true
		}
		return false
	}
}

func filterServiceAccounts(serviceAccounts []corev1.ServiceAccount, f func(object metav1.Object) bool) []corev1.ServiceAccount {
	// // filter based on OwnerReferences
	filteredServiceAccounts := []corev1.ServiceAccount{}
	for _, serviceAccount := range serviceAccounts {
		if f(&serviceAccount) {
			filteredServiceAccounts = append(filteredServiceAccounts, serviceAccount)
		}
	}
	return filteredServiceAccounts
}

func (r *OperatorConditionReconciler) ensureDeploymentEnvVars(operatorCondition operatorsv1.OperatorCondition, ownerReference metav1.OwnerReference) error {
	// get all deployments in the namespace
	deploymentList := appsv1.DeploymentList{}
	err := r.Client.List(context.TODO(), &deploymentList, client.InNamespace(operatorCondition.GetNamespace()))
	if err != nil {
		return err
	}

	// filter based on OwnerReferences
	for _, deployment := range filterDeployments(deploymentList.Items, hasOwnerFunc(ownerReference)) {
		for i := range deployment.Spec.Template.Spec.Containers {
			deployment.Spec.Template.Spec.Containers[i].Env = ensureEnvVar(corev1.EnvVar{Name: OperatorConditionEnvVarKey, Value: operatorCondition.GetName()}, deployment.Spec.Template.Spec.Containers[i].Env)
		}
		err = r.Client.Update(context.TODO(), &deployment)
		if err != nil {
			return err
		}
	}
	return nil
}

func filterDeployments(deployments []appsv1.Deployment, f func(object metav1.Object) bool) []appsv1.Deployment {
	// // filter based on OwnerReferences
	filteredDeployments := []appsv1.Deployment{}
	for _, deployment := range deployments {
		if f(&deployment) {
			filteredDeployments = append(filteredDeployments, deployment)
		}
	}
	return filteredDeployments
}

func ensureEnvVar(envVar corev1.EnvVar, envVars []corev1.EnvVar) []corev1.EnvVar {
	if len(envVars) == 0 {
		return []corev1.EnvVar{envVar}
	}

	result := []corev1.EnvVar{}
	found := false
	for i := range envVars {
		if envVars[i].Name != envVar.Name {
			result = append(result, envVars[i])
		} else {
			found = true
			result = append(result, envVar)
		}
	}

	if !found {
		result = append(result, envVar)
	}

	return result
}
