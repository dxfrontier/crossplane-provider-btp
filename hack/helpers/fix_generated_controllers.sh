#!/bin/bash
# Fix upjet v2.2.0-generated controller files for compatibility with
# controller-runtime v0.23+ and the plain (non-SDK/Framework) connector.
#
# Issues fixed:
#
# 1. Webhook API: The template generates ctrl.NewWebhookManagedBy(mgr).For(&T{})
#    but controller-runtime v0.23+ uses the generic ctrl.NewWebhookManagedBy(mgr, &T{}).
#
# 2. Unused metrics import: The template unconditionally imports
#    "github.com/crossplane/upjet/v2/pkg/metrics" but it is only used by the
#    TerraformPluginSDK/Framework connector paths. For the plain CLI connector
#    the import is unused and causes a compilation error.
#
# 3. SetupGated: The template generates SetupGated wrappers and zz_setup.go
#    calls them, but providerconfig (a custom controller) only has Setup.
#    Since the Gate feature is not configured, replace all SetupGated with Setup.

set -euo pipefail

# go:generate runs from the directory of the source file (apis/),
# so navigate to the project root first.
cd "$(dirname "$0")/../.."

find internal/controller -name 'zz_controller.go' -exec perl -0777 -pi -e '
  # Fix 1: Rewrite webhook call to generic two-arg form
  s/ctrl\.NewWebhookManagedBy\(mgr\)\.\n\s+For\(([^)]+)\)\./ctrl.NewWebhookManagedBy(mgr, $1)./g;

  # Fix 2: Remove unused metrics import
  s/\t"github\.com\/crossplane\/upjet\/v2\/pkg\/metrics"\n//g;
' {} +

# Fix 3: Replace SetupGated with Setup in the generated setup orchestrator.
# providerconfig is a custom controller without SetupGated, and we don't
# use the Gate feature anyway.
sed -i 's/\.SetupGated/.Setup/g' internal/controller/zz_setup.go
