package status

import (
	"context"
	"testing"

	lokiv1 "github.com/grafana/loki/operator/apis/loki/v1"
	"github.com/grafana/loki/operator/internal/external/k8s/k8sfakes"

	"github.com/stretchr/testify/require"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func setupFakesNoError(t *testing.T, stack *lokiv1.LokiStack) (*k8sfakes.FakeClient, *k8sfakes.FakeStatusWriter) {
	sw := &k8sfakes.FakeStatusWriter{}
	k := &k8sfakes.FakeClient{}
	k.GetStub = func(_ context.Context, name types.NamespacedName, object client.Object, _ ...client.GetOption) error {
		if name.Name == stack.Name && name.Namespace == stack.Namespace {
			k.SetClientObject(object, stack)
			return nil
		}
		return apierrors.NewNotFound(schema.GroupResource{}, "something wasn't found")
	}
	k.StatusStub = func() client.StatusWriter { return sw }

	sw.UpdateStub = func(_ context.Context, obj client.Object, _ ...client.UpdateOption) error {
		actual := obj.(*lokiv1.LokiStack)
		require.NotEmpty(t, actual.Status.Conditions)
		require.Equal(t, metav1.ConditionTrue, actual.Status.Conditions[0].Status)
		return nil
	}

	return k, sw
}

func TestSetReadyCondition_WhenGetLokiStackReturnsError_ReturnError(t *testing.T) {
	r := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "my-stack",
			Namespace: "some-ns",
		},
	}

	k := &k8sfakes.FakeClient{}
	k.GetStub = func(_ context.Context, name types.NamespacedName, object client.Object, _ ...client.GetOption) error {
		return apierrors.NewBadRequest("something wasn't found")
	}

	err := SetReadyCondition(context.Background(), k, r)
	require.Error(t, err)
}

func TestSetReadyCondition_WhenGetLokiStackReturnsNotFound_DoNothing(t *testing.T) {
	r := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "my-stack",
			Namespace: "some-ns",
		},
	}

	k := &k8sfakes.FakeClient{}
	k.GetStub = func(_ context.Context, name types.NamespacedName, object client.Object, _ ...client.GetOption) error {
		return apierrors.NewNotFound(schema.GroupResource{}, "something wasn't found")
	}

	err := SetReadyCondition(context.Background(), k, r)
	require.NoError(t, err)
}

func TestSetReadyCondition_WhenExisting_DoNothing(t *testing.T) {
	s := lokiv1.LokiStack{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-stack",
			Namespace: "some-ns",
		},
		Status: lokiv1.LokiStackStatus{
			Conditions: []metav1.Condition{
				{
					Type:    string(lokiv1.ConditionReady),
					Message: messageReady,
					Reason:  string(lokiv1.ReasonReadyComponents),
					Status:  metav1.ConditionTrue,
				},
			},
		},
	}

	r := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "my-stack",
			Namespace: "some-ns",
		},
	}

	k, _ := setupFakesNoError(t, &s)

	err := SetReadyCondition(context.Background(), k, r)
	require.NoError(t, err)
	require.Zero(t, k.StatusCallCount())
}

func TestSetReadyCondition_WhenExisting_SetReadyConditionTrue(t *testing.T) {
	s := lokiv1.LokiStack{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-stack",
			Namespace: "some-ns",
		},
		Status: lokiv1.LokiStackStatus{
			Conditions: []metav1.Condition{
				{
					Type:   string(lokiv1.ConditionReady),
					Status: metav1.ConditionFalse,
				},
			},
		},
	}

	r := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "my-stack",
			Namespace: "some-ns",
		},
	}

	k, sw := setupFakesNoError(t, &s)

	err := SetReadyCondition(context.Background(), k, r)
	require.NoError(t, err)

	require.NotZero(t, k.StatusCallCount())
	require.NotZero(t, sw.UpdateCallCount())
}

func TestSetReadyCondition_WhenNoneExisting_AppendReadyCondition(t *testing.T) {
	s := lokiv1.LokiStack{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-stack",
			Namespace: "some-ns",
		},
	}

	r := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "my-stack",
			Namespace: "some-ns",
		},
	}

	k, sw := setupFakesNoError(t, &s)

	err := SetReadyCondition(context.Background(), k, r)
	require.NoError(t, err)

	require.NotZero(t, k.StatusCallCount())
	require.NotZero(t, sw.UpdateCallCount())
}

func TestSetFailedCondition_WhenGetLokiStackReturnsError_ReturnError(t *testing.T) {
	r := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "my-stack",
			Namespace: "some-ns",
		},
	}

	k := &k8sfakes.FakeClient{}
	k.GetStub = func(_ context.Context, name types.NamespacedName, object client.Object, _ ...client.GetOption) error {
		return apierrors.NewBadRequest("something wasn't found")
	}

	err := SetFailedCondition(context.Background(), k, r)
	require.Error(t, err)
}

func TestSetFailedCondition_WhenGetLokiStackReturnsNotFound_DoNothing(t *testing.T) {
	r := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "my-stack",
			Namespace: "some-ns",
		},
	}

	k := &k8sfakes.FakeClient{}
	k.GetStub = func(_ context.Context, name types.NamespacedName, object client.Object, _ ...client.GetOption) error {
		return apierrors.NewNotFound(schema.GroupResource{}, "something wasn't found")
	}

	err := SetFailedCondition(context.Background(), k, r)
	require.NoError(t, err)
}

func TestSetFailedCondition_WhenExisting_DoNothing(t *testing.T) {
	s := lokiv1.LokiStack{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-stack",
			Namespace: "some-ns",
		},
		Status: lokiv1.LokiStackStatus{
			Conditions: []metav1.Condition{
				{
					Type:    string(lokiv1.ConditionFailed),
					Reason:  string(lokiv1.ReasonFailedComponents),
					Message: messageFailed,
					Status:  metav1.ConditionTrue,
				},
			},
		},
	}

	r := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "my-stack",
			Namespace: "some-ns",
		},
	}

	k, _ := setupFakesNoError(t, &s)

	err := SetFailedCondition(context.Background(), k, r)
	require.NoError(t, err)
	require.Zero(t, k.StatusCallCount())
}

func TestSetFailedCondition_WhenExisting_SetFailedConditionTrue(t *testing.T) {
	s := lokiv1.LokiStack{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-stack",
			Namespace: "some-ns",
		},
		Status: lokiv1.LokiStackStatus{
			Conditions: []metav1.Condition{
				{
					Type:   string(lokiv1.ConditionFailed),
					Status: metav1.ConditionFalse,
				},
			},
		},
	}

	r := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "my-stack",
			Namespace: "some-ns",
		},
	}

	k, sw := setupFakesNoError(t, &s)

	err := SetFailedCondition(context.Background(), k, r)
	require.NoError(t, err)

	require.NotZero(t, k.StatusCallCount())
	require.NotZero(t, sw.UpdateCallCount())
}

func TestSetFailedCondition_WhenNoneExisting_AppendFailedCondition(t *testing.T) {
	s := lokiv1.LokiStack{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-stack",
			Namespace: "some-ns",
		},
	}

	r := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "my-stack",
			Namespace: "some-ns",
		},
	}

	k, sw := setupFakesNoError(t, &s)

	err := SetFailedCondition(context.Background(), k, r)
	require.NoError(t, err)

	require.NotZero(t, k.StatusCallCount())
	require.NotZero(t, sw.UpdateCallCount())
}

func TestSetDegradedCondition_WhenGetLokiStackReturnsError_ReturnError(t *testing.T) {
	msg := "tell me nothing"
	reason := lokiv1.ReasonMissingObjectStorageSecret

	r := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "my-stack",
			Namespace: "some-ns",
		},
	}

	k := &k8sfakes.FakeClient{}
	k.GetStub = func(_ context.Context, name types.NamespacedName, object client.Object, _ ...client.GetOption) error {
		return apierrors.NewBadRequest("something wasn't found")
	}

	err := SetDegradedCondition(context.Background(), k, r, msg, reason)
	require.Error(t, err)
}

func TestSetPendingCondition_WhenGetLokiStackReturnsError_ReturnError(t *testing.T) {
	r := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "my-stack",
			Namespace: "some-ns",
		},
	}

	k := &k8sfakes.FakeClient{}
	k.GetStub = func(_ context.Context, name types.NamespacedName, object client.Object, _ ...client.GetOption) error {
		return apierrors.NewBadRequest("something wasn't found")
	}

	err := SetPendingCondition(context.Background(), k, r)
	require.Error(t, err)
}

func TestSetPendingCondition_WhenGetLokiStackReturnsNotFound_DoNothing(t *testing.T) {
	r := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "my-stack",
			Namespace: "some-ns",
		},
	}

	k := &k8sfakes.FakeClient{}
	k.GetStub = func(_ context.Context, name types.NamespacedName, object client.Object, _ ...client.GetOption) error {
		return apierrors.NewNotFound(schema.GroupResource{}, "something wasn't found")
	}

	err := SetPendingCondition(context.Background(), k, r)
	require.NoError(t, err)
}

func TestSetPendingCondition_WhenExisting_DoNothing(t *testing.T) {
	s := lokiv1.LokiStack{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-stack",
			Namespace: "some-ns",
		},
		Status: lokiv1.LokiStackStatus{
			Conditions: []metav1.Condition{
				{
					Type:    string(lokiv1.ConditionPending),
					Reason:  string(lokiv1.ReasonPendingComponents),
					Message: messagePending,
					Status:  metav1.ConditionTrue,
				},
			},
		},
	}

	r := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "my-stack",
			Namespace: "some-ns",
		},
	}

	k, _ := setupFakesNoError(t, &s)

	err := SetPendingCondition(context.Background(), k, r)
	require.NoError(t, err)
	require.Zero(t, k.StatusCallCount())
}

func TestSetPendingCondition_WhenExisting_SetPendingConditionTrue(t *testing.T) {
	s := lokiv1.LokiStack{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-stack",
			Namespace: "some-ns",
		},
		Status: lokiv1.LokiStackStatus{
			Conditions: []metav1.Condition{
				{
					Type:   string(lokiv1.ConditionPending),
					Status: metav1.ConditionFalse,
				},
			},
		},
	}

	r := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "my-stack",
			Namespace: "some-ns",
		},
	}

	k, sw := setupFakesNoError(t, &s)

	err := SetPendingCondition(context.Background(), k, r)
	require.NoError(t, err)
	require.NotZero(t, k.StatusCallCount())
	require.NotZero(t, sw.UpdateCallCount())
}

func TestSetPendingCondition_WhenNoneExisting_AppendPendingCondition(t *testing.T) {
	s := lokiv1.LokiStack{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-stack",
			Namespace: "some-ns",
		},
	}

	r := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "my-stack",
			Namespace: "some-ns",
		},
	}

	k, sw := setupFakesNoError(t, &s)

	err := SetPendingCondition(context.Background(), k, r)
	require.NoError(t, err)

	require.NotZero(t, k.StatusCallCount())
	require.NotZero(t, sw.UpdateCallCount())
}

func TestSetDegradedCondition_WhenGetLokiStackReturnsNotFound_DoNothing(t *testing.T) {
	msg := "tell me nothing"
	reason := lokiv1.ReasonMissingObjectStorageSecret

	r := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "my-stack",
			Namespace: "some-ns",
		},
	}

	k := &k8sfakes.FakeClient{}
	k.GetStub = func(_ context.Context, name types.NamespacedName, object client.Object, _ ...client.GetOption) error {
		return apierrors.NewNotFound(schema.GroupResource{}, "something wasn't found")
	}

	err := SetDegradedCondition(context.Background(), k, r, msg, reason)
	require.NoError(t, err)
}

func TestSetDegradedCondition_WhenExisting_DoNothing(t *testing.T) {
	msg := "tell me nothing"
	reason := lokiv1.ReasonMissingObjectStorageSecret
	s := lokiv1.LokiStack{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-stack",
			Namespace: "some-ns",
		},
		Status: lokiv1.LokiStackStatus{
			Conditions: []metav1.Condition{
				{
					Type:    string(lokiv1.ConditionDegraded),
					Reason:  string(reason),
					Message: msg,
					Status:  metav1.ConditionTrue,
				},
			},
		},
	}

	r := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "my-stack",
			Namespace: "some-ns",
		},
	}

	k, _ := setupFakesNoError(t, &s)

	err := SetDegradedCondition(context.Background(), k, r, msg, reason)
	require.NoError(t, err)
	require.Zero(t, k.StatusCallCount())
}

func TestSetDegradedCondition_WhenExisting_SetDegradedConditionTrue(t *testing.T) {
	msg := "tell me something"
	reason := lokiv1.ReasonMissingObjectStorageSecret
	s := lokiv1.LokiStack{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-stack",
			Namespace: "some-ns",
		},
		Status: lokiv1.LokiStackStatus{
			Conditions: []metav1.Condition{
				{
					Type:   string(lokiv1.ConditionDegraded),
					Reason: string(reason),
					Status: metav1.ConditionFalse,
				},
			},
		},
	}

	r := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "my-stack",
			Namespace: "some-ns",
		},
	}

	k, sw := setupFakesNoError(t, &s)

	err := SetDegradedCondition(context.Background(), k, r, msg, reason)
	require.NoError(t, err)
	require.NotZero(t, k.StatusCallCount())
	require.NotZero(t, sw.UpdateCallCount())
}

func TestSetDegradedCondition_WhenNoneExisting_AppendDegradedCondition(t *testing.T) {
	msg := "tell me something"
	reason := lokiv1.ReasonMissingObjectStorageSecret
	s := lokiv1.LokiStack{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-stack",
			Namespace: "some-ns",
		},
	}

	r := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "my-stack",
			Namespace: "some-ns",
		},
	}

	k, sw := setupFakesNoError(t, &s)

	err := SetDegradedCondition(context.Background(), k, r, msg, reason)
	require.NoError(t, err)

	require.NotZero(t, k.StatusCallCount())
	require.NotZero(t, sw.UpdateCallCount())
}
