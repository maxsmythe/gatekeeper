# sigs.k8s.io/controller-runtime

Forked from sigs.k8s.io/controller-runtime@6533c61cea9b083b68d8379fbe88c05b269cc4a2 (v0.13.1).

This fork adds the ability to dynamically
remove informers from the informer cache. Additionally, non-blocking APIs were added to fetch informers
without waiting for cache sync. `pkg/cache` was renamed to `pkg/dynamiccache` for clarity.

The original code can be found at https://github.com/kubernetes-sigs/controller-runtime/tree/v0.13.1/pkg/cache.

