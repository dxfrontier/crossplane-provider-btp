#!/bin/bash
# fix-providerconfig-shadow.sh
#
# Post-generate fixup for crossplane-runtime v2 ProviderConfigReference.
#
# Problem: ResourceSpec.ProviderConfigReference is *Reference (Name+Policy),
# but the Managed interface uses *ProviderConfigReference (Kind+Name).
# angryjet generates adapter code that loses the Kind field.
#
# Solution: Shadow the embedded *Reference field with *ProviderConfigReference
# in each Spec struct, then fix the generated getter/setter adapters.
#
# Usage:
#   From repo root (Makefile):  bash scripts/fix-providerconfig-shadow.sh
#   From apis/ (go:generate):   bash ../scripts/fix-providerconfig-shadow.sh .

set -euo pipefail

BASE_DIR="${1:-apis}"  # Default to 'apis' when run from repo root

echo ">> Fixing ProviderConfigReference shadow fields (base: $BASE_DIR)..."

# 1. Add shadow field to upjet-generated types that embed v1.ResourceSpec
for f in $(find "$BASE_DIR" -name 'zz_*_types.go'); do
    # Skip if shadow field already exists (hand-written types or re-run)
    if grep -q 'ProviderConfigReference \*v1\.ProviderConfigReference' "$f" 2>/dev/null; then
        continue
    fi
    # Skip if no ResourceSpec embedding (not a managed resource spec)
    if ! grep -q 'v1\.ResourceSpec' "$f" 2>/dev/null; then
        continue
    fi
    # Add shadow field after v1.ResourceSpec line
    awk '
    /v1\.ResourceSpec.*json:",inline"/ {
        print
        print "\t// ProviderConfigReference shadows ResourceSpec.ProviderConfigReference"
        print "\t// to include the Kind field required by crossplane-runtime v2."
        print "\t// +kubebuilder:default={\"name\":\"default\",\"kind\":\"ProviderConfig\"}"
        print "\tProviderConfigReference *v1.ProviderConfigReference `json:\"providerConfigRef,omitempty\"`"
        next
    }
    { print }
    ' "$f" > "${f}.tmp" && mv "${f}.tmp" "$f"
    echo "   patched types: $f"
done

# 2. Fix SetProviderConfigReference: creates *Reference instead of *ProviderConfigReference
find "$BASE_DIR" -name 'zz_generated.managed.go' -exec \
    sed -i 's/&xpv1\.Reference{Name: r\.Name}/\&xpv1.ProviderConfigReference{Name: r.Name, Kind: r.Kind}/g' {} +

# 3. Fix GetProviderConfigReference: copies only Name, losing Kind
find "$BASE_DIR" -name 'zz_generated.managed.go' -exec \
    sed -i 's/return &xpv1\.ProviderConfigReference{Name: mg\.Spec\.ProviderConfigReference\.Name}/return mg.Spec.ProviderConfigReference/g' {} +

echo ">> ProviderConfigReference fixes applied successfully"
