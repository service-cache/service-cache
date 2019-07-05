package utils

import (
	"sort"
	"strconv"
	"strings"

	cachev1alpha1 "service-cache-operator/pkg/apis/cache/v1alpha1"

	corev1 "k8s.io/api/core/v1"
)

// Diff the configuration between Service and ServiceCache objects.
// return true if has diff
func DiffServiceAndServiceCache(svc *corev1.Service, sc *cachev1alpha1.ServiceCache) bool {
	if svc == nil && sc == nil {
		return false
	}
	if (svc == nil && sc != nil) || (svc != nil && sc == nil) {
		return true
	}
  trimmedDefaultCacheable := strings.TrimSpace(svc.Annotations["service-cache.github.io/default"])
	if trimmedDefaultCacheable != strconv.FormatBool(sc.Spec.CacheableByDefault) {
		return true
	}

	trimmedCacheableURLsFromService := strings.TrimSpace(svc.Annotations["service-cache.github.io/URLs"])
	urlsFromService := strings.TrimSuffix(strings.TrimPrefix(trimmedCacheableURLsFromService, "["), "]")
	sliceOfUrlsFromService := strings.Split(urlsFromService, ",")
	sort.Strings(sliceOfUrlsFromService)
	sliceOfUrlsFromServiceCache := sc.Spec.URLs
	sort.Strings(sliceOfUrlsFromServiceCache)
	return strings.Join(sliceOfUrlsFromService, ",") != strings.Join(sliceOfUrlsFromServiceCache, ",")
}
