//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 deployment_install_client.go InstallStrategyDeploymentInterface
package wrappers

import (
	"context"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	depstatus "github.com/operator-framework/operator-lifecycle-manager/pkg/controller/install/status"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/lib/operatorclient"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/lib/operatorlister"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/lib/ownerutil"
)

var ErrNilObject = errors.New("Bad object supplied: <nil>")

type InstallStrategyDeploymentInterface interface {
	CreateRole(role *rbacv1.Role) (*rbacv1.Role, error)
	CreateRoleBinding(roleBinding *rbacv1.RoleBinding) (*rbacv1.RoleBinding, error)
	EnsureServiceAccount(serviceAccount *corev1.ServiceAccount, owner ownerutil.Owner) (*corev1.ServiceAccount, error)
	CreateDeployment(deployment *appsv1.Deployment) (*appsv1.Deployment, error)
	CreateOrUpdateDeployment(deployment *appsv1.Deployment) (*appsv1.Deployment, error)
	IsDeploymentReady(namespace, name string) (bool, error)
	DeleteDeployment(name string) error
	GetServiceAccountByName(serviceAccountName string) (*corev1.ServiceAccount, error)
	FindAnyDeploymentsMatchingNames(depNames []string) ([]*appsv1.Deployment, error)
	FindAnyDeploymentsMatchingLabels(label labels.Selector) ([]*appsv1.Deployment, error)
	GetOpClient() operatorclient.ClientInterface
	GetOpLister() operatorlister.OperatorLister
}

type InstallStrategyDeploymentClientForNamespace struct {
	opClient  operatorclient.ClientInterface
	opLister  operatorlister.OperatorLister
	Namespace string
}

var _ InstallStrategyDeploymentInterface = &InstallStrategyDeploymentClientForNamespace{}

func NewInstallStrategyDeploymentClient(opClient operatorclient.ClientInterface, opLister operatorlister.OperatorLister, namespace string) InstallStrategyDeploymentInterface {
	return &InstallStrategyDeploymentClientForNamespace{
		opClient:  opClient,
		opLister:  opLister,
		Namespace: namespace,
	}
}

func (c *InstallStrategyDeploymentClientForNamespace) GetOpClient() operatorclient.ClientInterface {
	return c.opClient
}

func (c *InstallStrategyDeploymentClientForNamespace) GetOpLister() operatorlister.OperatorLister {
	return c.opLister
}

func (c *InstallStrategyDeploymentClientForNamespace) CreateRole(role *rbacv1.Role) (*rbacv1.Role, error) {
	return c.opClient.KubernetesInterface().RbacV1().Roles(c.Namespace).Create(context.TODO(), role, metav1.CreateOptions{})
}

func (c *InstallStrategyDeploymentClientForNamespace) CreateRoleBinding(roleBinding *rbacv1.RoleBinding) (*rbacv1.RoleBinding, error) {
	return c.opClient.KubernetesInterface().RbacV1().RoleBindings(c.Namespace).Create(context.TODO(), roleBinding, metav1.CreateOptions{})
}

func (c *InstallStrategyDeploymentClientForNamespace) EnsureServiceAccount(serviceAccount *corev1.ServiceAccount, owner ownerutil.Owner) (*corev1.ServiceAccount, error) {
	if serviceAccount == nil {
		return nil, ErrNilObject
	}

	foundAccount, err := c.opLister.CoreV1().ServiceAccountLister().ServiceAccounts(c.Namespace).Get(serviceAccount.Name)
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, errors.Wrap(err, "checking for existing serviceacccount failed")
	}

	// create if not found
	if err != nil && apierrors.IsNotFound(err) {
		serviceAccount.SetNamespace(c.Namespace)
		createdAccount, err := c.opClient.CreateServiceAccount(serviceAccount)
		if err != nil && !apierrors.IsAlreadyExists(err) {
			return nil, errors.Wrap(err, "creating serviceacccount failed")
		}
		if apierrors.IsAlreadyExists(err) {
			return serviceAccount, nil
		}
		return createdAccount, nil
	}

	// if found, ensure ownerreferences
	if ownerutil.IsOwnedBy(foundAccount, owner) {
		return foundAccount, nil
	}
	// set owner if missing
	ownerutil.AddNonBlockingOwner(foundAccount, owner)
	return c.opClient.UpdateServiceAccount(foundAccount)
}

func (c *InstallStrategyDeploymentClientForNamespace) CreateDeployment(deployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	return c.opClient.CreateDeployment(deployment)
}

func (c *InstallStrategyDeploymentClientForNamespace) IsDeploymentReady(namespace, name string) (bool, error) {
	deployment, err := c.opClient.GetDeployment(namespace, name)
	if err != nil {
		return false, err
	}
	_, ready, err := depstatus.DeploymentStatus(deployment)
	return ready, err
}

func (c *InstallStrategyDeploymentClientForNamespace) DeleteDeployment(name string) error {
	foregroundDelete := metav1.DeletePropagationForeground // cascading delete
	immediate := int64(0)
	immediateForegroundDelete := &metav1.DeleteOptions{GracePeriodSeconds: &immediate, PropagationPolicy: &foregroundDelete}
	return c.opClient.DeleteDeployment(c.Namespace, name, immediateForegroundDelete)
}

func (c *InstallStrategyDeploymentClientForNamespace) CreateOrUpdateDeployment(deployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	_, err := c.opClient.GetDeployment(deployment.Namespace, deployment.Name)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, err
		}
		created, err := c.CreateDeployment(deployment)
		if err != nil {
			return nil, err
		}
		return created, err
	}
	return c.opClient.KubernetesInterface().AppsV1().Deployments(deployment.GetNamespace()).Update(context.TODO(), deployment, metav1.UpdateOptions{})
}

func (c *InstallStrategyDeploymentClientForNamespace) GetServiceAccountByName(serviceAccountName string) (*corev1.ServiceAccount, error) {
	return c.opLister.CoreV1().ServiceAccountLister().ServiceAccounts(c.Namespace).Get(serviceAccountName)
}

func (c *InstallStrategyDeploymentClientForNamespace) FindAnyDeploymentsMatchingNames(depNames []string) ([]*appsv1.Deployment, error) {
	var deployments []*appsv1.Deployment
	for _, depName := range depNames {
		fetchedDep, err := c.opLister.AppsV1().DeploymentLister().Deployments(c.Namespace).Get(depName)
		if err == nil {
			deployments = append(deployments, fetchedDep)
		} else {
			// Any errors other than !exists are propagated up
			if !apierrors.IsNotFound(err) {
				return deployments, err
			}
		}
	}
	return deployments, nil
}

func (c *InstallStrategyDeploymentClientForNamespace) FindAnyDeploymentsMatchingLabels(label labels.Selector) ([]*appsv1.Deployment, error) {
	deployments, err := c.opLister.AppsV1().DeploymentLister().Deployments(c.Namespace).List(label)
	if err != nil {
		return nil, err
	}

	return deployments, nil
}
