/*
Copyright 2021 The Tekton Authors

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

package v1beta2

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	pipelinev1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	triggersv1beta1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1beta1"
	"github.com/tektoncd/triggers/pkg/apis/triggers/v1beta2"
	triggersv1beta2 "github.com/tektoncd/triggers/pkg/apis/triggers/v1beta2"
	"github.com/tektoncd/triggers/test"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/ptr"
)

var (
	myClassAnnotation = map[string]string{v1beta2.EventListenerClassAnnotationKey: myClass}
	myObjectMeta      = metav1.ObjectMeta{
		Name:        "name",
		Namespace:   "namespace",
		Annotations: myClassAnnotation,
	}
)

func Test_EventListenerValidate(t *testing.T) {
	tests := []struct {
		name    string
		el      *EventListenerKubernetesResource
		wantErr *apis.FieldError
	}{{
		name: "TriggerTemplate Does Not Exist",
		el: &EventListenerKubernetesResource{triggersv1beta2.EventListener{
			ObjectMeta: myObjectMeta,
			Spec: triggersv1beta2.EventListenerSpec{
				Triggers: []triggersv1beta1.EventListenerTrigger{{
					Template: &triggersv1beta1.EventListenerTemplate{
						Ref:        ptr.String("dne"),
						APIVersion: "v1beta1",
					},
				}},
			},
		}},
	}, {
		name: "Valid EventListener No TriggerBinding",
		el: &EventListenerKubernetesResource{triggersv1beta2.EventListener{
			ObjectMeta: myObjectMeta,
			Spec: triggersv1beta2.EventListenerSpec{
				Triggers: []triggersv1beta1.EventListenerTrigger{{
					Template: &triggersv1beta1.EventListenerTemplate{
						Ref:        ptr.String("tt"),
						APIVersion: "v1beta1",
					},
				}},
			},
		}},
	}, {
		name: "Invalid EventListener name too long",
		el: &EventListenerKubernetesResource{triggersv1beta2.EventListener{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "name8888888888888888888888888888888888888888888888888888888888888888888",
				Namespace:   "namespace",
				Annotations: myClassAnnotation,
			},
			Spec: triggersv1beta2.EventListenerSpec{
				Triggers: []triggersv1beta1.EventListenerTrigger{{
					Template: &triggersv1beta1.EventListenerTemplate{
						Ref:        ptr.String("tt"),
						APIVersion: "v1beta1",
					},
				}},
			},
		}},
		wantErr: apis.ErrInvalidValue(`eventListener name 'name8888888888888888888888888888888888888888888888888888888888888888888' must be no more than 60 characters long`, "metadata.name"),
	}, {
		name: "Valid EventListener with TriggerRef",
		el: &EventListenerKubernetesResource{triggersv1beta2.EventListener{
			ObjectMeta: myObjectMeta,
			Spec: triggersv1beta2.EventListenerSpec{
				Triggers: []triggersv1beta1.EventListenerTrigger{{
					TriggerRef: "tt",
				}},
			},
		}},
	}, {
		name: "Valid EventListener with Annotation",
		el: &EventListenerKubernetesResource{triggersv1beta2.EventListener{
			ObjectMeta: myObjectMeta,
			Spec: triggersv1beta2.EventListenerSpec{
				Triggers: []triggersv1beta1.EventListenerTrigger{{
					TriggerRef: "tt",
				}},
			},
		}},
	}, {
		name: "Valid EventListener with TriggerBinding",
		el: &EventListenerKubernetesResource{triggersv1beta2.EventListener{
			ObjectMeta: myObjectMeta,
			Spec: triggersv1beta2.EventListenerSpec{
				Triggers: []triggersv1beta1.EventListenerTrigger{{
					Bindings: []*triggersv1beta1.EventListenerBinding{{
						Ref:        "tb",
						Kind:       triggersv1beta1.NamespacedTriggerBindingKind,
						APIVersion: "triggersv1beta1", // TODO: APIVersions seem wrong?
					}},
					Template: &triggersv1beta1.EventListenerTemplate{
						Ref: ptr.String("tt2"),
					},
				}},
			},
		}},
	}, {
		name: "Valid EventListener with ClusterTriggerBinding",
		el: &EventListenerKubernetesResource{triggersv1beta2.EventListener{
			ObjectMeta: myObjectMeta,
			Spec: triggersv1beta2.EventListenerSpec{
				Triggers: []triggersv1beta1.EventListenerTrigger{{
					Bindings: []*triggersv1beta1.EventListenerBinding{{
						Ref:        "tb",
						Kind:       triggersv1beta1.ClusterTriggerBindingKind,
						APIVersion: "v1alpha1",
					}},
					Template: &triggersv1beta1.EventListenerTemplate{
						Ref: ptr.String("tt2"),
					},
				}},
			},
		}},
	}, {
		name: "Valid EventListener with multiple TriggerBindings",
		el: &EventListenerKubernetesResource{triggersv1beta2.EventListener{
			ObjectMeta: myObjectMeta,
			Spec: triggersv1beta2.EventListenerSpec{
				Triggers: []triggersv1beta1.EventListenerTrigger{{
					Bindings: []*triggersv1beta1.EventListenerBinding{{
						Ref:        "tb1",
						Kind:       triggersv1beta1.ClusterTriggerBindingKind,
						APIVersion: "triggersv1beta1", // TODO: APIVersions seem wrong?
					}, {
						Ref:        "tb2",
						Kind:       triggersv1beta1.NamespacedTriggerBindingKind,
						APIVersion: "triggersv1beta1", // TODO: APIVersions seem wrong?
					}, {
						Ref:        "tb3",
						Kind:       triggersv1beta1.NamespacedTriggerBindingKind,
						APIVersion: "triggersv1beta1", // TODO: APIVersions seem wrong?
					}},
					Template: &triggersv1beta1.EventListenerTemplate{
						Ref: ptr.String("tt2"),
					},
				}},
			},
		}},
	}, {
		name: "Valid EventListener Webhook Interceptor",
		el: &EventListenerKubernetesResource{triggersv1beta2.EventListener{
			ObjectMeta: myObjectMeta,
			Spec: triggersv1beta2.EventListenerSpec{
				Triggers: []triggersv1beta1.EventListenerTrigger{{
					Interceptors: []*triggersv1beta1.EventInterceptor{{
						Webhook: &triggersv1beta1.WebhookInterceptor{
							ObjectRef: &corev1.ObjectReference{
								Name:       "svc",
								Kind:       "Service",
								APIVersion: "v1",
								Namespace:  "namespace",
							},
						},
					}},
					Bindings: []*triggersv1beta1.EventListenerBinding{{
						Ref:        "tb",
						Kind:       triggersv1beta1.NamespacedTriggerBindingKind,
						APIVersion: "triggersv1beta1", // TODO: APIVersions seem wrong?
					}},
					Template: &triggersv1beta1.EventListenerTemplate{
						Ref: ptr.String("tt2"),
					},
				}},
			},
		}},
	}, {
		name: "Valid EventListener Interceptor With Header",
		el: &EventListenerKubernetesResource{triggersv1beta2.EventListener{
			ObjectMeta: myObjectMeta,
			Spec: triggersv1beta2.EventListenerSpec{
				Triggers: []triggersv1beta1.EventListenerTrigger{{
					Interceptors: []*triggersv1beta1.EventInterceptor{{
						Webhook: &triggersv1beta1.WebhookInterceptor{
							Header: []pipelinev1alpha1.Param{{
								Name: "Valid-Header-Key",
								Value: pipelinev1alpha1.ArrayOrString{
									Type:      pipelinev1alpha1.ParamTypeString,
									StringVal: "valid-value",
								},
							}, {
								Name: "Valid-Header-Key2",
								Value: pipelinev1alpha1.ArrayOrString{
									Type:      pipelinev1alpha1.ParamTypeString,
									StringVal: "valid value 2",
								},
							}},
							ObjectRef: &corev1.ObjectReference{
								Name:       "svc",
								Kind:       "Service",
								APIVersion: "v1",
								Namespace:  "namespace",
							},
						},
					}},
					Bindings: []*triggersv1beta1.EventListenerBinding{{
						Ref:        "tb",
						Kind:       triggersv1beta1.NamespacedTriggerBindingKind,
						APIVersion: "triggersv1beta1", // TODO: APIVersions seem wrong?
					}},
					Template: &triggersv1beta1.EventListenerTemplate{
						Ref: ptr.String("tt2"),
					},
				}},
			},
		}},
	}, {
		name: "Valid EventListener Two Triggers",
		el: &EventListenerKubernetesResource{triggersv1beta2.EventListener{
			ObjectMeta: myObjectMeta,
			Spec: triggersv1beta2.EventListenerSpec{
				Triggers: []triggersv1beta1.EventListenerTrigger{{
					Bindings: []*triggersv1beta1.EventListenerBinding{{
						Ref:        "tb",
						Kind:       triggersv1beta1.NamespacedTriggerBindingKind,
						APIVersion: "triggersv1beta1", // TODO: APIVersions seem wrong?
					}},
					Template: &triggersv1beta1.EventListenerTemplate{
						Ref: ptr.String("tt"),
					},
				}, {
					Template: &triggersv1beta1.EventListenerTemplate{
						Ref: ptr.String("tt2"),
					},
					Bindings: []*triggersv1beta1.EventListenerBinding{{
						Ref:        "tb",
						Kind:       triggersv1beta1.NamespacedTriggerBindingKind,
						APIVersion: "triggersv1beta1", // TODO: APIVersions seem wrong?
					}},
				}},
			},
		}},
	}, {
		name: "Valid EventListener with CEL interceptor",
		el: &EventListenerKubernetesResource{triggersv1beta2.EventListener{
			ObjectMeta: myObjectMeta,
			Spec: triggersv1beta2.EventListenerSpec{
				Triggers: []triggersv1beta1.EventListenerTrigger{{
					Interceptors: []*triggersv1beta1.EventInterceptor{{
						Ref: triggersv1beta1.InterceptorRef{
							Name:       "cel",
							Kind:       triggersv1beta1.ClusterInterceptorKind,
							APIVersion: "triggers.tekton.dev/triggersv1beta1",
						},
						Params: []triggersv1beta1.InterceptorParams{{
							Name:  "filter",
							Value: test.ToV1JSON(t, "body.value == test"),
						}, {
							Name: "overlays",
							Value: test.ToV1JSON(t, []triggersv1beta1.CELOverlay{{
								Key:        "value",
								Expression: "testing",
							}}),
						}},
					}},
					Bindings: []*triggersv1beta1.EventListenerBinding{{
						Ref:        "tb",
						Kind:       triggersv1beta1.NamespacedTriggerBindingKind,
						APIVersion: "triggersv1beta1", // TODO: APIVersions seem wrong?
					}},
					Template: &triggersv1beta1.EventListenerTemplate{
						Ref: ptr.String("tt"),
					},
				}},
			},
		}},
	}, {
		name: "Valid EventListener with no trigger name",
		el: &EventListenerKubernetesResource{triggersv1beta2.EventListener{
			ObjectMeta: myObjectMeta,
			Spec: triggersv1beta2.EventListenerSpec{
				Triggers: []triggersv1beta1.EventListenerTrigger{{
					Bindings: []*triggersv1beta1.EventListenerBinding{{
						Ref:        "tb",
						Kind:       triggersv1beta1.NamespacedTriggerBindingKind,
						APIVersion: "triggersv1beta1", // TODO: APIVersions seem wrong?
					}},
					Template: &triggersv1beta1.EventListenerTemplate{
						Ref: ptr.String("tt"),
					},
				}},
			},
		}},
	}, {
		name: "Namespace selector with label selector",
		el: &EventListenerKubernetesResource{triggersv1beta2.EventListener{
			ObjectMeta: myObjectMeta,
			Spec: triggersv1beta2.EventListenerSpec{
				NamespaceSelector: triggersv1beta1.NamespaceSelector{
					MatchNames: []string{"foo"},
				},
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"foo": "bar",
					},
				},
				Triggers: []triggersv1beta1.EventListenerTrigger{{
					Bindings: []*triggersv1beta1.EventListenerBinding{{Ref: "bindingRef", Kind: triggersv1beta1.NamespacedTriggerBindingKind}},
					Template: &triggersv1beta1.EventListenerTemplate{Ref: ptr.String("tt")},
				}},
			},
		}},
	}, {
		name: "Valid EventListener with kubernetes env for podspec",
		el: &EventListenerKubernetesResource{triggersv1beta2.EventListener{
			ObjectMeta: myObjectMeta,
			Spec: triggersv1beta2.EventListenerSpec{
				Triggers: []triggersv1beta1.EventListenerTrigger{{
					Template: &triggersv1beta1.EventListenerTemplate{
						Ref: ptr.String("tt"),
					},
				}},
				Resources: createValidKubernetesResource(),
			},
		}},
	}, {
		name: "Invalid EventListener with too many containers",
		el: &EventListenerKubernetesResource{triggersv1beta2.EventListener{
			ObjectMeta: myObjectMeta,
			Spec: triggersv1beta2.EventListenerSpec{
				Triggers: []triggersv1beta1.EventListenerTrigger{{
					Template: &triggersv1beta1.EventListenerTemplate{
						Ref: ptr.String("tt"),
					},
				}},
				Resources: createInvalidKubernetesResourceTooManyContainers(),
			},
		}},
		// TODO(vaikas): This seems like wrong error message here?
		wantErr: apis.ErrMultipleOneOf("spec.template.spec.containers"),
	}, {
		name: "Valid Replicas for EventListener",
		el: &EventListenerKubernetesResource{triggersv1beta2.EventListener{
			ObjectMeta: myObjectMeta,
			Spec: triggersv1beta2.EventListenerSpec{
				Resources: createValidReplicas(),
			},
		},
		},
	}, {
		name: "Invalid Replicas for EventListener",
		el: &EventListenerKubernetesResource{triggersv1beta2.EventListener{
			ObjectMeta: myObjectMeta,
			Spec: triggersv1beta2.EventListenerSpec{
				Resources: createInvalidReplicas(),
			},
		}},
		wantErr: apis.ErrInvalidValue("-1", "spec.replicas"),
	}, {
		name: "Valid EventListener with env for TLS connection",
		el: &EventListenerKubernetesResource{triggersv1beta2.EventListener{
			ObjectMeta: myObjectMeta,
			Spec: triggersv1beta2.EventListenerSpec{
				Triggers: []triggersv1beta1.EventListenerTrigger{{
					Template: &triggersv1beta1.EventListenerTemplate{
						Ref: ptr.String("tt"),
					},
				}},
				Resources: createWithTLS(),
			},
		}},
	}, {
		name: "Invalid EventListener with env for TLS connection missing one",
		el: &EventListenerKubernetesResource{triggersv1beta2.EventListener{
			ObjectMeta: myObjectMeta,
			Spec: triggersv1beta2.EventListenerSpec{
				Triggers: []triggersv1beta1.EventListenerTrigger{{
					Template: &triggersv1beta1.EventListenerTemplate{
						Ref: ptr.String("tt"),
					},
				}},
				Resources: createWithInvalidTLS(),
			},
		}},
		wantErr: apis.ErrGeneric("Expected env's are TLS_CERT and TLS_KEY, but got only one env TLS_CERT"),
	}, {
		name: "EventListener with invalid custom resources is ignored",
		el: &EventListenerKubernetesResource{triggersv1beta2.EventListener{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "name",
				Namespace: "namespace",
			},
			Spec: triggersv1beta2.EventListenerSpec{
				Triggers: []triggersv1beta1.EventListenerTrigger{{
					TriggerRef: "triggerref",
				}},
				Resources: createInvalidCustomResource(),
			},
		}},
	}, {
		name: "EventListener with invalid k8s resource fails",
		el: &EventListenerKubernetesResource{triggersv1beta2.EventListener{
			ObjectMeta: myObjectMeta,
			Spec: triggersv1beta2.EventListenerSpec{
				Triggers: []triggersv1beta1.EventListenerTrigger{{
					TriggerRef: "triggerref",
				}},
				Resources: createInvalidCustomResource(),
			},
		}},
		wantErr: apis.ErrGeneric("failed to decode resources", "spec.resources"),
	},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.el.Validate(context.Background())
			if diff := cmp.Diff(tc.wantErr.Error(), got.Error()); diff != "" {
				t.Error("Broker.CheckImmutableFields (-want, +got) =", diff)
			}
		})
	}
}

func TestImmutableFields(t *testing.T) {
	original := &triggersv1beta2.EventListener{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{"triggers.tekton.dev/eventlistener.class": "original"},
		},
	}
	current := &EventListenerKubernetesResource{triggersv1beta2.EventListener{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{"triggers.tekton.dev/eventlistener.class": "KubernetesResource"},
		},
	}}
	currentRealEL := &triggersv1beta2.EventListener{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{"triggers.tekton.dev/eventlistener.class": "KubernetesResource"},
		},
	}

	tests := map[string]struct {
		og      *triggersv1beta2.EventListener
		wantErr *apis.FieldError
	}{
		"nil original": {
			wantErr: nil,
		},
		"no EventListenerClassAnnotation mutation": {
			og:      currentRealEL,
			wantErr: nil,
		},
		"EventListenerClassAnnotation mutated": {
			og: original,
			wantErr: &apis.FieldError{
				Message: "Immutable fields changed (-old +new)",
				Paths:   []string{"annotations"},
				Details: `{string}:
	-: "original"
	+: "KubernetesResource"
`,
			},
		},
	}

	for n, test := range tests {
		t.Run(n, func(t *testing.T) {
			ctx := context.Background()
			ctx = apis.WithinUpdate(ctx, test.og)
			got := current.Validate(ctx)
			if diff := cmp.Diff(test.wantErr.Error(), got.Error()); diff != "" {
				t.Error("Broker.CheckImmutableFields (-want, +got) =", diff)
			}
		})
	}
}

func TestValidateFunc(t *testing.T) {
	tests := []struct {
		name     string
		el       *unstructured.Unstructured
		original *triggersv1beta2.EventListener
		want     *apis.FieldError
	}{{
		name: "not my broker class",
		el:   createOtherEventListener(),
	}, {
		name: "Invalid replicas",
		el:   createEventListenerResourceInvalidReplicas(),
		want: apis.ErrInvalidValue(-1, "spec.replicas"),
	}, {
		name: "valid replicas",
		el:   createEventListenerResourceValidReplicas(),
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			if test.original != nil {
				ctx = apis.WithinUpdate(ctx, test.original)
			}
			got := ValidateFunc(ctx, test.el)
			if test.want.Error() != "" && got == nil {
				t.Errorf("EventListener.Validate want: %q got nil", test.want.Error())
			} else if test.want.Error() == "" && got != nil {
				t.Errorf("EventListener.Validate want: nil got %q", got)
			} else if got != nil {
				if diff := cmp.Diff(test.want.Error(), got.Error()); diff != "" {
					t.Error("EventListener.Validate (-want, +got) =", diff)
				}
			}
		})
	}
}

// krToRawExtension is a helper to construct runtime.RawExtension from
// KubernetesResource. This will panic if json.Marshal fails, which is fine
// for test code.
func krToRawExtension(kr *KubernetesResource) *runtime.RawExtension {
	out, err := json.Marshal(kr)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal test resource: %v", err))
	}
	return &runtime.RawExtension{
		Raw: []byte(out),
	}
}

func createValidKubernetesResource() *runtime.RawExtension {
	return krToRawExtension(&KubernetesResource{
		WithPodSpec: duckv1.WithPodSpec{
			Template: duckv1.PodSpecable{
				Spec: corev1.PodSpec{
					ServiceAccountName: "k8sresource",
					Tolerations: []corev1.Toleration{{
						Key:      "key",
						Operator: "Equal",
						Value:    "value",
						Effect:   "NoSchedule",
					}},
					NodeSelector: map[string]string{"beta.kubernetes.io/os": "linux"},
					Containers: []corev1.Container{{
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.Quantity{Format: resource.DecimalSI},
								corev1.ResourceMemory: resource.Quantity{Format: resource.BinarySI},
							},
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.Quantity{Format: resource.DecimalSI},
								corev1.ResourceMemory: resource.Quantity{Format: resource.BinarySI},
							},
						},
					}},
				},
			},
		},
	})
}

func createInvalidKubernetesResourceTooManyContainers() *runtime.RawExtension {
	return krToRawExtension(&KubernetesResource{
		WithPodSpec: duckv1.WithPodSpec{
			Template: duckv1.PodSpecable{
				Spec: corev1.PodSpec{
					ServiceAccountName: "k8sresource",
					Tolerations: []corev1.Toleration{{
						Key:      "key",
						Operator: "Equal",
						Value:    "value",
						Effect:   "NoSchedule",
					}},
					NodeSelector: map[string]string{"beta.kubernetes.io/os": "linux"},
					Containers: []corev1.Container{{
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.Quantity{Format: resource.DecimalSI},
								corev1.ResourceMemory: resource.Quantity{Format: resource.BinarySI},
							},
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.Quantity{Format: resource.DecimalSI},
								corev1.ResourceMemory: resource.Quantity{Format: resource.BinarySI},
							},
						},
					}, {
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.Quantity{Format: resource.DecimalSI},
								corev1.ResourceMemory: resource.Quantity{Format: resource.BinarySI},
							},
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.Quantity{Format: resource.DecimalSI},
								corev1.ResourceMemory: resource.Quantity{Format: resource.BinarySI},
							},
						},
					}},
				},
			},
		},
	})
}

func createWithTLS() *runtime.RawExtension {
	return krToRawExtension(&KubernetesResource{
		WithPodSpec: duckv1.WithPodSpec{
			Template: duckv1.PodSpecable{
				Spec: corev1.PodSpec{
					ServiceAccountName: "k8sresource",
					Containers: []corev1.Container{{
						Env: []corev1.EnvVar{{
							Name: "TLS_CERT",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{Name: "secret-name"},
									Key:                  "tls.crt",
								},
							},
						}, {
							Name: "TLS_KEY",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{Name: "secret-name"},
									Key:                  "tls.key",
								},
							},
						}},
					}},
				},
			},
		},
	})
}

func createWithInvalidTLS() *runtime.RawExtension {
	return krToRawExtension(&KubernetesResource{
		WithPodSpec: duckv1.WithPodSpec{
			Template: duckv1.PodSpecable{
				Spec: corev1.PodSpec{
					ServiceAccountName: "k8sresource",
					Containers: []corev1.Container{{
						Env: []corev1.EnvVar{{
							Name: "TLS_CERT",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{Name: "secret-name"},
									Key:                  "tls.crt",
								},
							},
						}},
					}},
				},
			},
		},
	})
}

func createValidReplicas() *runtime.RawExtension {
	return krToRawExtension(&KubernetesResource{Replicas: ptr.Int32(1)})
}

func createInvalidReplicas() *runtime.RawExtension {
	return krToRawExtension(&KubernetesResource{Replicas: ptr.Int32(-1)})
}

// This creates invalid JSON string. It's used to validate that invalid
// resource that's not annotated with my class will be ignored.
func createInvalidCustomResource() *runtime.RawExtension {
	return &runtime.RawExtension{
		Raw: []byte(`some garbage here`),
	}
}

func createOtherEventListener() *unstructured.Unstructured {
	annotations := map[string]interface{}{
		"triggers.tekton.dev/eventlistener.class": "NotKubernetesResource",
	}

	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "triggers.tekton.dev/v1beta2",
			"kind":       "EventListener",
			"metadata": map[string]interface{}{
				"creationTimestamp": nil,
				"namespace":         "namespace",
				"name":              "el",
				"annotations":       annotations,
			},
			"spec": map[string]interface{}{
				"config": map[string]interface{}{
					"namespace": "namespace",
					"name":      "name",
				},
			},
		},
	}
}

func createEventListenerResourceValidReplicas() *unstructured.Unstructured {
	annotations := map[string]interface{}{
		"triggers.tekton.dev/eventlistener.class": "KubernetesResource",
	}

	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "triggers.tekton.dev/v1beta2",
			"kind":       "EventListener",
			"metadata": map[string]interface{}{
				"creationTimestamp": nil,
				"namespace":         "namespace",
				"name":              "el",
				"annotations":       annotations,
			},
			"spec": map[string]interface{}{
				"resources": map[string]interface{}{
					"replicas": 1,
				},
			},
		},
	}
}

func createEventListenerResourceInvalidReplicas() *unstructured.Unstructured {
	annotations := map[string]interface{}{
		"triggers.tekton.dev/eventlistener.class": "KubernetesResource",
	}

	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "triggers.tekton.dev/v1beta2",
			"kind":       "EventListener",
			"metadata": map[string]interface{}{
				"creationTimestamp": nil,
				"namespace":         "namespace",
				"name":              "el",
				"annotations":       annotations,
			},
			"spec": map[string]interface{}{
				"resources": map[string]interface{}{
					"replicas": -1,
				},
			},
		},
	}
}
