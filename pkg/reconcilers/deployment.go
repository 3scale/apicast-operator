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

	var existingReplicas int32 = 1
	if existing.Spec.Replicas != nil {
		existingReplicas = *existing.Spec.Replicas
	}

	var desiredReplicas int32 = 1
	if desired.Spec.Replicas != nil {
		desiredReplicas = *desired.Spec.Replicas
	}

	if desiredReplicas != existingReplicas {
		existing.Spec.Replicas = &desiredReplicas
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

// DeploymentPodTemplateAnnotationsMutator ensures Pod Template Annotations is reconciled
func DeploymentPodTemplateAnnotationsMutator(desired, existing *appsv1.Deployment) bool {
	update := false

	// Use merge instead of set
	// See THREESCALE-11239
	k8sutils.MergeMapStringString(&update, &existing.Spec.Template.Annotations, desired.Spec.Template.Annotations)

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

func DeploymentTemplateLabelsMutator(desired, existing *appsv1.Deployment) bool {
	update := false

	k8sutils.MergeMapStringString(&update, &existing.Spec.Template.Labels, desired.Spec.Template.Labels)

	return update
}

func DeploymentAffinityMutator(desired, existing *appsv1.Deployment) bool {
	update := false
	if !reflect.DeepEqual(existing.Spec.Template.Spec.Affinity, desired.Spec.Template.Spec.Affinity) {
		existing.Spec.Template.Spec.Affinity = desired.Spec.Template.Spec.Affinity
		update = true
	}
	return update
}

func DeploymentTolerationsMutator(desired, existing *appsv1.Deployment) bool {
	updated := false

	if !reflect.DeepEqual(existing.Spec.Template.Spec.Tolerations, desired.Spec.Template.Spec.Tolerations) {
		existing.Spec.Template.Spec.Tolerations = desired.Spec.Template.Spec.Tolerations
		updated = true
	}

	return updated
}

// DeploymentTopologySpreadConstraintsMutator ensures TopologySpreadConstraints is reconciled
func DeploymentTopologySpreadConstraintsMutator(desired, existing *appsv1.Deployment) bool {
	updated := false

	if !reflect.DeepEqual(existing.Spec.Template.Spec.TopologySpreadConstraints, desired.Spec.Template.Spec.TopologySpreadConstraints) {
		existing.Spec.Template.Spec.TopologySpreadConstraints = desired.Spec.Template.Spec.TopologySpreadConstraints
		updated = true
	}

	return updated
}

// DeploymentPriorityClassMutator ensures priorityclass is reconciled
func DeploymentPriorityClassNameMutator(desired, existing *appsv1.Deployment) bool {
	updated := false

	if existing.Spec.Template.Spec.PriorityClassName != desired.Spec.Template.Spec.PriorityClassName {
		existing.Spec.Template.Spec.PriorityClassName = desired.Spec.Template.Spec.PriorityClassName
		updated = true
	}

	return updated
}
