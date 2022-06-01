package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type SilenceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Silence `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Silence struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   SilenceSpec   `json:"spec,omitempty"`
	Status SilenceStatus `json:"status,omitempty"`
}

type SilenceSpec struct {
	URL       string      `json:"url"`
	Matchers  []Matcher   `json:"matchers,omitempty"`
	StartsAt  metav1.Time `json:"startsAt"`
	EndsAt    metav1.Time `json:"endsAt"`
	CreatedBy string      `json:"createdBy"`
	Comment   string      `json:"comment"`
}

type SilenceStatus struct {
	ID    string `json:"id,omitempty"`
	State string `json:"state,omitempty"`
}

type Matcher struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Regex bool   `json:"isRegex"`
	Equal bool   `json:"isEqual"`
}
