// +build !ignore_autogenerated

// Code generated by openapi-gen. DO NOT EDIT.

// This file was autogenerated by openapi-gen. Do not edit it manually!

package v1alpha1

import (
	spec "github.com/go-openapi/spec"
	common "k8s.io/kube-openapi/pkg/common"
)

func GetOpenAPIDefinitions(ref common.ReferenceCallback) map[string]common.OpenAPIDefinition {
	return map[string]common.OpenAPIDefinition{
		"service-cache-operator/pkg/apis/cache/v1alpha1.ServiceCache":       schema_pkg_apis_cache_v1alpha1_ServiceCache(ref),
		"service-cache-operator/pkg/apis/cache/v1alpha1.ServiceCacheSpec":   schema_pkg_apis_cache_v1alpha1_ServiceCacheSpec(ref),
		"service-cache-operator/pkg/apis/cache/v1alpha1.ServiceCacheStatus": schema_pkg_apis_cache_v1alpha1_ServiceCacheStatus(ref),
	}
}

func schema_pkg_apis_cache_v1alpha1_ServiceCache(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "ServiceCache is the Schema for the servicecaches API",
				Properties: map[string]spec.Schema{
					"kind": {
						SchemaProps: spec.SchemaProps{
							Description: "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"apiVersion": {
						SchemaProps: spec.SchemaProps{
							Description: "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"metadata": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"),
						},
					},
					"spec": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("service-cache-operator/pkg/apis/cache/v1alpha1.ServiceCacheSpec"),
						},
					},
					"status": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("service-cache-operator/pkg/apis/cache/v1alpha1.ServiceCacheStatus"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta", "service-cache-operator/pkg/apis/cache/v1alpha1.ServiceCacheSpec", "service-cache-operator/pkg/apis/cache/v1alpha1.ServiceCacheStatus"},
	}
}

func schema_pkg_apis_cache_v1alpha1_ServiceCacheSpec(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "ServiceCacheSpec defines the desired state of ServiceCache",
				Properties:  map[string]spec.Schema{},
			},
		},
		Dependencies: []string{},
	}
}

func schema_pkg_apis_cache_v1alpha1_ServiceCacheStatus(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "ServiceCacheStatus defines the observed state of ServiceCache",
				Properties:  map[string]spec.Schema{},
			},
		},
		Dependencies: []string{},
	}
}
