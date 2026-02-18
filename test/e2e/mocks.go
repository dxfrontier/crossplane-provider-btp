package e2e

import (
	"encoding/json"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/e2e-framework/klient/k8s"
)

type MockList struct {
	client.ObjectList

	Items []k8s.Object
}
type FakeManaged struct {
	metav1.TypeMeta
	metav1.ObjectMeta
	xpv1.ConditionedStatus
	ManagementPolicies xpv1.ManagementPolicies
}

func (m *FakeManaged) SetManagementPolicies(p xpv1.ManagementPolicies) {
	m.ManagementPolicies = p
}

func (m *FakeManaged) GetManagementPolicies() xpv1.ManagementPolicies {
	return m.ManagementPolicies
}

// DeepCopyObject returns a copy of the object as runtime.Object
func (m *FakeManaged) DeepCopyObject() runtime.Object {
	out := &FakeManaged{}
	j, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	_ = json.Unmarshal(j, out)
	return out
}
