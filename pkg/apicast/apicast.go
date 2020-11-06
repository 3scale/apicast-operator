package apicast

import (
	"fmt"
	"strconv"
	"strings"

	appsv1alpha1 "github.com/3scale/apicast-operator/apis/apps/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
}

func NewAPIcast(opts *APIcastOptions) *APIcast {
	return &APIcast{options: opts}
}

func Factory(cr *appsv1alpha1.APIcast, cl client.Client) (*APIcast, error) {
	optsProvider := NewApicastOptionsProvider(cr, cl)
	opts, err := optsProvider.GetApicastOptions()
	if err != nil {
		return nil, err
	}
	return NewAPIcast(opts), nil
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

	if a.options.UpstreamRetryCases != nil {
		env = append(env, a.envVarFromValue("APICAST_UPSTREAM_RETRY_CASES", *a.options.UpstreamRetryCases))
	}

	if a.options.CacheMaxTime != nil {
		env = append(env, a.envVarFromValue("APICAST_CACHE_MAX_TIME", *a.options.CacheMaxTime))
	}

	if a.options.CacheStatusCodes != nil {
		env = append(env, a.envVarFromValue("APICAST_CACHE_STATUS_CODES", *a.options.CacheStatusCodes))
	}

	if a.options.OidcLogLevel != nil {
		env = append(env, a.envVarFromValue("APICAST_OIDC_LOG_LEVEL", *a.options.OidcLogLevel))
	}

	if a.options.LoadServicesWhenNeeded != nil {
		env = append(env, a.envVarFromValue("APICAST_LOAD_SERVICES_WHEN_NEEDED", strconv.FormatBool(*a.options.LoadServicesWhenNeeded)))
	}

	if a.options.ServicesFilterByURL != nil {
		env = append(env, a.envVarFromValue("APICAST_SERVICES_FILTER_BY_URL", *a.options.ServicesFilterByURL))
	}

	for serviceID, serviceVersion := range a.options.ServiceConfigurationVersionOverride {
		env = append(env, a.envVarFromValue(fmt.Sprintf("APICAST_SERVICE_%s_CONFIGURATION_VERSION", serviceID), serviceVersion))
	}

	if a.options.HTTPSPort != nil {
		env = append(env, a.envVarFromValue("APICAST_HTTPS_PORT", strconv.Itoa(*a.options.HTTPSPort)))
	}

	return env
}

func (a *APIcast) Deployment() *appsv1.Deployment {
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
							Resources:       a.options.ResourceRequirements,
							LivenessProbe:   a.livenessProbe(),
							ReadinessProbe:  a.readinessProbe(),
							VolumeMounts:    a.deploymentVolumeMounts(),
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

	addOwnerRefToObject(deployment, *a.options.Owner)
	return deployment
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

func (a *APIcast) Service() *v1.Service {
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

	addOwnerRefToObject(service, *a.options.Owner)
	return service
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

func (a *APIcast) Ingress() *extensions.Ingress {
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

	addOwnerRefToObject(ingress, *a.options.Owner)
	return ingress
}

func (a *APIcast) AdminPortalCredentialsSecret() *v1.Secret {
	if a.options.AdminPortalCredentialsSecret == nil {
		return nil
	}

	secret := a.options.AdminPortalCredentialsSecret

	addOwnerRefToObject(secret, *a.options.Owner)
	return secret
}

func (a *APIcast) GatewayConfigurationSecret() *v1.Secret {
	if a.options.GatewayConfigurationSecret == nil {
		return nil
	}

	secret := a.options.GatewayConfigurationSecret

	addOwnerRefToObject(secret, *a.options.Owner)
	return secret
}

func addOwnerRefToObject(o metav1.Object, owner metav1.OwnerReference) {
	o.SetOwnerReferences(append(o.GetOwnerReferences(), owner))
}
