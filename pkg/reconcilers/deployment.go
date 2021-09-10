package reconcilers

import (
	"fmt"
	"reflect"

	"github.com/3scale/apicast-operator/pkg/k8sutils"

	appsv1 "k8s.io/api/apps/v1"
)

// DeploymentMutateFn is a function which mutates the existing Deployment into it's desired state.
type DeploymentMutateFn func(desired, existing *appsv1.Deployment) bool

func DeploymentMutator(opts ...DeploymentMutateFn) MutateFn {
	return func(existingObj, desiredObj k8sutils.KubernetesObject) (bool, error) {
		existing, ok := existingObj.(*appsv1.Deployment)
		if !ok {
			return false, fmt.Errorf("%T is not a *appsv1.Deployment", existingObj)
		}
		desired, ok := desiredObj.(*appsv1.Deployment)
		if !ok {
			return false, fmt.Errorf("%T is not a *appsv1.Deployment", desiredObj)
		}

		update := false

		// Loop through each option
		for _, opt := range opts {
			tmpUpdate := opt(desired, existing)
			update = update || tmpUpdate
		}

		return update, nil
	}
}

func DeploymentReplicasMutator(desired, existing *appsv1.Deployment) bool {
	update := false

	if desired.Spec.Replicas != existing.Spec.Replicas {
		existing.Spec.Replicas = desired.Spec.Replicas
		update = true
	}

	return update
}

func DeploymentImageMutator(desired, existing *appsv1.Deployment) bool {
	update := false

	if existing.Spec.Template.Spec.Containers[0].Image != desired.Spec.Template.Spec.Containers[0].Image {
		existing.Spec.Template.Spec.Containers[0].Image = desired.Spec.Template.Spec.Containers[0].Image
		update = true
	}

	return update
}

func DeploymentServiceAccountNameMutator(desired, existing *appsv1.Deployment) bool {
	update := false

	if existing.Spec.Template.Spec.ServiceAccountName != desired.Spec.Template.Spec.ServiceAccountName {
		update = true
		existing.Spec.Template.Spec.ServiceAccountName = desired.Spec.Template.Spec.ServiceAccountName
	}

	return update
}

func DeploymentEnvVarsMutator(desired, existing *appsv1.Deployment) bool {
	return ReconcileEnvVar(&existing.Spec.Template.Spec.Containers[0].Env, desired.Spec.Template.Spec.Containers[0].Env)
}

func DeploymentResourceMutator(desired, existing *appsv1.Deployment) bool {
	desiredName := k8sutils.ObjectInfo(desired)
	update := false

	//
	// Check container resource requirements
	//
	if len(desired.Spec.Template.Spec.Containers) != 1 {
		panic(fmt.Sprintf("%s desired spec.template.spec.containers length changed to '%d', should be 1", desiredName, len(desired.Spec.Template.Spec.Containers)))
	}

	if len(existing.Spec.Template.Spec.Containers) != 1 {
		existing.Spec.Template.Spec.Containers = desired.Spec.Template.Spec.Containers
		update = true
	}

	if !k8sutils.CmpResources(&existing.Spec.Template.Spec.Containers[0].Resources, &desired.Spec.Template.Spec.Containers[0].Resources) {
		existing.Spec.Template.Spec.Containers[0].Resources = desired.Spec.Template.Spec.Containers[0].Resources
		update = true
	}

	return update
}

func DeploymentPodTemplateAnnotationsMutator(desired, existing *appsv1.Deployment) bool {
	update := false

	// They are annotations of the PodTemplate, part of the Spec, not part of the meta info of the Pod or Environment object itself
	// It is not expected any controller to update them, so we use "set" approach, instead of merge.
	// This way any removed annotation from desired (due to change in CR) will be removed in existing too.
	if !reflect.DeepEqual(existing.Spec.Template.Annotations, desired.Spec.Template.Annotations) {
		update = true
		existing.Spec.Template.Annotations = desired.Spec.Template.Annotations
	}

	return update
}

func DeploymentVolumesMutator(desired, existing *appsv1.Deployment) bool {
	update := false

	if !reflect.DeepEqual(existing.Spec.Template.Spec.Volumes, desired.Spec.Template.Spec.Volumes) {
		update = true
		existing.Spec.Template.Spec.Volumes = desired.Spec.Template.Spec.Volumes
	}

	return update
}

func DeploymentVolumeMountsMutator(desired, existing *appsv1.Deployment) bool {
	update := false

	if !reflect.DeepEqual(existing.Spec.Template.Spec.Containers[0].VolumeMounts, desired.Spec.Template.Spec.Containers[0].VolumeMounts) {
		update = true
		existing.Spec.Template.Spec.Containers[0].VolumeMounts = desired.Spec.Template.Spec.Containers[0].VolumeMounts
	}

	return update
}

func DeploymentPortsMutator(desired, existing *appsv1.Deployment) bool {
	update := false

	if !reflect.DeepEqual(existing.Spec.Template.Spec.Containers[0].Ports, desired.Spec.Template.Spec.Containers[0].Ports) {
		update = true
		existing.Spec.Template.Spec.Containers[0].Ports = desired.Spec.Template.Spec.Containers[0].Ports
	}

	return update
}
