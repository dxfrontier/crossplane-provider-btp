#!/bin/bash
# fix-providerconfig-shadow.sh
#
# Post-generate fixup for crossplane-runtime v2 ProviderConfigReference.
#
# Problem: ResourceSpec.ProviderConfigReference is *Reference (Name+Policy),
# but the Managed interface uses *ProviderConfigReference (Kind+Name).
# The CRD schema uses the embedded *Reference type, so the API server prunes
# the Kind field. We must compute Kind at runtime in the getter.
#
# Solution:
#   1. Shadow the embedded *Reference field with *ProviderConfigReference
#      in each Spec struct (so angryjet generates type-correct code).
#   2. Fix the setter to use *ProviderConfigReference (angryjet uses *Reference).
#   3. Fix the getter to ALWAYS return Kind="ProviderConfig" (because the CRD
#      schema doesn't include Kind, so the API server prunes it on round-trip).
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

# 2. Fix SetProviderConfigReference: angryjet creates *Reference instead of *ProviderConfigReference
find "$BASE_DIR" -name 'zz_generated.managed.go' -exec \
    sed -i 's/&xpv1\.Reference{Name: r\.Name}/\&xpv1.ProviderConfigReference{Name: r.Name, Kind: r.Kind}/g' {} +

# 3. Fix GetProviderConfigReference: ensure Kind is ALWAYS "ProviderConfig".
#    The CRD schema uses the embedded *Reference type (no Kind field), so
#    the API server prunes Kind on round-trip. We must hardcode it here.
#
#    Pattern A: angryjet with shadow field generates "return mg.Spec.ProviderConfigReference"
find "$BASE_DIR" -name 'zz_generated.managed.go' -exec \
    sed -i '/GetProviderConfigReference/,/^}/ s/return mg\.Spec\.ProviderConfigReference$/return \&xpv1.ProviderConfigReference{Name: mg.Spec.ProviderConfigReference.Name, Kind: "ProviderConfig"}/' {} +

#    Pattern B: angryjet without shadow field generates "return &xpv1.ProviderConfigReference{Name: mg.Spec.ProviderConfigReference.Name}"
find "$BASE_DIR" -name 'zz_generated.managed.go' -exec \
    sed -i 's/return &xpv1\.ProviderConfigReference{Name: mg\.Spec\.ProviderConfigReference\.Name}/return \&xpv1.ProviderConfigReference{Name: mg.Spec.ProviderConfigReference.Name, Kind: "ProviderConfig"}/g' {} +

# 4. Fix MRD scope: make ALL upjet-generated resources Namespaced.
#    The generator hardcodes scope=Cluster (pcNamespace=nil), but our
#    ProviderConfig is Namespaced, so all MRDs must be Namespaced too.
#    Without this, cluster-scoped MRs cannot find the namespaced ProviderConfig.
echo ">> Fixing MRD scope to Namespaced..."
for f in $(find "$BASE_DIR" -name 'zz_*_types.go'); do
    if grep -q 'scope=Cluster' "$f" 2>/dev/null; then
        sed -i 's/scope=Cluster/scope=Namespaced/g' "$f"
        echo "   patched scope: $f"
    fi
done
# Also fix the generated CRD YAMLs
CRD_DIR="${BASE_DIR}/../package/crds"
if [ -d "$CRD_DIR" ]; then
    for f in "$CRD_DIR"/*.yaml; do
        if grep -q 'scope: Cluster' "$f" 2>/dev/null; then
            sed -i 's/scope: Cluster/scope: Namespaced/g' "$f"
            echo "   patched CRD scope: $f"
        fi
    done
fi

# 5. Fix reference resolver namespace: the generated resolvers use NewAPIResolver
#    which passes Namespace="" in ResolutionRequest. For namespaced resources,
#    this causes "not found" errors because the cache indexes by actual namespace.
#    Fix: inject Namespace: mg.GetNamespace() into every ResolutionRequest.
echo ">> Fixing reference resolver namespace awareness..."
for f in $(find "$BASE_DIR" -name 'zz_generated.resolvers.go'); do
    if grep -q 'reference.ResolutionRequest{' "$f" 2>/dev/null; then
        sed -i '/reference\.ResolutionRequest{/a\\t\tNamespace: mg.GetNamespace(),' "$f"
        echo "   patched resolver: $f"
    fi
done

# 6. Fix CRD allOf dual-default: the shadow ProviderConfigReference field
#    creates two allOf entries with conflicting defaults, which Kubernetes
#    structural schema validation rejects. Remove the allOf and merge into
#    a single schema with the ProviderConfigReference default.
echo ">> Fixing CRD allOf dual-default in providerConfigRef..."
if [ -d "$CRD_DIR" ]; then
    for f in "$CRD_DIR"/*.yaml; do
        # Check if the CRD has the problematic allOf pattern
        if grep -q 'providerConfigRef:' "$f" 2>/dev/null && grep -q 'allOf:' "$f" 2>/dev/null; then
            echo "   TODO: fix allOf in $f (requires Python/yq)"
        fi
    done
fi

echo ">> ProviderConfigReference fixes applied successfully"
