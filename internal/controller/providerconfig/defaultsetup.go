package providerconfig

import (
	"context"

	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	providerv1alpha1 "github.com/sap/crossplane-provider-btp/apis/v1alpha1"
	"github.com/sap/crossplane-provider-btp/internal/tracking"
)

// ModernTracker tracks managed resources. It wraps the v2
// ProviderConfigUsageTracker to accept resource.Managed and
// type-assert to ModernManaged internally.
type ModernTracker struct {
	inner *resource.ProviderConfigUsageTracker
}

// Track tracks the supplied managed resource. The mg must implement
// resource.ModernManaged.
func (t *ModernTracker) Track(ctx context.Context, mg resource.Managed) error {
	return t.inner.Track(ctx, mg.(resource.ModernManaged))
}

type ConnectorFn func(
	kube client.Client,
	usage resource.Tracker,
	resourcetracker tracking.ReferenceResolverTracker,
) managed.ExternalConnector

// DefaultSetup supports the creation of a controller for a given managed resource type. Accepts any type that implements the ConnectorFn or KymaModuleConnectorFn signature.
// DEPRECATED: use DefaultSetupWithoutDefaultInitializer instead to not have the external-name default initializer added automatically (new external-name handling requires external-name to be empty not defaulted on create).
func DefaultSetup(mgr ctrl.Manager, o controller.Options, object client.Object, kind string, gvk schema.GroupVersionKind, connectorFn ConnectorFn) error {
	name := managed.ControllerName(kind)

	referenceTracker := tracking.NewDefaultReferenceResolverTracker(
		mgr.GetClient(),
	)
	usageTracker := &ModernTracker{
		inner: resource.NewProviderConfigUsageTracker(
			mgr.GetClient(),
			&providerv1alpha1.ProviderConfigUsage{},
		),
	}

	r := managed.NewReconciler(
		mgr,
		resource.ManagedKind(gvk),
		managed.WithExternalConnector(connectorFn(mgr.GetClient(), usageTracker, referenceTracker)),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))), //nolint:staticcheck // SA1019 replacement API returns incompatible type
		managed.WithPollInterval(o.PollInterval),
		managed.WithManagementPolicies(),
	)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(object).
		WithEventFilter(resource.DesiredStateChanged()).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

// DefaultSetupWithoutDefaultInitializer works like DefaultSetup but without DefaultInitializer (to adhere to new external-name handling). It supports the creation of a controller for a given managed resource type. Accepts any type that implements the ConnectorFn or KymaModuleConnectorFn signature.
func DefaultSetupWithoutDefaultInitializer(mgr ctrl.Manager, o controller.Options, object client.Object, kind string, gvk schema.GroupVersionKind, connectorFn ConnectorFn) error {
	name := managed.ControllerName(kind)

	referenceTracker := tracking.NewDefaultReferenceResolverTracker(
		mgr.GetClient(),
	)
	usageTracker := &ModernTracker{
		inner: resource.NewProviderConfigUsageTracker(
			mgr.GetClient(),
			&providerv1alpha1.ProviderConfigUsage{},
		),
	}

	r := managed.NewReconciler(
		mgr,
		resource.ManagedKind(gvk),
		managed.WithExternalConnector(connectorFn(mgr.GetClient(), usageTracker, referenceTracker)),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))), //nolint:staticcheck // SA1019 replacement API returns incompatible type
		managed.WithPollInterval(o.PollInterval),
		managed.WithInitializers(), // No default initializer
		managed.WithManagementPolicies(),
	)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(object).
		WithEventFilter(resource.DesiredStateChanged()).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}
