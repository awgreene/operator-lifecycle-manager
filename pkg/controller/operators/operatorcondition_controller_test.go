package operators

import (
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFilterServiceAccounts(t *testing.T) {
	ownerReference := metav1.OwnerReference{
		Name: "foo",
		Kind: "bar",
		UID:  "uid",
	}
	f := hasOwnerFunc(ownerReference)
	tests := []struct {
		description     string
		serviceAccounts []corev1.ServiceAccount
		expectedResult  []corev1.ServiceAccount
	}{
		{
			description:     "nil ServiceAccounts list doesn't cause a panic",
			serviceAccounts: nil,
			expectedResult:  []corev1.ServiceAccount{},
		},
		{
			description: "ServiceAccounts list with no owners returns empty array",
			serviceAccounts: []corev1.ServiceAccount{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo",
						Namespace: "bar",
					},
				},
			},
			expectedResult: []corev1.ServiceAccount{},
		},
		{
			description: "ServiceAccounts list with wrong owner name returns empty array",
			serviceAccounts: []corev1.ServiceAccount{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo",
						Namespace: "bar",
						OwnerReferences: []metav1.OwnerReference{
							{
								Name: "notFoo",
								Kind: "bar",
								UID:  "uid",
							},
						},
					},
				},
			},
			expectedResult: []corev1.ServiceAccount{},
		},
		{
			description: "ServiceAccounts list with wrong owner kind returns empty array",
			serviceAccounts: []corev1.ServiceAccount{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo",
						Namespace: "bar",
						OwnerReferences: []metav1.OwnerReference{
							{
								Name: "foo",
								Kind: "notBar",
								UID:  "uid",
							},
						},
					},
				},
			},
			expectedResult: []corev1.ServiceAccount{},
		},
		{
			description: "ServiceAccounts list with wrong owner uid returns empty array",
			serviceAccounts: []corev1.ServiceAccount{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo",
						Namespace: "bar",
						OwnerReferences: []metav1.OwnerReference{
							{
								Name: "foo",
								Kind: "bar",
								UID:  "notUID",
							},
						},
					},
				},
			},
			expectedResult: []corev1.ServiceAccount{},
		},
		{
			description: "ServiceAccounts list with correct owner uid returns the same array",
			serviceAccounts: []corev1.ServiceAccount{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo",
						Namespace: "bar",
						OwnerReferences: []metav1.OwnerReference{
							ownerReference,
						},
					},
				},
			},
			expectedResult: []corev1.ServiceAccount{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo",
						Namespace: "bar",
						OwnerReferences: []metav1.OwnerReference{
							ownerReference,
						},
					},
				},
			},
		},
		{
			description: "ServiceAccounts list returns only ServiceAccounts with the correct owner",
			serviceAccounts: []corev1.ServiceAccount{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo",
						Namespace: "bar",
						OwnerReferences: []metav1.OwnerReference{
							ownerReference,
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo2",
						Namespace: "bar",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo3",
						Namespace: "bar",
						OwnerReferences: []metav1.OwnerReference{
							ownerReference,
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo4",
						Namespace: "bar",
						OwnerReferences: []metav1.OwnerReference{
							ownerReference,
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo5",
						Namespace: "bar",
						OwnerReferences: []metav1.OwnerReference{
							ownerReference,
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo6",
						Namespace: "bar",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo7",
						Namespace: "bar",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo8",
						Namespace: "bar",
						OwnerReferences: []metav1.OwnerReference{
							ownerReference,
						},
					},
				},
			},
			expectedResult: []corev1.ServiceAccount{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo",
						Namespace: "bar",
						OwnerReferences: []metav1.OwnerReference{
							ownerReference,
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo3",
						Namespace: "bar",
						OwnerReferences: []metav1.OwnerReference{
							ownerReference,
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo4",
						Namespace: "bar",
						OwnerReferences: []metav1.OwnerReference{
							ownerReference,
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo5",
						Namespace: "bar",
						OwnerReferences: []metav1.OwnerReference{
							ownerReference,
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo8",
						Namespace: "bar",
						OwnerReferences: []metav1.OwnerReference{
							ownerReference,
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			require.Equal(t, tt.expectedResult, filterServiceAccounts(tt.serviceAccounts, f))
		})
	}
}

func TestFilterDeployments(t *testing.T) {
	ownerReference := metav1.OwnerReference{
		Name: "foo",
		Kind: "bar",
		UID:  "uid",
	}
	f := hasOwnerFunc(ownerReference)
	tests := []struct {
		description    string
		deployments    []appsv1.Deployment
		expectedResult []appsv1.Deployment
	}{
		{
			description:    "nil Deployment list doesn't cause a panic",
			deployments:    nil,
			expectedResult: []appsv1.Deployment{},
		},
		{
			description: "Deployments list with no owners returns empty array",
			deployments: []appsv1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo",
						Namespace: "bar",
					},
				},
			},
			expectedResult: []appsv1.Deployment{},
		},
		{
			description: "Deployments list with wrong owner name returns empty array",
			deployments: []appsv1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo",
						Namespace: "bar",
						OwnerReferences: []metav1.OwnerReference{
							{
								Name: "notFoo",
								Kind: "bar",
								UID:  "uid",
							},
						},
					},
				},
			},
			expectedResult: []appsv1.Deployment{},
		},
		{
			description: "ServiceAccounts list with wrong owner kind returns empty array",
			deployments: []appsv1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo",
						Namespace: "bar",
						OwnerReferences: []metav1.OwnerReference{
							{
								Name: "foo",
								Kind: "notBar",
								UID:  "uid",
							},
						},
					},
				},
			},
			expectedResult: []appsv1.Deployment{},
		},
		{
			description: "Deployments list with wrong owner uid returns empty array",
			deployments: []appsv1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo",
						Namespace: "bar",
						OwnerReferences: []metav1.OwnerReference{
							{
								Name: "foo",
								Kind: "bar",
								UID:  "notUID",
							},
						},
					},
				},
			},
			expectedResult: []appsv1.Deployment{},
		},
		{
			description: "Deployments list with correct owner uid returns the same array",
			deployments: []appsv1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo",
						Namespace: "bar",
						OwnerReferences: []metav1.OwnerReference{
							ownerReference,
						},
					},
				},
			},
			expectedResult: []appsv1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo",
						Namespace: "bar",
						OwnerReferences: []metav1.OwnerReference{
							ownerReference,
						},
					},
				},
			},
		},
		{
			description: "Deployments list returns only Deployments with the correct owner",
			deployments: []appsv1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo",
						Namespace: "bar",
						OwnerReferences: []metav1.OwnerReference{
							ownerReference,
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo2",
						Namespace: "bar",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo3",
						Namespace: "bar",
						OwnerReferences: []metav1.OwnerReference{
							ownerReference,
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo4",
						Namespace: "bar",
						OwnerReferences: []metav1.OwnerReference{
							ownerReference,
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo5",
						Namespace: "bar",
						OwnerReferences: []metav1.OwnerReference{
							ownerReference,
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo6",
						Namespace: "bar",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo7",
						Namespace: "bar",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo8",
						Namespace: "bar",
						OwnerReferences: []metav1.OwnerReference{
							ownerReference,
						},
					},
				},
			},
			expectedResult: []appsv1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo",
						Namespace: "bar",
						OwnerReferences: []metav1.OwnerReference{
							ownerReference,
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo3",
						Namespace: "bar",
						OwnerReferences: []metav1.OwnerReference{
							ownerReference,
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo4",
						Namespace: "bar",
						OwnerReferences: []metav1.OwnerReference{
							ownerReference,
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo5",
						Namespace: "bar",
						OwnerReferences: []metav1.OwnerReference{
							ownerReference,
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "foo8",
						Namespace: "bar",
						OwnerReferences: []metav1.OwnerReference{
							ownerReference,
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			require.Equal(t, tt.expectedResult, filterDeployments(tt.deployments, f))
		})
	}
}

func TestEnsureEnvVar(t *testing.T) {

	tests := []struct {
		description    string
		envVar         corev1.EnvVar
		envVars        []corev1.EnvVar
		expectedResult []corev1.EnvVar
	}{
		{
			description:    "Adds envVar to empty EnvVar array",
			envVar:         corev1.EnvVar{Name: "foo", Value: "bar"},
			envVars:        []corev1.EnvVar{},
			expectedResult: []corev1.EnvVar{{Name: "foo", Value: "bar"}},
		},
		{
			description:    "Adds envVar to EnvVar array containing different key",
			envVar:         corev1.EnvVar{Name: "foo", Value: "bar"},
			envVars:        []corev1.EnvVar{{Name: "notFoo", Value: "bar"}},
			expectedResult: []corev1.EnvVar{{Name: "notFoo", Value: "bar"}, {Name: "foo", Value: "bar"}},
		},
		{
			description:    "Adds envVar to a large EnvVar array containing different keys",
			envVar:         corev1.EnvVar{Name: "foo", Value: "bar"},
			envVars:        []corev1.EnvVar{{Name: "foo1", Value: "bar"}, {Name: "foo2", Value: "bar"}, {Name: "foo3", Value: "bar"}},
			expectedResult: []corev1.EnvVar{{Name: "foo1", Value: "bar"}, {Name: "foo2", Value: "bar"}, {Name: "foo3", Value: "bar"}, {Name: "foo", Value: "bar"}},
		},
		{
			description:    "Updates existing envVars with same Name",
			envVar:         corev1.EnvVar{Name: "foo", Value: "bar"},
			envVars:        []corev1.EnvVar{{Name: "foo", Value: "notBar"}},
			expectedResult: []corev1.EnvVar{{Name: "foo", Value: "bar"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			require.Equal(t, tt.expectedResult, ensureEnvVar(tt.envVar, tt.envVars))
		})
	}
}
