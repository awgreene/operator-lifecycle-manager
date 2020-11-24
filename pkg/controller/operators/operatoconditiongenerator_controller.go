package operators

import (
	"context"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	operatorsv1 "github.com/operator-framework/api/pkg/operators/v1"
	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/controller/operators/decorators"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/lib/ownerutil"
)

// OperatorReconciler reconciles a Operator object.
type OperatorConditionGenReconciler struct {
	client.Client
	factory decorators.OperatorFactory
	log     logr.Logger
}

// +kubebuilder:rbac:groups=operators.coreos.com,resources=operatorconditions,verbs=get;list;update;patch;delete
// +kubebuilder:rbac:groups=operators.coreos.com,resources=operatorconditions/status,verbs=update;patch

// SetupWithManager adds the OperatorCondition Reconciler reconciler to the given controller manager.
func (r *OperatorConditionGenReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorsv1.Operator{}).
		Complete(r)
}

// NewOperatorReconciler constructs and returns an OperatorReconciler.
// As a side effect, the given scheme has operator discovery types added to it.
func NewOperatorConditionGenReconciler(cli client.Client, log logr.Logger, scheme *runtime.Scheme) (*OperatorConditionGenReconciler, error) {
	// Add watched types to scheme.
	if err := AddToScheme(scheme); err != nil {
		return nil, err
	}

	factory, err := decorators.NewSchemedOperatorFactory(scheme)
	if err != nil {
		return nil, err
	}

	return &OperatorConditionGenReconciler{
		Client:  cli,
		log:     log,
		factory: factory,
	}, nil
}

// Implement reconcile.Reconciler so the controller can reconcile objects
var _ reconcile.Reconciler = &OperatorConditionGenReconciler{}

func (r *OperatorConditionGenReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	// Set up a convenient log object so we don't have to type request over and over again
	log := r.log.WithValues("request", req)
	log.V(1).Info("reconciling operatorcondition")

	in := &operatorsv1.Operator{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, in)
	if err != nil {
		log.V(1).Info("Unable to find operator", "error", err)
		return ctrl.Result{}, err
	}

	// Wrap with convenience decorator
	operator, err := r.factory.NewOperator(in)
	if err != nil {
		log.Error(err, "Could not wrap Operator with convenience decorator")
		return reconcile.Result{Requeue: true}, nil
	}

	selector, err := operator.ComponentSelector()
	if err != nil {
		return reconcile.Result{Requeue: true}, nil
	}
	operatorConditionList := operatorsv1.OperatorConditionList{}

	err = r.Client.List(context.TODO(), &operatorConditionList, &client.ListOptions{LabelSelector: selector})
	if err != nil {
		return ctrl.Result{}, err
	}
	if len(operatorConditionList.Items) == 0 {
		r.log.V(1).Info("Creating OperatorCondition")

		// Get CSV Namespace
		csvList := operatorsv1alpha1.ClusterServiceVersionList{}
		err = r.Client.List(context.TODO(), &csvList, &client.ListOptions{LabelSelector: selector})
		if err != nil {
			return ctrl.Result{}, err
		}
		if len(csvList.Items) != 1 {
			return ctrl.Result{}, err
		}
		labelKey, err := operator.ComponentLabelKey()
		if err != nil {
			return ctrl.Result{}, err
		}
		operatorCondition := operatorsv1.OperatorCondition{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: csvList.Items[0].GetName() + "-",
				Namespace:    csvList.Items[0].Namespace,
				Labels:       map[string]string{labelKey: ""},
			},
		}

		ownerutil.AddNonBlockingOwner(&operatorCondition, &csvList.Items[0])
		err = r.Client.Create(context.TODO(), &operatorCondition)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}
