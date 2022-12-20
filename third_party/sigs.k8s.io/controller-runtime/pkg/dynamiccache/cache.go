/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Modified from the original source (available at
// https://github.com/kubernetes-sigs/controller-runtime/tree/v0.13.1/pkg/cache)

package dynamiccache

import (
	"fmt"
	"time"

	"github.com/open-policy-agent/gatekeeper/third_party/sigs.k8s.io/controller-runtime/pkg/dynamiccache/internal"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	toolscache "k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("object-cache")

var defaultResyncTime = 10 * time.Hour

// New initializes and returns a new Cache.
func New(config *rest.Config, opts cache.Options) (cache.Cache, error) {
	opts, err := defaultOpts(config, opts)
	if err != nil {
		return nil, err
	}
	selectorsByGVK, err := convertToSelectorsByGVK(opts.SelectorsByObject, opts.DefaultSelector, opts.Scheme)
	if err != nil {
		return nil, err
	}
	disableDeepCopyByGVK, err := convertToDisableDeepCopyByGVK(opts.UnsafeDisableDeepCopyByObject, opts.Scheme)
	if err != nil {
		return nil, err
	}
	transformByGVK, err := convertToTransformByKindAndGVK(opts.TransformByObject, opts.DefaultTransform, opts.Scheme)
	if err != nil {
		return nil, err
	}

	im := internal.NewInformersMap(config, opts.Scheme, opts.Mapper, *opts.Resync, opts.Namespace, selectorsByGVK, disableDeepCopyByGVK, transformByGVK)
	return &dynamicInformerCache{InformersMap: im}, nil
}

// BuilderWithOptions returns a Cache constructor that will build the a cache
// honoring the options argument, this is useful to specify options like
// SelectorsByObject
// WARNING: If SelectorsByObject is specified, filtered out resources are not
// returned.
// WARNING: If UnsafeDisableDeepCopy is enabled, you must DeepCopy any object
// returned from cache get/list before mutating it.
func BuilderWithOptions(options cache.Options) cache.NewCacheFunc {
	return func(config *rest.Config, opts cache.Options) (cache.Cache, error) {
		if options.Scheme == nil {
			options.Scheme = opts.Scheme
		}
		if options.Mapper == nil {
			options.Mapper = opts.Mapper
		}
		if options.Resync == nil {
			options.Resync = opts.Resync
		}
		if options.Namespace == "" {
			options.Namespace = opts.Namespace
		}
		if opts.Resync == nil {
			opts.Resync = options.Resync
		}

		return New(config, options)
	}
}

func defaultOpts(config *rest.Config, opts cache.Options) (cache.Options, error) {
	// Use the default Kubernetes Scheme if unset
	if opts.Scheme == nil {
		opts.Scheme = scheme.Scheme
	}

	// Construct a new Mapper if unset
	if opts.Mapper == nil {
		var err error
		opts.Mapper, err = apiutil.NewDynamicRESTMapper(config)
		if err != nil {
			log.WithName("setup").Error(err, "Failed to get API Group-Resources")
			return opts, fmt.Errorf("could not create RESTMapper from config")
		}
	}

	// Default the resync period to 10 hours if unset
	if opts.Resync == nil {
		opts.Resync = &defaultResyncTime
	}
	return opts, nil
}

func convertToSelectorsByGVK(selectorsByObject cache.SelectorsByObject, defaultSelector cache.ObjectSelector, scheme *runtime.Scheme) (internal.SelectorsByGVK, error) {
	selectorsByGVK := internal.SelectorsByGVK{}
	for object, selector := range selectorsByObject {
		gvk, err := apiutil.GVKForObject(object, scheme)
		if err != nil {
			return nil, err
		}
		selectorsByGVK[gvk] = internal.Selector(selector)
	}
	selectorsByGVK[schema.GroupVersionKind{}] = internal.Selector(defaultSelector)
	return selectorsByGVK, nil
}

func convertToDisableDeepCopyByGVK(disableDeepCopyByObject cache.DisableDeepCopyByObject, scheme *runtime.Scheme) (internal.DisableDeepCopyByGVK, error) {
	disableDeepCopyByGVK := internal.DisableDeepCopyByGVK{}
	for obj, disable := range disableDeepCopyByObject {
		switch obj.(type) {
		case cache.ObjectAll, *cache.ObjectAll:
			disableDeepCopyByGVK[internal.GroupVersionKindAll] = disable
		default:
			gvk, err := apiutil.GVKForObject(obj, scheme)
			if err != nil {
				return nil, err
			}
			disableDeepCopyByGVK[gvk] = disable
		}
	}
	return disableDeepCopyByGVK, nil
}

// TransformByObject associate a client.Object's GVK to a transformer function
// to be applied when storing the object into the cache.
type TransformByObject map[client.Object]toolscache.TransformFunc

func convertToTransformByKindAndGVK(t cache.TransformByObject, defaultTransform toolscache.TransformFunc, scheme *runtime.Scheme) (internal.TransformFuncByObject, error) {
	result := internal.NewTransformFuncByObject()
	for obj, transformation := range t {
		if err := result.Set(obj, scheme, transformation); err != nil {
			return nil, err
		}
	}
	result.SetDefault(defaultTransform)
	return result, nil
}
