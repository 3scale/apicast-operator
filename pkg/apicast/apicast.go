package apicast

import (
	"strconv"
	"strings"

	appsv1alpha1 "github.com/3scale/apicast-operator/pkg/apis/apps/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	AdminPortalURLAttributeName = "AdminPortalURL"
)

const (
	EmbeddedConfigurationMountPath  = "/tmp/gateway-configuration-volume"
	EmbeddedConfigurationVolumeName = "gateway-configuration-volume"
	EmbeddedConfigurationSecretKey  = "config.json"
)

type APIcast struct {
	options *APIcastOptions
	scheme  *runtime.Scheme
}

func NewAPIcast(opts *APIcastOptions, scheme *runtime.Scheme) *APIcast {
	return &APIcast{options: opts, scheme: scheme}
}

func Factory(cr *appsv1alpha1.APIcast, cl client.Client, scheme *runtime.Scheme) (*APIcast, error) {
	optsProvider := NewApicastOptionsProvider(cr, cl)
	opts, err := optsProvider.GetApicastOptions()
	if err != nil {
		return nil, err
	}
	return NewAPIcast(opts, scheme), nil
}

func (a *APIcast) deploymentVolumeMounts() []v1.VolumeMount {
	var volumeMounts []v1.VolumeMount
	if a.options.GatewayConfigurationSecret != nil {
		volumeMounts = append(volumeMounts, v1.VolumeMount{
			Name:      EmbeddedConfigurationVolumeName,
			MountPath: EmbeddedConfigurationMountPath,
			ReadOnly:  true,
		})
	}

	return volumeMounts
}

func (a *APIcast) deploymentVolumes() []v1.Volume {
	var volumes []v1.Volume
	if a.options.GatewayConfigurationSecret != nil {
		volumes = append(volumes, v1.Volume{
			Name: EmbeddedConfigurationVolumeName,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: a.options.GatewayConfigurationSecret.Name,
					Items: []v1.KeyToPath{
						v1.KeyToPath{
							Key:  EmbeddedConfigurationSecretKey,
							Path: EmbeddedConfigurationSecretKey,
						},
					},
				},
			},
		})
	}

	return volumes
}

func (a *APIcast) envVarFromValue(name string, value string) v1.EnvVar {
	return v1.EnvVar{
		Name:  name,
		Value: value,
	}
}

func (a *APIcast) envVarFromSecretKey(name string, secretName string, secretKey string) v1.EnvVar {
	return v1.EnvVar{
		Name: name,
		ValueFrom: &v1.EnvVarSource{
			SecretKeyRef: &v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{
					Name: secretName,
				},
				Key: secretKey,
			},
		},
	}
}

func (a *APIcast) deploymentEnv() []v1.EnvVar {
	var env []v1.EnvVar

	if a.options.AdminPortalCredentialsSecret != nil {
		env = append(env, a.envVarFromSecretKey("THREESCALE_PORTAL_ENDPOINT", a.options.AdminPortalCredentialsSecret.Name, AdminPortalURLAttributeName))
	}

	if a.options.DeploymentEnvironment != nil {
		env = append(env, a.envVarFromValue("THREESCALE_DEPLOYMENT_ENV", *a.options.DeploymentEnvironment))
	}

	if a.options.DNSResolverAddress != nil {
		env = append(env, a.envVarFromValue("RESOLVER", *a.options.DNSResolverAddress))
	}

	if a.options.EnabledServices != nil {
		joinedStr := strings.Join(a.options.EnabledServices, ",")
		if joinedStr != "" {
			env = append(env, a.envVarFromValue("APICAST_SERVICES_LIST", joinedStr))
		}
	}

	if a.options.ConfigurationLoadMode != nil {
		env = append(env, a.envVarFromValue("APICAST_CONFIGURATION_LOADER", *a.options.ConfigurationLoadMode))
	}

	if a.options.LogLevel != nil {
		env = append(env, a.envVarFromValue("APICAST_LOG_LEVEL", *a.options.LogLevel))
	}

	if a.options.PathRoutingEnabled != nil {
		env = append(env, a.envVarFromValue("APICAST_PATH_ROUTING", strconv.FormatBool(*a.options.PathRoutingEnabled)))
	}

	if a.options.ResponseCodesIncluded != nil {
		env = append(env, a.envVarFromValue("APICAST_RESPONSE_CODES", strconv.FormatBool(*a.options.ResponseCodesIncluded)))
	}

	if a.options.CacheConfigurationSeconds != nil {
		env = append(env, a.envVarFromValue("APICAST_CONFIGURATION_CACHE", strconv.FormatInt(*a.options.CacheConfigurationSeconds, 10)))
	}

	if a.options.ManagementAPIScope != nil {
		env = append(env, a.envVarFromValue("APICAST_MANAGEMENT_API", *a.options.ManagementAPIScope))
	}

	if a.options.OpenSSLPeerVerificationEnabled != nil {
		env = append(env, a.envVarFromValue("OPENSSL_VERIFY", strconv.FormatBool(*a.options.OpenSSLPeerVerificationEnabled)))
	}

	if a.options.GatewayConfigurationSecret != nil {
		env = append(env, v1.EnvVar{
			Name:  "THREESCALE_CONFIG_FILE",
			Value: EmbeddedConfigurationMountPath + "/" + EmbeddedConfigurationSecretKey,
		})
	}

	return env
}

func (a *APIcast) Deployment() (*appsv1.Deployment, error) {
	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      a.options.DeploymentName,
			Namespace: a.options.Namespace,
			Labels:    a.commonLabels(),
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: a.deploymentLabelSelector(),
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      a.deploymentLabelSelector(),
					Annotations: a.podAnnotations(),
				},
				Spec: v1.PodSpec{
					ServiceAccountName: a.options.ServiceAccountName,
					Volumes:            a.deploymentVolumes(),
					Containers: []v1.Container{
						v1.Container{
							Name: a.options.DeploymentName,
							Ports: []v1.ContainerPort{
								v1.ContainerPort{Name: "proxy", ContainerPort: 8080, Protocol: v1.ProtocolTCP},
								v1.ContainerPort{Name: "management", ContainerPort: 8090, Protocol: v1.ProtocolTCP},
								v1.ContainerPort{Name: "metrics", ContainerPort: 9421, Protocol: v1.ProtocolTCP},
							},
							Image:           a.options.Image,
							ImagePullPolicy: v1.PullAlways, // This is different than the currently used which is IfNotPresent
							Resources: v1.ResourceRequirements{
								Limits: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("1"),
									v1.ResourceMemory: resource.MustParse("128Mi"),
								},
								Requests: v1.ResourceList{
									v1.ResourceCPU:    resource.MustParse("500m"),
									v1.ResourceMemory: resource.MustParse("64Mi"),
								},
							},
							LivenessProbe:  a.livenessProbe(),
							ReadinessProbe: a.readinessProbe(),
							VolumeMounts:   a.deploymentVolumeMounts(),
							// Env takes precedence with respect to EnvFrom on duplicated
							// var values
							Env: a.deploymentEnv(),
						},
					},
				},
			},
			Replicas: &a.options.Replicas,
		},
	}

	if err := controllerutil.SetControllerReference(a.options.Owner, deployment, a.scheme); err != nil {
		return nil, err
	}

	return deployment, nil
}

func (a *APIcast) deploymentLabelSelector() map[string]string {
	return map[string]string{
		"deployment": a.options.DeploymentName,
	}
}

func (a *APIcast) commonLabels() map[string]string {
	return map[string]string{
		"app":                  a.options.AppLabel,
		"threescale_component": "apicast",
	}
}

func (a *APIcast) podAnnotations() map[string]string {
	annotations := map[string]string{
		"prometheus.io/scrape": "true",
		"prometheus.io/port":   "9421",
	}

	for key, val := range a.options.AdditionalAnnotations {
		annotations[key] = val
	}

	return annotations
}

func (a *APIcast) Service() (*v1.Service, error) {
	service := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      a.options.ServiceName,
			Namespace: a.options.Namespace,
			Labels:    a.commonLabels(),
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{Name: "proxy", Port: 8080, Protocol: v1.ProtocolTCP, TargetPort: intstr.FromInt(8080)},
				v1.ServicePort{Name: "management", Port: 8090, Protocol: v1.ProtocolTCP, TargetPort: intstr.FromInt(8090)},
			},
			Selector: a.deploymentLabelSelector(),
		},
	}

	if err := controllerutil.SetControllerReference(a.options.Owner, service, a.scheme); err != nil {
		return nil, err
	}

	return service, nil
}

func (a *APIcast) livenessProbe() *v1.Probe {
	return &v1.Probe{
		Handler: v1.Handler{
			HTTPGet: &v1.HTTPGetAction{
				Path: "/status/live",
				Port: intstr.FromInt(8090),
			},
		},
		InitialDelaySeconds: 10,
		TimeoutSeconds:      5,
		PeriodSeconds:       10,
	}
}

func (a *APIcast) readinessProbe() *v1.Probe {
	return &v1.Probe{
		Handler: v1.Handler{
			HTTPGet: &v1.HTTPGetAction{
				Path: "/status/ready",
				Port: intstr.FromInt(8090),
			},
		},
		InitialDelaySeconds: 15,
		TimeoutSeconds:      5,
		PeriodSeconds:       30,
	}
}

func (a *APIcast) Ingress() (*extensions.Ingress, error) {
	ingress := &extensions.Ingress{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.k8s.io/v1beta1",
			Kind:       "Ingress",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      a.options.DeploymentName,
			Namespace: a.options.Namespace,
			Labels:    a.commonLabels(),
		},
		Spec: extensions.IngressSpec{
			TLS: a.options.ExposedHost.TLS,
			Rules: []extensions.IngressRule{
				{
					Host: a.options.ExposedHost.Host,
					IngressRuleValue: extensions.IngressRuleValue{
						HTTP: &extensions.HTTPIngressRuleValue{
							Paths: []extensions.HTTPIngressPath{
								{
									Backend: extensions.IngressBackend{
										ServiceName: a.options.DeploymentName,
										ServicePort: intstr.FromString("proxy"),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(a.options.Owner, ingress, a.scheme); err != nil {
		return nil, err
	}

	return ingress, nil
}

func (a *APIcast) AdminPortalCredentialsSecret() (*v1.Secret, error) {
	if a.options.AdminPortalCredentialsSecret == nil {
		return nil, nil
	}

	secret := a.options.AdminPortalCredentialsSecret

	if err := controllerutil.SetControllerReference(a.options.Owner, secret, a.scheme); err != nil {
		return nil, err
	}

	return secret, nil
}

func (a *APIcast) GatewayConfigurationSecret() (*v1.Secret, error) {
	if a.options.GatewayConfigurationSecret == nil {
		return nil, nil
	}

	secret := a.options.GatewayConfigurationSecret

	if err := controllerutil.SetControllerReference(a.options.Owner, secret, a.scheme); err != nil {
		return nil, err
	}

	return secret, nil
}
