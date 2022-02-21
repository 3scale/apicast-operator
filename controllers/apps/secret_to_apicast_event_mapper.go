package controllers

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1alpha1 "github.com/3scale/apicast-operator/apis/apps/v1alpha1"
)

// SecretToApicastEventMapper is an EventHandler that maps secret object to apicast CR's
type SecretToApicastEventMapper struct {
	K8sClient client.Client
	Logger    logr.Logger
	Namespace string
}

func (s *SecretToApicastEventMapper) Map(obj client.Object) []reconcile.Request {
	apicastList := &appsv1alpha1.APIcastList{}

	// filter by Secret UID
	opts := []client.ListOption{client.HasLabels{apicastSecretLabelKey(string(obj.GetUID()))}}

	// Support namespace scope or cluster scoped
	if s.Namespace != "" {
		opts = append(opts, client.InNamespace(s.Namespace))
	}

	err := s.K8sClient.List(context.Background(), apicastList, opts...)
	if err != nil {
		s.Logger.Error(err, "reading apicast list")
		return nil
	}

	s.Logger.V(1).Info("Processing object", "key", client.ObjectKeyFromObject(obj), "accepted", len(apicastList.Items) > 0)

	requests := []reconcile.Request{}
	for idx := range apicastList.Items {
		requests = append(requests, reconcile.Request{NamespacedName: types.NamespacedName{
			Name:      apicastList.Items[idx].GetName(),
			Namespace: apicastList.Items[idx].GetNamespace(),
		}})
	}

	return requests
}
