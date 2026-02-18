package tfclient

import (
	"context"

	ujresource "github.com/crossplane/upjet/v2/pkg/resource"
	"github.com/crossplane/upjet/v2/pkg/terraform"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var errUpdateStatusFmt = "cannot update status of the resource %s after an async %s"

func NewAPICallbacks(kube client.Client, saveConditionsFn SaveConditionsFn) *APICallbacks {
	return &APICallbacks{
		kube:           kube,
		saveCallbackFn: saveConditionsFn,
	}
}

type APICallbacks struct {
	kube client.Client

	saveCallbackFn SaveConditionsFn
}

// Create makes sure the error is saved in async operation condition.
func (ac *APICallbacks) Create(nn types.NamespacedName) terraform.CallbackFn {
	return func(err error, ctx context.Context) error {
		uErr := ac.saveCallbackFn(ctx, ac.kube, nn.String(), ujresource.LastAsyncOperationCondition(err), ujresource.AsyncOperationFinishedCondition())
		return errors.Wrapf(uErr, errUpdateStatusFmt, nn.String(), "create")
	}
}

// Update makes sure the error is saved in async operation condition.
func (ac *APICallbacks) Update(nn types.NamespacedName) terraform.CallbackFn {
	return func(err error, ctx context.Context) error {
		uErr := ac.saveCallbackFn(ctx, ac.kube, nn.String(), ujresource.LastAsyncOperationCondition(err), ujresource.AsyncOperationFinishedCondition())
		return errors.Wrapf(uErr, errUpdateStatusFmt, nn.String(), "update")
	}
}

// Destroy makes sure the error is saved in async operation condition.
func (ac *APICallbacks) Destroy(nn types.NamespacedName) terraform.CallbackFn {
	return func(err error, ctx context.Context) error {
		uErr := ac.saveCallbackFn(ctx, ac.kube, nn.String(), ujresource.LastAsyncOperationCondition(err), ujresource.AsyncOperationFinishedCondition())
		return errors.Wrapf(uErr, errUpdateStatusFmt, nn.String(), "destroy")
	}
}
