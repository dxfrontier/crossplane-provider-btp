package btp_subaccount

import (
	"context"

	"github.com/crossplane/upjet/v2/pkg/config"
)

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("btp_subaccount", func(r *config.Resource) {
		r.ShortGroup = "account"
		r.Kind = "Subaccount"
		r.UseAsync = true

		r.ExternalName.GetIDFn = func(_ context.Context, externalName string,
			_ map[string]any, _ map[string]any) (string, error) {
			if externalName == "" {
				return "NOT_EMPTY_GUID", nil
			}
			return externalName, nil
		}

		// parent_id → Crossplane reference to Directory
		r.References["parent_id"] = config.Reference{
			Type:              "github.com/sap/crossplane-provider-btp/apis/account/v1alpha1.Directory",
			Extractor:         "github.com/sap/crossplane-provider-btp/apis/account/v1alpha1.DirectoryUuid()",
			RefFieldName:      "ParentRef",
			SelectorFieldName: "ParentSelector",
		}
	})
}
