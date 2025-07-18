package reconcilers

import (
	"testing"

	networkingv1 "k8s.io/api/networking/v1"
)

func TestIngressMutator(t *testing.T) {
	ingressClassName := "default"

	tests := []struct {
		name     string
		existing *networkingv1.Ingress
		desired  *networkingv1.Ingress
		expected bool
	}{
		{
			"test false when desired and existing are the same",
			&networkingv1.Ingress{
				Spec: networkingv1.IngressSpec{
					IngressClassName: &ingressClassName,
					Rules: []networkingv1.IngressRule{
						{
							Host: "foo",
						},
					},
					TLS: []networkingv1.IngressTLS{
						{Hosts: []string{"foo"}},
					},
				},
			},
			&networkingv1.Ingress{
				Spec: networkingv1.IngressSpec{
					IngressClassName: &ingressClassName,
					Rules: []networkingv1.IngressRule{
						{
							Host: "foo",
						},
					},
					TLS: []networkingv1.IngressTLS{
						{Hosts: []string{"foo"}},
					},
				},
			},
			false,
		},
		{
			"test true when desired and existing host do not match",
			&networkingv1.Ingress{
				Spec: networkingv1.IngressSpec{
					IngressClassName: &ingressClassName,
					Rules: []networkingv1.IngressRule{
						{
							Host: "foo",
						},
					},
				},
			},
			&networkingv1.Ingress{
				Spec: networkingv1.IngressSpec{
					IngressClassName: &ingressClassName,
					Rules: []networkingv1.IngressRule{
						{
							Host: "bar",
						},
					},
				},
			},
			true,
		},
		{
			"test true when desired and existing ingressClassName do not match",
			&networkingv1.Ingress{
				Spec: networkingv1.IngressSpec{
					IngressClassName: &ingressClassName,
					Rules: []networkingv1.IngressRule{
						{
							Host: "foo",
						},
					},
				},
			},
			&networkingv1.Ingress{
				Spec: networkingv1.IngressSpec{
					Rules: []networkingv1.IngressRule{
						{
							Host: "foo",
						},
					},
				},
			},
			true,
		},
		{
			"test true when desired and existing TLS do not match",
			&networkingv1.Ingress{
				Spec: networkingv1.IngressSpec{
					IngressClassName: &ingressClassName,
					Rules: []networkingv1.IngressRule{
						{
							Host: "foo",
						},
					},
					TLS: []networkingv1.IngressTLS{
						{Hosts: []string{"foo"}},
					},
				},
			},
			&networkingv1.Ingress{
				Spec: networkingv1.IngressSpec{
					Rules: []networkingv1.IngressRule{
						{
							Host: "foo",
						},
					},
					TLS: []networkingv1.IngressTLS{
						{Hosts: []string{"bar"}},
					},
				},
			},
			true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			changed, err := IngressMutator(tc.existing, tc.desired)
			if err != nil {
				t.Error("unexpected error: ", err)
			}
			if changed != tc.expected {
				t.Error("expected mutator return ", tc.expected, " but got: ", changed)
			}
		})
	}
}
