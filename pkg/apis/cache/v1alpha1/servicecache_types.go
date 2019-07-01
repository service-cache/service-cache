package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ServiceCacheSpec defines the desired state of ServiceCache
// +k8s:openapi-gen=true
type ServiceCacheSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	CacheableByDefault bool     `json:"service-cache.github.io/default"`
	URLs               []string `json:"service-cache.github.io/URLs"`
}

// ServiceCacheStatus defines the observed state of ServiceCache
// +k8s:openapi-gen=true
type ServiceCacheStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceCache is the Schema for the servicecaches API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type ServiceCache struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceCacheSpec   `json:"spec,omitempty"`
	Status ServiceCacheStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceCacheList contains a list of ServiceCache
type ServiceCacheList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceCache `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ServiceCache{}, &ServiceCacheList{})
}
