package resources_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	nodev1 "k8s.io/api/node/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"

	"kubelearn/pkg/resources"
	"kubelearn/pkg/utils"
)

type check struct {
	name string
	fn   resources.Check
}

// allChecks names every registered check after its question number.
var allChecks = func() []check {
	checks := make([]check, 0, len(resources.Questions))
	for i, question := range resources.Questions {
		checks = append(checks, check{fmt.Sprintf("q%d", i+1), question.Check})
	}
	return checks
}()

// gatewayResources maps Gateway API kinds to their resource names. The fake
// dynamic client would otherwise guess "gatewaies" for Gateway.
var gatewayResources = map[string]string{"Gateway": "gateways", "HTTPRoute": "httproutes"}

// clients builds an in-memory cluster holding objects. Custom resources
// (unstructured objects) go to the dynamic client, the rest to the typed one.
func clients(objects ...runtime.Object) resources.Clients {
	dynamicClient := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme())
	var typed []runtime.Object
	for _, object := range objects {
		custom, ok := object.(*unstructured.Unstructured)
		if !ok {
			typed = append(typed, object)
			continue
		}
		resource := custom.GroupVersionKind().GroupVersion().WithResource(gatewayResources[custom.GetKind()])
		if err := dynamicClient.Tracker().Create(resource, custom.DeepCopy(), custom.GetNamespace()); err != nil {
			panic(err)
		}
	}
	return resources.Clients{Kube: fake.NewSimpleClientset(typed...), Dynamic: dynamicClient}
}

// runCheck runs one checker and converts a panic into a test failure.
func runCheck(t *testing.T, c check, cs resources.Clients) (passed bool) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("%s panicked: %v", c.name, r)
		}
	}()
	return c.fn(context.Background(), cs).Passed
}

func ptr[T any](v T) *T { return &v }

func meta(namespace, name string) metav1.ObjectMeta {
	return metav1.ObjectMeta{Namespace: namespace, Name: name}
}

func podSpec(image string) corev1.PodSpec {
	return corev1.PodSpec{Containers: []corev1.Container{{Name: "main", Image: image}}}
}

func deployment(namespace, name string, replicas int32, spec corev1.PodSpec) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: meta(namespace, name),
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr(replicas),
			Template: corev1.PodTemplateSpec{Spec: spec},
		},
	}
}

// solvedCluster returns objects that satisfy every question.
func solvedCluster() []runtime.Object {
	mark42 := podSpec("nginx:alpine")
	mark42.ServiceAccountName = "america-sa"
	mark42.Containers[0].SecurityContext = &corev1.SecurityContext{AllowPrivilegeEscalation: ptr(false)}

	webserver := podSpec("nginx:alpine")
	webserver.Containers[0].VolumeMounts = []corev1.VolumeMount{{Name: "unicorn-pv", MountPath: "/usr/share/nginx/html"}}
	webserver.Volumes = []corev1.Volume{{
		Name: "unicorn-pv",
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: "unicorn-pvc"},
		},
	}}

	purple := podSpec("redis:alpine")
	purple.Volumes = []corev1.Volume{{
		Name:         "secret",
		VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{SecretName: "secret-purple"}},
	}}

	mark50 := podSpec("nginx:alpine")
	mark50.Containers[0].LivenessProbe = &corev1.Probe{
		InitialDelaySeconds: 5,
		PeriodSeconds:       10,
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{Path: "/", Port: intstr.FromInt(80)},
		},
	}

	sleeper := podSpec("busybox:1.28")
	sleeper.Containers[0].Command = []string{"sleep 3600"}
	cronSleeper := *sleeper.DeepCopy()
	cronSleeper.RestartPolicy = corev1.RestartPolicyNever

	pathType := networkingv1.PathTypePrefix

	objects := []runtime.Object{
		&corev1.Pod{ObjectMeta: meta("default", "nginx"), Spec: podSpec("nginx:alpine")},
		deployment("default", "nginx-deployment", 4, podSpec("nginx:alpine")),
		deployment("latam", "redis", 1, podSpec("redis:alpine")),
		&corev1.Service{
			ObjectMeta: meta("latam", "redis-service"),
			Spec:       corev1.ServiceSpec{Ports: []corev1.ServicePort{{Port: 6379}}},
		},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "europe"}},
		&corev1.ConfigMap{ObjectMeta: meta("default", "europe-configmap"), Data: map[string]string{"France": "Paris"}},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Namespace: "asia", Name: "tshoot", Labels: map[string]string{"country": "china"}},
			Spec:       podSpec("amazon/amazon-ecs-network-sidecar:latest"),
		},
		&corev1.PersistentVolume{
			ObjectMeta: metav1.ObjectMeta{Name: "unicorn-pv"},
			Spec: corev1.PersistentVolumeSpec{
				Capacity:    corev1.ResourceList{corev1.ResourceStorage: resource.MustParse("1Gi")},
				AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
				PersistentVolumeSource: corev1.PersistentVolumeSource{
					HostPath: &corev1.HostPathVolumeSource{Path: "/tmp/data"},
				},
			},
		},
		&corev1.PersistentVolumeClaim{
			ObjectMeta: meta("default", "unicorn-pvc"),
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{corev1.ResourceStorage: resource.MustParse("400Mi")},
				},
			},
		},
		&corev1.Pod{ObjectMeta: meta("public", "webserver"), Spec: webserver},
		&corev1.Pod{ObjectMeta: meta("bandai", "gundamv"), Spec: podSpec("nginx:alpine")},
		&networkingv1.NetworkPolicy{
			ObjectMeta: meta("colors", "allow-policy-colors"),
			Spec: networkingv1.NetworkPolicySpec{
				Ingress: []networkingv1.NetworkPolicyIngressRule{{
					From: []networkingv1.NetworkPolicyPeer{{
						PodSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"tier": "frontend"}},
					}},
					Ports: []networkingv1.NetworkPolicyPort{{Port: ptr(intstr.FromInt(6379))}},
				}},
			},
		},
		&corev1.Secret{ObjectMeta: meta("colors", "secret-colors"), Data: map[string][]byte{"color": []byte("red")}},
		&corev1.Pod{ObjectMeta: meta("colors", "purple"), Spec: purple},
		&corev1.Secret{ObjectMeta: meta("colors", "secret-purple"), Data: map[string][]byte{"singer": []byte("prince")}},
		&corev1.ServiceAccount{ObjectMeta: meta("default", "america-sa")},
		deployment("default", "mark42", 5, mark42),
		&autoscalingv2.HorizontalPodAutoscaler{
			ObjectMeta: meta("default", "mark43"),
			Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{Kind: "Deployment", Name: "mark43"},
				MinReplicas:    ptr(int32(2)),
				MaxReplicas:    8,
				Metrics: []autoscalingv2.MetricSpec{{
					Type: autoscalingv2.ResourceMetricSourceType,
					Resource: &autoscalingv2.ResourceMetricSource{
						Name: corev1.ResourceCPU,
						Target: autoscalingv2.MetricTarget{
							Type:               autoscalingv2.UtilizationMetricType,
							AverageUtilization: ptr(int32(80)),
						},
					},
				}},
			},
		},
		&corev1.Pod{ObjectMeta: meta("shield", "mark50"), Spec: mark50},
		deployment("colors", "yellow-deployment", 2, podSpec("bonovoo/node-app:1.0")),
		&corev1.Service{
			ObjectMeta: meta("colors", "yellow-service"),
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{"app": "yellow-deployment"},
				Ports:    []corev1.ServicePort{{Port: 80, TargetPort: intstr.FromInt(3000)}},
			},
		},
		&networkingv1.Ingress{
			ObjectMeta: meta("colors", "ingress-colors"),
			Spec: networkingv1.IngressSpec{
				Rules: []networkingv1.IngressRule{{
					Host: "yellow.com",
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{{
								Path:     "/yellow",
								PathType: &pathType,
								Backend: networkingv1.IngressBackend{
									Service: &networkingv1.IngressServiceBackend{Name: "yellow-service"},
								},
							}},
						},
					},
				}},
			},
		},
		&rbacv1.Role{
			ObjectMeta: meta("fruits", "apple-one"),
			Rules:      []rbacv1.PolicyRule{{Resources: []string{"pods"}, Verbs: []string{"get", "list", "watch"}}},
		},
		&batchv1.Job{
			ObjectMeta: meta("default", "job-gain"),
			Spec: batchv1.JobSpec{
				Parallelism:           ptr(int32(2)),
				Completions:           ptr(int32(4)),
				BackoffLimit:          ptr(int32(3)),
				ActiveDeadlineSeconds: ptr(int64(40)),
			},
		},
		&batchv1.CronJob{
			ObjectMeta: meta("default", "cronjob-gain"),
			Spec: batchv1.CronJobSpec{
				Schedule: "*/5 * * * *",
				JobTemplate: batchv1.JobTemplateSpec{
					Spec: batchv1.JobSpec{Template: corev1.PodTemplateSpec{Spec: cronSleeper}},
				},
			},
		},
		&appsv1.StatefulSet{
			ObjectMeta: meta("default", "statefulset-gain"),
			Spec:       appsv1.StatefulSetSpec{Template: corev1.PodTemplateSpec{Spec: sleeper}},
			Status:     appsv1.StatefulSetStatus{ReadyReplicas: 3},
		},
	}
	return append(objects, append(solvedCKS(), solvedCKNE()...)...)
}

func serviceAccountSubject(name, namespace string) []rbacv1.Subject {
	return []rbacv1.Subject{{Kind: "ServiceAccount", Name: name, Namespace: namespace}}
}

// solvedCKS returns objects that satisfy the security questions (27-37).
func solvedCKS() []runtime.Object {
	seccomp := podSpec("nginx:alpine")
	seccomp.SecurityContext = &corev1.PodSecurityContext{
		SeccompProfile: &corev1.SeccompProfile{Type: corev1.SeccompProfileTypeRuntimeDefault},
	}

	strict := podSpec("nginx:alpine")
	strict.Containers[0].SecurityContext = &corev1.SecurityContext{
		RunAsNonRoot:           ptr(true),
		ReadOnlyRootFilesystem: ptr(true),
		Capabilities: &corev1.Capabilities{
			Drop: []corev1.Capability{"ALL"},
			Add:  []corev1.Capability{"NET_BIND_SERVICE"},
		},
	}

	sandboxed := podSpec("nginx:alpine")
	sandboxed.RuntimeClassName = ptr("gvisor")

	udp := corev1.ProtocolUDP
	tcp := corev1.ProtocolTCP

	return []runtime.Object{
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "baseline-labs", Labels: map[string]string{"pod-security.kubernetes.io/enforce": "baseline"}}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "restricted-labs", Labels: map[string]string{"pod-security.kubernetes.io/enforce": "restricted"}}},
		&corev1.Pod{ObjectMeta: meta("hardening", "seccomp-default"), Spec: seccomp},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace:   "hardening",
				Name:        "apparmor-default",
				Annotations: map[string]string{"container.apparmor.security.beta.kubernetes.io/main": "runtime/default"},
			},
			Spec: podSpec("nginx:alpine"),
		},
		&corev1.Pod{ObjectMeta: meta("hardening", "strict-security"), Spec: strict},
		&networkingv1.NetworkPolicy{
			ObjectMeta: meta("secure-egress", "restrict-egress"),
			Spec: networkingv1.NetworkPolicySpec{
				PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeEgress},
				Egress: []networkingv1.NetworkPolicyEgressRule{
					{Ports: []networkingv1.NetworkPolicyPort{{Protocol: &udp, Port: ptr(intstr.FromInt(53))}}},
					{
						To:    []networkingv1.NetworkPolicyPeer{{IPBlock: &networkingv1.IPBlock{CIDR: "10.10.0.0/16"}}},
						Ports: []networkingv1.NetworkPolicyPort{{Protocol: &tcp}},
					},
				},
			},
		},
		&corev1.ServiceAccount{ObjectMeta: meta("hardening", "no-token-sa"), AutomountServiceAccountToken: ptr(false)},
		&rbacv1.Role{
			ObjectMeta: meta("audit-labs", "secret-reader"),
			Rules:      []rbacv1.PolicyRule{{Resources: []string{"secrets"}, Verbs: []string{"get"}, ResourceNames: []string{"db-credentials"}}},
		},
		&rbacv1.RoleBinding{
			ObjectMeta: meta("audit-labs", "secret-reader-binding"),
			RoleRef:    rbacv1.RoleRef{Kind: "Role", Name: "secret-reader"},
			Subjects:   serviceAccountSubject("auditor", "audit-labs"),
		},
		&rbacv1.RoleBinding{
			ObjectMeta: meta("hardening", "dev-view"),
			RoleRef:    rbacv1.RoleRef{Kind: "ClusterRole", Name: "view"},
			Subjects:   serviceAccountSubject("dev", "hardening"),
		},
		deployment("hardening", "legacy-app", 1, podSpec("nginx:alpine")),
		&corev1.Secret{ObjectMeta: meta("hardening", "api-keys"), Data: map[string][]byte{"token": []byte("s3cr3t")}, Immutable: ptr(true)},
		&nodev1.RuntimeClass{ObjectMeta: metav1.ObjectMeta{Name: "gvisor"}, Handler: "runsc"},
		&corev1.Pod{ObjectMeta: meta("hardening", "sandboxed"), Spec: sandboxed},
	}
}

// selfSignedCertificate returns a PEM certificate for host.
func selfSignedCertificate(host string) []byte {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	template := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: host}, DNSNames: []string{host}}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}

func webService(name string, spec corev1.ServiceSpec) *corev1.Service {
	spec.Selector = map[string]string{"app": "web"}
	if len(spec.Ports) == 0 {
		spec.Ports = []corev1.ServicePort{{Port: 80}}
	}
	return &corev1.Service{ObjectMeta: meta("netlab", name), Spec: spec}
}

func gatewayObject(kind, name string, spec map[string]any) *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "gateway.networking.k8s.io/v1",
		"kind":       kind,
		"metadata":   map[string]any{"namespace": "netlab", "name": name},
		"spec":       spec,
	}}
}

func backendRef(name string, port int64, weight ...int64) map[string]any {
	ref := map[string]any{"name": name, "port": port}
	if len(weight) > 0 {
		ref["weight"] = weight[0]
	}
	return ref
}

// solvedCKNE returns objects that satisfy the networking questions (38-54).
func solvedCKNE() []runtime.Object {
	ready := podSpec("nginx:alpine")
	ready.Containers[0].ReadinessProbe = &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{HTTPGet: &corev1.HTTPGetAction{Path: "/", Port: intstr.FromInt(80)}},
	}

	dns := podSpec("nginx:alpine")
	dns.DNSPolicy = corev1.DNSNone
	dns.DNSConfig = &corev1.PodDNSConfig{
		Nameservers: []string{"1.1.1.1"},
		Searches:    []string{"netlab.svc.cluster.local"},
		Options:     []corev1.PodDNSConfigOption{{Name: "ndots", Value: ptr("2")}},
	}

	pathType := networkingv1.PathTypePrefix
	parent := []any{map[string]any{"name": "web-gateway"}}

	return []runtime.Object{
		webService("web-nodeport", corev1.ServiceSpec{Type: corev1.ServiceTypeNodePort, Ports: []corev1.ServicePort{{Port: 80, NodePort: 30080}}}),
		webService("web-headless", corev1.ServiceSpec{ClusterIP: corev1.ClusterIPNone}),
		&corev1.Service{
			ObjectMeta: meta("netlab", "external-api"),
			Spec:       corev1.ServiceSpec{Type: corev1.ServiceTypeExternalName, ExternalName: "api.example.com"},
		},
		webService("web-sticky", corev1.ServiceSpec{
			SessionAffinity:       corev1.ServiceAffinityClientIP,
			SessionAffinityConfig: &corev1.SessionAffinityConfig{ClientIP: &corev1.ClientIPConfig{TimeoutSeconds: ptr(int32(600))}},
		}),
		webService("web-local", corev1.ServiceSpec{Type: corev1.ServiceTypeNodePort, ExternalTrafficPolicy: corev1.ServiceExternalTrafficPolicyLocal}),
		deployment("netlab", "web", 2, ready),
		&corev1.ConfigMap{
			ObjectMeta: meta("kube-system", "coredns"),
			Data: map[string]string{"Corefile": `.:53 {
    errors
    rewrite name legacy.internal web.netlab.svc.cluster.local
    kubernetes cluster.local in-addr.arpa ip6.arpa
    forward . /etc/resolv.conf
}`},
		},
		&corev1.Pod{ObjectMeta: meta("netlab", "dns-custom"), Spec: dns},
		&networkingv1.NetworkPolicy{
			ObjectMeta: meta("netlab", "default-deny-ingress"),
			Spec:       networkingv1.NetworkPolicySpec{PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress}},
		},
		&networkingv1.NetworkPolicy{
			ObjectMeta: meta("netlab", "api-from-frontend"),
			Spec: networkingv1.NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{MatchLabels: map[string]string{"app": "api"}},
				Ingress: []networkingv1.NetworkPolicyIngressRule{{
					From:  []networkingv1.NetworkPolicyPeer{{NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"team": "frontend"}}}},
					Ports: []networkingv1.NetworkPolicyPort{{Port: ptr(intstr.FromInt(8080))}},
				}},
			},
		},
		&networkingv1.NetworkPolicy{
			ObjectMeta: meta("netlab", "web-egress-internet"),
			Spec: networkingv1.NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{MatchLabels: map[string]string{"app": "web"}},
				PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeEgress},
				Egress: []networkingv1.NetworkPolicyEgressRule{{
					To:    []networkingv1.NetworkPolicyPeer{{IPBlock: &networkingv1.IPBlock{CIDR: "0.0.0.0/0", Except: []string{"10.0.0.0/8"}}}},
					Ports: []networkingv1.NetworkPolicyPort{{Port: ptr(intstr.FromInt(443))}},
				}},
			},
		},
		&corev1.Secret{
			ObjectMeta: meta("netlab", "shop-tls"),
			Type:       corev1.SecretTypeTLS,
			Data:       map[string][]byte{"tls.crt": selfSignedCertificate("shop.example.com"), "tls.key": []byte("key")},
		},
		&networkingv1.Ingress{
			ObjectMeta: meta("netlab", "shop"),
			Spec: networkingv1.IngressSpec{
				IngressClassName: ptr("nginx"),
				TLS:              []networkingv1.IngressTLS{{Hosts: []string{"shop.example.com"}, SecretName: "shop-tls"}},
				Rules: []networkingv1.IngressRule{{
					Host: "shop.example.com",
					IngressRuleValue: networkingv1.IngressRuleValue{HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: []networkingv1.HTTPIngressPath{{
							Path:     "/",
							PathType: &pathType,
							Backend: networkingv1.IngressBackend{Service: &networkingv1.IngressServiceBackend{
								Name: "web", Port: networkingv1.ServiceBackendPort{Number: 80},
							}},
						}},
					}},
				}},
			},
		},
		gatewayObject("Gateway", "web-gateway", map[string]any{
			"gatewayClassName": "kubelearn",
			"listeners": []any{
				map[string]any{"name": "http", "protocol": "HTTP", "port": int64(80)},
				map[string]any{
					"name": "https", "protocol": "HTTPS", "port": int64(443), "hostname": "shop.example.com",
					"tls": map[string]any{"mode": "Terminate", "certificateRefs": []any{map[string]any{"name": "shop-tls"}}},
				},
			},
		}),
		gatewayObject("HTTPRoute", "web-route", map[string]any{
			"parentRefs": parent,
			"hostnames":  []any{"shop.example.com"},
			"rules": []any{
				map[string]any{
					"matches":     []any{map[string]any{"path": map[string]any{"type": "PathPrefix", "value": "/api"}}},
					"backendRefs": []any{backendRef("api", 8080)},
				},
				map[string]any{"backendRefs": []any{backendRef("web", 80)}},
			},
		}),
		gatewayObject("HTTPRoute", "web-canary", map[string]any{
			"parentRefs": parent,
			"hostnames":  []any{"canary.example.com"},
			"rules":      []any{map[string]any{"backendRefs": []any{backendRef("web", 80, 90), backendRef("web-v2", 80, 10)}}},
		}),
	}
}

// stripped returns the solved cluster with every object reduced according to
// level, keeping only names so the checkers still find the objects:
//
//	0: empty specs (no containers, volumes, ports, rules, replicas...)
//	1: one bare container / volume / port / rule with no nested fields set
func stripped(level int) []runtime.Object {
	bare := corev1.PodSpec{}
	if level == 1 {
		bare = corev1.PodSpec{
			Containers: []corev1.Container{{Name: "main", SecurityContext: &corev1.SecurityContext{}}},
			Volumes:    []corev1.Volume{{Name: "data"}},
		}
	}
	template := corev1.PodTemplateSpec{Spec: bare}

	var objects []runtime.Object
	for _, obj := range solvedCluster() {
		switch o := obj.(type) {
		case *corev1.Pod:
			o.Spec = *bare.DeepCopy()
		case *appsv1.Deployment:
			o.Spec = appsv1.DeploymentSpec{Template: *template.DeepCopy()}
		case *appsv1.StatefulSet:
			o.Spec = appsv1.StatefulSetSpec{Template: *template.DeepCopy()}
		case *batchv1.Job:
			o.Spec = batchv1.JobSpec{}
		case *batchv1.CronJob:
			o.Spec = batchv1.CronJobSpec{
				Schedule:    "*/5 * * * *",
				JobTemplate: batchv1.JobTemplateSpec{Spec: batchv1.JobSpec{Template: *template.DeepCopy()}},
			}
		case *corev1.Service:
			o.Spec = corev1.ServiceSpec{}
			if level == 1 {
				o.Spec.Ports = []corev1.ServicePort{{}}
			}
		case *corev1.PersistentVolume:
			o.Spec = corev1.PersistentVolumeSpec{}
			if level == 1 {
				o.Spec.AccessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany}
				o.Spec.Capacity = corev1.ResourceList{corev1.ResourceStorage: resource.MustParse("1Gi")}
			}
		case *corev1.PersistentVolumeClaim:
			o.Spec = corev1.PersistentVolumeClaimSpec{}
		case *networkingv1.NetworkPolicy:
			o.Spec = networkingv1.NetworkPolicySpec{}
			if level == 1 {
				o.Spec.Ingress = []networkingv1.NetworkPolicyIngressRule{{
					From:  []networkingv1.NetworkPolicyPeer{{}},
					Ports: []networkingv1.NetworkPolicyPort{{}},
				}}
			}
		case *networkingv1.Ingress:
			o.Spec = networkingv1.IngressSpec{}
			if level == 1 {
				o.Spec.Rules = []networkingv1.IngressRule{{Host: "yellow.com"}}
			}
		case *autoscalingv2.HorizontalPodAutoscaler:
			o.Spec = autoscalingv2.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{Name: "mark43"},
				MaxReplicas:    8,
			}
			if level == 1 {
				o.Spec.MinReplicas = ptr(int32(2))
				o.Spec.Metrics = []autoscalingv2.MetricSpec{{Type: autoscalingv2.PodsMetricSourceType}}
			}
		case *rbacv1.Role:
			o.Rules = nil
			if level == 1 {
				o.Rules = []rbacv1.PolicyRule{{}}
			}
		}
		objects = append(objects, obj)
	}
	return objects
}

func TestSolvedClusterPassesEveryQuestion(t *testing.T) {
	cs := clients(solvedCluster()...)
	for _, c := range allChecks {
		if !runCheck(t, c, cs) {
			t.Errorf("%s: expected Passed=true on a solved cluster", c.name)
		}
	}
}

func TestEmptyClusterFailsEveryQuestionWithoutPanic(t *testing.T) {
	cs := clients()
	for _, c := range allChecks {
		if runCheck(t, c, cs) {
			t.Errorf("%s: expected Passed=false on an empty cluster", c.name)
		}
	}
}

// Objects that exist but are incomplete must fail the question, never panic.
func TestIncompleteObjectsDoNotPanic(t *testing.T) {
	for level := 0; level <= 1; level++ {
		t.Run(fmt.Sprintf("level%d", level), func(t *testing.T) {
			cs := clients(stripped(level)...)
			for _, c := range allChecks {
				runCheck(t, c, cs)
			}
		})
	}
}

// Questions 3 and 13 read two objects; the second one existing must not hide
// that the first one is missing.
func TestSecondObjectAloneDoesNotPass(t *testing.T) {
	var objects []runtime.Object
	for _, obj := range solvedCluster() {
		switch o := obj.(type) {
		case *corev1.Service:
			if o.Name == "redis-service" {
				objects = append(objects, o)
			}
		case *corev1.Secret:
			if o.Name == "secret-purple" {
				objects = append(objects, o)
			}
		}
	}
	cs := clients(objects...)
	for _, c := range allChecks {
		if c.name != "q3" && c.name != "q13" {
			continue
		}
		if runCheck(t, c, cs) {
			t.Errorf("%s: expected Passed=false when only the second object exists", c.name)
		}
	}
}

// Regression: pod mark50 existing in namespace shield without a liveness probe
// (the state before the question is solved) used to crash /questions and
// /finish.
func TestLivenessProbeQuestionOnFreshLab(t *testing.T) {
	fresh := &corev1.Pod{ObjectMeta: meta("shield", "mark50"), Spec: podSpec("nginx:alpine")}
	if result(t, "q19", fresh).Passed {
		t.Errorf("q19: expected Passed=false for a pod without liveness probe")
	}
}

// The registry position defines the question number shown to the user.
func TestRegistryOrderMatchesQuestionNumbers(t *testing.T) {
	results := evaluate(t, clients())
	if len(results) != len(resources.Questions) {
		t.Fatalf("Evaluate returned %d results for %d checks", len(results), len(resources.Questions))
	}
	for i, result := range results {
		prefix := fmt.Sprintf("Question %d - ", i+1)
		if !strings.HasPrefix(result.TestName, prefix) {
			t.Errorf("position %d holds %q, want a name starting with %q", i+1, result.TestName, prefix)
		}
	}
}

func TestScore(t *testing.T) {
	if got := resources.Score(nil); got != 0 {
		t.Errorf("Score(nil) = %v, want 0", got)
	}
	if got := resources.Score(evaluate(t, clients())); got != 0 {
		t.Errorf("empty cluster score = %v, want 0", got)
	}
	if got := resources.Score(evaluate(t, clients(solvedCluster()...))); got != 100 {
		t.Errorf("solved cluster score = %v, want 100", got)
	}
	half := []utils.Result{{Passed: true}, {Passed: false}}
	if got := resources.Score(half); got != 50 {
		t.Errorf("half passed score = %v, want 50", got)
	}
}

func evaluate(t *testing.T, cs resources.Clients) []utils.Result {
	t.Helper()
	results, err := resources.Evaluate(context.Background(), cs, "")
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	return results
}

// A cancelled or expired context must surface as an error, not as a list of
// failed questions.
func TestEvaluateReportsContextErrors(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	results, err := resources.Evaluate(ctx, clients(solvedCluster()...), "")
	if !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want context.Canceled", err)
	}
	if results != nil {
		t.Errorf("got %d results from a cancelled evaluation, want none", len(results))
	}
}

// result runs the named question against a cluster holding objects.
func result(t *testing.T, name string, objects ...runtime.Object) utils.Result {
	t.Helper()
	cs := clients(objects...)
	for _, c := range allChecks {
		if c.name == name {
			return c.fn(context.Background(), cs)
		}
	}
	t.Fatalf("no check named %s", name)
	return utils.Result{}
}

// Every unsolved question must say why; a solved one must not.
func TestFeedbackIsPresentExactlyWhenFailing(t *testing.T) {
	for _, r := range evaluate(t, clients()) {
		if r.Feedback == "" {
			t.Errorf("%q failed on an empty cluster without feedback", r.TestName)
		}
	}
	for level := 0; level <= 1; level++ {
		for _, r := range evaluate(t, clients(stripped(level)...)) {
			if r.Passed != (r.Feedback == "") {
				t.Errorf("level %d: %q has Passed=%v with feedback %q", level, r.TestName, r.Passed, r.Feedback)
			}
		}
	}
	for _, r := range evaluate(t, clients(solvedCluster()...)) {
		if r.Feedback != "" {
			t.Errorf("%q passed but carries feedback %q", r.TestName, r.Feedback)
		}
	}
}

func TestFeedbackMessages(t *testing.T) {
	wrongImage := &corev1.Pod{ObjectMeta: meta("default", "nginx"), Spec: podSpec("nginx:latest")}
	noProbe := &corev1.Pod{ObjectMeta: meta("shield", "mark50"), Spec: podSpec("nginx:alpine")}
	oneReplica := deployment("default", "nginx-deployment", 1, podSpec("nginx:alpine"))
	readOnlyRole := &rbacv1.Role{
		ObjectMeta: meta("fruits", "apple-one"),
		Rules:      []rbacv1.PolicyRule{{Resources: []string{"pods"}, Verbs: []string{"get"}}},
	}

	cases := []struct {
		question string
		objects  []runtime.Object
		want     string
	}{
		{"q1", nil, "Pod nginx not found in namespace default."},
		{"q4", nil, "Namespace europe not found."},
		{"q1", []runtime.Object{wrongImage}, "Container image is nginx:latest, expected nginx:alpine."},
		{"q2", []runtime.Object{oneReplica}, "Replicas is 1, expected 4."},
		{"q19", []runtime.Object{noProbe}, "The first container has no liveness probe."},
		{"q23", []runtime.Object{readOnlyRole}, "Missing verbs: list, watch."},
		{"q3", nil, "Deployment redis not found in namespace latam. Service redis-service not found in namespace latam."},
	}
	for _, tc := range cases {
		if got := result(t, tc.question, tc.objects...).Feedback; got != tc.want {
			t.Errorf("%s: feedback %q, want %q", tc.question, got, tc.want)
		}
	}
}

// The troubleshooting question must not give the answer away.
func TestTroubleshootingFeedbackDoesNotRevealTheFix(t *testing.T) {
	broken := &corev1.Pod{ObjectMeta: meta("bandai", "gundamv"), Spec: podSpec("nginxxx:alpine")}
	feedback := result(t, "q10", broken).Feedback
	if feedback == "" || strings.Contains(feedback, "nginx") || strings.Contains(feedback, "image") {
		t.Errorf("feedback %q should hint without naming the image", feedback)
	}
}

// A request that fails for a reason other than "not found" is a cluster
// problem and must be reported as such.
func TestFeedbackReportsClusterErrors(t *testing.T) {
	kube := fake.NewSimpleClientset()
	kube.PrependReactor("get", "pods", func(k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("connection refused")
	})
	r := resources.Questions[0].Check(context.Background(), resources.Clients{Kube: kube, Dynamic: clients().Dynamic})
	if r.Passed || !strings.Contains(r.Feedback, "Could not read pod nginx from the cluster") ||
		!strings.Contains(r.Feedback, "connection refused") {
		t.Errorf("unexpected result for an unreachable cluster: %+v", r)
	}
}

// "sleep 3600" must be accepted however the command is split; the single
// string ["sleep 3600"] cannot even start in a real container.
func TestSleepCommandIsAcceptedInAnyForm(t *testing.T) {
	forms := [][2][]string{
		{{"sleep", "3600"}, nil},
		{{"sh", "-c", "sleep 3600"}, nil},
		{{"sleep"}, {"3600"}},
		{nil, {"sleep", "3600"}},
		{{"sleep 3600"}, nil},
	}
	for _, form := range forms {
		spec := podSpec("busybox:1.28")
		spec.Containers[0].Command, spec.Containers[0].Args = form[0], form[1]
		sts := &appsv1.StatefulSet{
			ObjectMeta: meta("default", "statefulset-gain"),
			Spec:       appsv1.StatefulSetSpec{Template: corev1.PodTemplateSpec{Spec: spec}},
			Status:     appsv1.StatefulSetStatus{ReadyReplicas: 3},
		}
		if r := result(t, "q26", sts); !r.Passed {
			t.Errorf("command %v args %v rejected: %s", form[0], form[1], r.Feedback)
		}
	}

	wrong := podSpec("busybox:1.28")
	wrong.Containers[0].Command = []string{"sleep", "60"}
	sts := &appsv1.StatefulSet{
		ObjectMeta: meta("default", "statefulset-gain"),
		Spec:       appsv1.StatefulSetSpec{Template: corev1.PodTemplateSpec{Spec: wrong}},
		Status:     appsv1.StatefulSetStatus{ReadyReplicas: 3},
	}
	if r := result(t, "q26", sts); r.Passed {
		t.Errorf("command sleep 60 was accepted")
	}
}

func TestTracks(t *testing.T) {
	want := map[string]int{resources.TrackCore: 26, resources.TrackCKS: 11, resources.TrackCKNE: 17}
	total := 0
	for track, count := range want {
		results, err := resources.Evaluate(context.Background(), clients(), track)
		if err != nil {
			t.Fatal(err)
		}
		if len(results) != count {
			t.Errorf("track %s has %d questions, want %d", track, len(results), count)
		}
		for _, r := range results {
			if r.Track != track {
				t.Errorf("%q is tagged %q in a %s evaluation", r.TestName, r.Track, track)
			}
		}
		total += count
	}
	if len(resources.Questions) != total {
		t.Errorf("registry has %d questions, the tracks add up to %d", len(resources.Questions), total)
	}

	for name, wantTrack := range map[string]string{"": "", "all": "", "cks": "CKS", "CKNE": "CKNE", "Core": "Core"} {
		if got, ok := resources.ParseTrack(name); !ok || got != wantTrack {
			t.Errorf("ParseTrack(%q) = %q, %v", name, got, ok)
		}
	}
	if _, ok := resources.ParseTrack("cka"); ok {
		t.Errorf("ParseTrack accepted an unknown track")
	}
}

// The catalog describes every question without a cluster and without results.
func TestCatalog(t *testing.T) {
	catalog := resources.Catalog()
	if len(catalog) != len(resources.Questions) {
		t.Fatalf("catalog has %d entries, want %d", len(catalog), len(resources.Questions))
	}
	for i, entry := range catalog {
		if entry.TestName == "" || entry.Difficulty == "" || entry.Track != resources.Questions[i].Track {
			t.Errorf("incomplete catalog entry %d: %+v", i+1, entry)
		}
		if entry.Passed || entry.Feedback != "" {
			t.Errorf("catalog entry %d carries a result: %+v", i+1, entry)
		}
	}
}

func TestSecurityAndNetworkFeedback(t *testing.T) {
	stillAdmin := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: "dev-admin"},
		RoleRef:    rbacv1.RoleRef{Kind: "ClusterRole", Name: "cluster-admin"},
		Subjects:   serviceAccountSubject("dev", "hardening"),
	}
	privileged := podSpec("nginx:alpine")
	privileged.HostPID = true
	privileged.Containers[0].SecurityContext = &corev1.SecurityContext{Privileged: ptr(true)}
	openEgress := &networkingv1.NetworkPolicy{
		ObjectMeta: meta("secure-egress", "restrict-egress"),
		Spec: networkingv1.NetworkPolicySpec{
			PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeEgress},
			Egress:      []networkingv1.NetworkPolicyEgressRule{{}},
		},
	}
	plainCoreDNS := &corev1.ConfigMap{ObjectMeta: meta("kube-system", "coredns"), Data: map[string]string{"Corefile": ".:53 {\n    forward . /etc/resolv.conf\n}"}}
	wrongCertificate := &corev1.Secret{
		ObjectMeta: meta("netlab", "shop-tls"),
		Type:       corev1.SecretTypeTLS,
		Data:       map[string][]byte{"tls.crt": selfSignedCertificate("other.example.com"), "tls.key": []byte("key")},
	}
	httpOnly := gatewayObject("Gateway", "web-gateway", map[string]any{
		"gatewayClassName": "kubelearn",
		"listeners":        []any{map[string]any{"name": "http", "protocol": "HTTP", "port": int64(80)}},
	})
	evenSplit := gatewayObject("HTTPRoute", "web-canary", map[string]any{
		"parentRefs": []any{map[string]any{"name": "web-gateway"}},
		"hostnames":  []any{"canary.example.com"},
		"rules":      []any{map[string]any{"backendRefs": []any{backendRef("web", 80, 50), backendRef("web-v2", 80, 50)}}},
	})

	cases := []struct {
		question string
		objects  []runtime.Object
		want     string
	}{
		{"q34", append(solvedCKS(), stillAdmin), "Cluster role binding dev-admin still gives service account dev the cluster-admin role."},
		{"q35", []runtime.Object{deployment("hardening", "legacy-app", 1, privileged)}, "Container main still runs privileged. The pods still share the host PID namespace."},
		{"q31", []runtime.Object{openEgress}, "No egress rule allows DNS on UDP port 53. No egress rule allows TCP to CIDR 10.10.0.0/16. An egress rule without destinations or ports allows all traffic."},
		{"q44", []runtime.Object{plainCoreDNS}, "The Corefile has no rewrite rule from legacy.internal to web.netlab.svc.cluster.local."},
		{"q49", []runtime.Object{wrongCertificate}, "tls.crt is not a certificate for shop.example.com."},
		{"q51", nil, "Gateway web-gateway not found in namespace netlab."},
		{"q54", []runtime.Object{httpOnly}, "No listener accepts HTTPS on port 443."},
		{"q53", []runtime.Object{evenSplit}, "No rule splits traffic 90 to service web and 10 to service web-v2 on port 80."},
	}
	for _, tc := range cases {
		if got := result(t, tc.question, tc.objects...).Feedback; got != tc.want {
			t.Errorf("%s: feedback %q, want %q", tc.question, got, tc.want)
		}
	}

	// The gateway with only the HTTP listener still solves question 51.
	if r := result(t, "q51", httpOnly); !r.Passed {
		t.Errorf("q51 rejected a valid HTTP gateway: %s", r.Feedback)
	}
}
