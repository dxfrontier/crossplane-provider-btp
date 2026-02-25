package btp_subaccount_environment_instance

import (
	"context"

	"github.com/crossplane/upjet/v2/pkg/config"
)

// sentinelEnvironmentID is a non-existent UUID used as a placeholder when the
// resource has not yet been created (no external name annotation). This causes
// the BTP Provisioning API to return 404 (not found) instead of 403 (forbidden)
// which happens when the environment instance ID is empty and the API path
// resolves to GET /provisioning/v1/environments/ (the list endpoint).
// Terraform treats the 404 as "resource does not exist" and proceeds to Create.
const sentinelEnvironmentID = "00000000-0000-0000-0000-000000000000"

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("btp_subaccount_environment_instance", func(r *config.Resource) {
		r.ShortGroup = "environment"
		r.Kind = "SubaccountEnvironmentInstance"
		r.UseAsync = true

		// Fix: Override GetIDFn to return a sentinel UUID when the external name
		// is empty (resource not yet created). Without this fix, upjet writes
		// terraform state with id:"" which causes terraform refresh to call the
		// BTP TF provider's Read function with an empty environment instance ID.
		// The BTP CLI translates this to GET /provisioning/v1/environments/
		// (no ID path segment) which returns 403 Forbidden, blocking the Create.
		r.ExternalName = config.NewExternalNameFrom(config.IdentifierFromProvider,
			config.WithGetIDFn(func(origFn config.GetIDFn, ctx context.Context, externalName string, parameters map[string]any, providerConfig map[string]any) (string, error) {
				if externalName == "" {
					return sentinelEnvironmentID, nil
				}
				return origFn(ctx, externalName, parameters, providerConfig)
			}),
		)

		// subaccount_id → Reference to Subaccount
		r.References["subaccount_id"] = config.Reference{
			Type:              "github.com/sap/crossplane-provider-btp/apis/account/v1alpha1.Subaccount",
			Extractor:         "github.com/sap/crossplane-provider-btp/apis/account/v1alpha1.SubaccountUuid()",
			RefFieldName:      "SubaccountRef",
			SelectorFieldName: "SubaccountSelector",
		}
	})
}
