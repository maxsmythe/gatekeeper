package target

import (
	"encoding/json"

	"github.com/open-policy-agent/gatekeeper/pkg/mutation/types"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/admission"
	auditinternal "k8s.io/apiserver/pkg/apis/audit"
	"k8s.io/apiserver/pkg/authentication/user"
)

type AugmentedReview struct {
	AdmissionRequest *admissionv1.AdmissionRequest
	Namespace        *corev1.Namespace
	Source           types.SourceType
}

type gkReview struct {
	admissionv1.AdmissionRequest
	namespace *corev1.Namespace
	source    types.SourceType
}

// TODO FRICTION: can we avoid mocking unnecessary methods?
var _ admission.Attributes = &gkReview{}

func (r *gkReview) GetName() string {
	return r.Name
}

func (r *gkReview) GetNamespace() string {
	return r.Namespace
}

func (r *gkReview) GetResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{Group: r.Resource.Group, Version: r.Resource.Version, Resource: r.Resource.Resource}
}

func (r *gkReview) GetSubresource() string {
	return r.SubResource
}

func (r *gkReview) GetOperation() admission.Operation {
	return admission.Operation(r.Operation)
}

// TODO define this
func (r *gkReview) GetOperationOptions() runtime.Object {
	return nil
}

func (r *gkReview) IsDryRun() bool {
	return *r.DryRun
}

func (r *gkReview) GetObject() runtime.Object {
	if len(r.Object.Raw) == 0 {
		return nil
	}
	u := &unstructured.Unstructured{}
	// TODO figure out what to do if err != nil
	if err := json.Unmarshal(r.Object.Raw, u); err != nil {
		panic(err)
	}
	return u
}

func (r *gkReview) GetOldObject() runtime.Object {
	if len(r.OldObject.Raw) == 0 {
		return nil
	}
	u := &unstructured.Unstructured{}
	// TODO figure out what to do if err != nil
	if err := json.Unmarshal(r.OldObject.Raw, u); err != nil {
		panic(err)
	}
	return u
}

func (r *gkReview) GetKind() schema.GroupVersionKind {
	return schema.GroupVersionKind{Group: r.Kind.Group, Version: r.Kind.Version, Kind: r.Kind.Kind}
}

func (r *gkReview) GetUserInfo() user.Info {
	extra := map[string][]string{}
	for k, v := range r.UserInfo.Extra {
		extra[k] = []string(v)
	}
	return &user.DefaultInfo{Name: r.UserInfo.Username, UID: r.UserInfo.UID, Groups: r.UserInfo.Groups, Extra: extra}
}

func (r *gkReview) AddAnnotation(k, v string) error { return nil }

func (r *gkReview) AddAnnotationWithLevel(k, v string, l auditinternal.Level) error { return nil }

func (r *gkReview) GetReinvocationContext() admission.ReinvocationContext {
	return nil
}
