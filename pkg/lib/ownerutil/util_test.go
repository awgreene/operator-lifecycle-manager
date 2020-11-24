package ownerutil

import (
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIsOwnedBy(t *testing.T) {
	return
}

func TestHasOwnerReference(t *testing.T) {
	ownerReference := metav1.OwnerReference{
		Name: "foo",
		Kind: "bar",
		UID:  "uid",
	}
	tests := []struct {
		description    string
		object         metav1.Object
		owner          metav1.OwnerReference
		expectedResult bool
	}{
		{
			description:    "objectWithNoOwnerReferenceReturnsFalse",
			object:         &appsv1.Deployment{},
			owner:          ownerReference,
			expectedResult: false,
		},
		{
			description: "ownerReferenceWithWrongNameReturnsFalse",
			object: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{
						{
							Name: "NotFoo",
							Kind: "bar",
							UID:  "uid",
						},
					},
				},
			},
			owner:          ownerReference,
			expectedResult: false,
		},
		{
			description: "ownerReferenceWithWrongKindReturnsFalse",
			object: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{
						{
							Name: "foo",
							Kind: "notBar",
							UID:  "uid",
						},
					},
				},
			},
			owner:          ownerReference,
			expectedResult: false,
		},
		{
			description: "ownerReferenceWithWrongUIDReturnsFalse",
			object: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{
						{
							Name: "foo",
							Kind: "bar",
							UID:  "notUID",
						},
					},
				},
			},
			owner:          ownerReference,
			expectedResult: false,
		},
		{
			description: "objectWithOwnerReferenceReturnsTrue",
			object: &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{
						ownerReference,
					},
				},
			},
			owner:          ownerReference,
			expectedResult: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			require.Equal(t, tt.expectedResult, HasOwnerReference(tt.object, tt.owner))
		})
	}
}

/*
package ownerutil

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = FDescribe("The HasOwnerReference Function", func() {
	ownerReference := metav1.OwnerReference{
		Name: "foo",
		Kind: "bar",
		UID:  "uid",
	}
	It("Returns false when an object does not contain the given OwnerReference", func() {
		Expect(HasOwnerReference(&appsv1.Deployment{}, ownerReference)).ToNot(BeTrue())
	})
	It("Returns false when the object has an OwnerReference with a different name", func() {
		Expect(HasOwnerReference(&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				OwnerReferences: []metav1.OwnerReference{
					{
						Name: "notFoo",
						Kind: "bar",
						UID:  "uid",
					},
				},
			},
		}, ownerReference)).ToNot(BeTrue())
	})

	It("Returns false when the object has an OwnerReference with a different kind", func() {
		Expect(HasOwnerReference(&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				OwnerReferences: []metav1.OwnerReference{
					{
						Name: "foo",
						Kind: "notBar",
						UID:  "uid",
					},
				},
			},
		}, ownerReference)).ToNot(BeTrue())
	})

	It("Returns false when the object has an OwnerReference with a different UID", func() {
		Expect(HasOwnerReference(&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				OwnerReferences: []metav1.OwnerReference{
					{
						Name: "foo",
						Kind: "bar",
						UID:  "notUID",
					},
				},
			},
		}, ownerReference)).ToNot(BeTrue())
	})
	It("Returns true when the object has an OwnerReference", func() {
		Expect(HasOwnerReference(&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				OwnerReferences: []metav1.OwnerReference{
					{
						Name: "foo",
						Kind: "bar",
						UID:  "uid",
					},
				},
			},
		}, ownerReference)).ToNot(BeFalse())
	})
})

*/
