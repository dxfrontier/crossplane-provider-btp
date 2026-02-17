package v1alpha1

func (md *Directory) SetExternalID(newID string) {
	md.Spec.ForProvider.DisplayName = &newID
}

func (md *Directory) GetExternalID() string {
	return *md.Spec.ForProvider.DisplayName
}

// NOTE: Subaccount external ID functions removed — upjet-generated type uses *string fields.
// After `make generate`, update if the Tested interface is still needed.
