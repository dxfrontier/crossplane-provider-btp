# Crossplane Provider for SAP BTP (DXFrontier Fork)

Fork of [SAP/crossplane-provider-btp](https://github.com/SAP/crossplane-provider-btp) (upstream `v1.6.1`).

This fork fixes issues with custom Identity Provider (IDP) authentication that prevent the upstream provider from working in environments where the default SAP IDP is not used.

## Installation

```yaml
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-btp
spec:
  package: ghcr.io/dxfrontier/crossplane-provider-btp/crossplane/provider-btp:v1.6.1-dxf.6
```

## Changes vs Upstream

### Subaccount: Terraform-backed controller (replaces hand-written)

The hand-written Subaccount controller used the CIS API with `client_credentials` grant, which has no user context. This caused the creating user to be resolved via the default SAP IDP (`sap.default`) instead of the custom IDP — making the SA user unable to administer the subaccount (403 on ServiceManager, CloudManagement, etc.).

**Fix:** Replaced the hand-written controller with an upjet-generated Terraform-backed controller. Terraform authenticates via BTP CLI using `username + password + idp`, so the creator is automatically registered as subaccount admin with the correct IDP origin.

**Field changes (CRD):**

| Old (hand-written) | New (upjet/Terraform) |
|---|---|
| `forProvider.displayName` | `forProvider.name` |
| `forProvider.usedForProduction` | `forProvider.usage` |
| `forProvider.directoryRef` | `forProvider.parentRef` |
| `forProvider.globalAccountGuid` | _(removed)_ |
| `forProvider.subaccountAdmins` | _(removed — creator is auto-admin)_ |
| `status.atProvider.subaccountGuid` | `status.atProvider.id` |

### CF Environment: IDP origin for API login

The CloudFoundry environment controller did not pass the IDP origin when authenticating against the CF API, causing login failures with custom IDPs.

**Fix:** CF API login now passes the `origin` parameter from the ProviderConfig credentials.

## Upstream Documentation

For general usage, development setup, E2E tests, and provider configuration, see the [upstream README](https://github.com/SAP/crossplane-provider-btp#readme).

## Releases

Releases are published to GHCR under the `dxfrontier` org:

- **Provider package:** `ghcr.io/dxfrontier/crossplane-provider-btp/crossplane/provider-btp`
- **Controller image:** `ghcr.io/dxfrontier/crossplane-provider-btp/crossplane/provider-btp-controller`

Tags follow the pattern `v1.6.1-dxf.<n>` based on the upstream version.

## Contributing

Pull requests are welcome. For major changes, please open an issue first
to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License

Copyright 2024 SAP SE or an SAP affiliate company and crossplane-provider-btp contributors.

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for details.
