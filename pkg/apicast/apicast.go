package apicast

import (
	"strconv"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type APIcast struct {
	Namespace                        string
	DeploymentName                   string
	ServiceName                      string
	Replicas                         int32
	AppLabel                         string
	AdditionalAnnotations            map[string]string
	ServiceAccountName               string
	Image                            string
	ExposedHost                      ExposedHost
	OwnerReference                   *metav1.OwnerReference
	AdminPortalCredentialsSecretName *string

	DeploymentEnvironment          *string
	DNSResolverAddress             *string
	EnabledServices                []string
	ConfigurationLoadMode          *int64
	LogLevel                       *string
	PathRoutingEnabled             *bool
	ResponseCodesIncluded          *bool
	CacheConfigurationSeconds      *int64
	ManagementAPIScope             *string
	OpenSSLPeerVerificationEnabled *bool
	GatewayConfigurationSecretName *string
}

type ExposedHost struct {
	Host string
	TLS  []extensions.IngressTLS
}

const (
	AdminPortalURLAttributeName = "AdminPortalURL"
)

const (
	EmbeddedConfigurationMountPath  = "/tmp/gateway-configuration-volume"
	EmbeddedConfigurationVolumeName = "gateway-configuration-volume"
	EmbeddedConfigurationSecretKey  = "config.json"
)

func (a *APIcast) deploymentVolumeMounts() []v1.VolumeMount {
	var volumeMounts []v1.VolumeMount
	if a.GatewayConfigurationSecretName != nil {
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
	if a.GatewayConfigurationSecretName != nil {
		volumes = append(volumes, v1.Volume{
			Name: EmbeddedConfigurationVolumeName,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: *a.GatewayConfigurationSecretName,
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

	if a.AdminPortalCredentialsSecretName != nil {
		env = append(env, a.envVarFromSecretKey("THREESCALE_PORTAL_ENDPOINT", *a.AdminPortalCredentialsSecretName, AdminPortalURLAttributeName))
	}

	if a.DeploymentEnvironment != nil {
		env = append(env, a.envVarFromValue("THREESCALE_DEPLOYMENT_ENV", *a.DeploymentEnvironment))
	}

	if a.DNSResolverAddress != nil {
		env = append(env, a.envVarFromValue("RESOLVER", *a.DNSResolverAddress))
	}

	if a.EnabledServices != nil {
		joinedStr := strings.Join(a.EnabledServices, ",")
		if joinedStr != "" {
			env = append(env, a.envVarFromValue("APICAST_SERVICES_LIST", joinedStr))
		}
	}

	if a.ConfigurationLoadMode != nil {
		env = append(env, a.envVarFromValue("APICAST_CONFIGURATION_LOADER", string(*a.ConfigurationLoadMode)))
	}

	if a.LogLevel != nil {
		env = append(env, a.envVarFromValue("APICAST_LOG_LEVEL", *a.LogLevel))
	}

	if a.PathRoutingEnabled != nil {
		env = append(env, a.envVarFromValue("APICAST_PATH_ROUTING", strconv.FormatBool(*a.PathRoutingEnabled)))
	}

	if a.ResponseCodesIncluded != nil {
		env = append(env, a.envVarFromValue("APICAST_RESPONSE_CODES", strconv.FormatBool(*a.ResponseCodesIncluded)))
	}

	if a.CacheConfigurationSeconds != nil {
		env = append(env, a.envVarFromValue("APICAST_CONFIGURATION_CACHE", string(*a.CacheConfigurationSeconds)))
	}

	if a.ManagementAPIScope != nil {
		env = append(env, a.envVarFromValue("APICAST_MANAGEMENT_API", *a.ManagementAPIScope))
	}

	if a.OpenSSLPeerVerificationEnabled != nil {
		env = append(env, a.envVarFromValue("OPENSSL_VERIFY", strconv.FormatBool(*a.OpenSSLPeerVerificationEnabled)))
	}

	if a.GatewayConfigurationSecretName != nil {
		env = append(env, v1.EnvVar{
			Name:  "THREESCALE_CONFIG_FILE",
			Value: EmbeddedConfigurationMountPath + "/" + EmbeddedConfigurationSecretKey,
		})
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
			Name:      a.DeploymentName,
			Namespace: a.Namespace,
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
					ServiceAccountName: a.ServiceAccountName,
					Volumes:            a.deploymentVolumes(),
					Containers: []v1.Container{
						v1.Container{
							Name: a.DeploymentName,
							Ports: []v1.ContainerPort{
								v1.ContainerPort{Name: "proxy", ContainerPort: 8080, Protocol: v1.ProtocolTCP},
								v1.ContainerPort{Name: "management", ContainerPort: 8090, Protocol: v1.ProtocolTCP},
								v1.ContainerPort{Name: "metrics", ContainerPort: 9421, Protocol: v1.ProtocolTCP},
							},
							Image:           a.Image,
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
			Replicas: &a.Replicas, // TODO set to nil?
		},
	}
	if a.OwnerReference != nil {
		addOwnerRefToObject(deployment, *a.OwnerReference)
	}

	return deployment
}

func (a *APIcast) deploymentLabelSelector() map[string]string {
	return map[string]string{
		"deployment": a.DeploymentName,
	}
}

func (a *APIcast) commonLabels() map[string]string {
	return map[string]string{
		"app":                  a.AppLabel,
		"threescale_component": "apicast",
	}
}

func (a *APIcast) podAnnotations() map[string]string {
	annotations := map[string]string{
		"prometheus.io/scrape": "true",
		"prometheus.io/port":   "9421",
	}

	for key, val := range a.AdditionalAnnotations {
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
			Name:      a.ServiceName,
			Namespace: a.Namespace,
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

	if a.OwnerReference != nil {
		addOwnerRefToObject(service, *a.OwnerReference)
	}

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
			Name:      a.DeploymentName,
			Namespace: a.Namespace,
			Labels:    a.commonLabels(),
		},
		Spec: extensions.IngressSpec{
			TLS: a.ExposedHost.TLS,
			Rules: []extensions.IngressRule{
				{
					Host: a.ExposedHost.Host,
					IngressRuleValue: extensions.IngressRuleValue{
						HTTP: &extensions.HTTPIngressRuleValue{
							Paths: []extensions.HTTPIngressPath{
								{
									Backend: extensions.IngressBackend{
										ServiceName: a.DeploymentName,
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

	if a.OwnerReference != nil {
		addOwnerRefToObject(ingress, *a.OwnerReference)
	}

	return ingress
}

func addOwnerRefToObject(o metav1.Object, r metav1.OwnerReference) {
	o.SetOwnerReferences(append(o.GetOwnerReferences(), r))
}
