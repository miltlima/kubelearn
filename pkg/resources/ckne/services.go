// Package ckne holds the networking questions (CKNE track).
package ckne

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"strings"

	"kubelearn/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const netlab = "netlab"

// expectWebPort80 checks that the service selects the web pods on port 80.
func expectWebPort80(report *utils.Report, service *corev1.Service) {
	app := service.Spec.Selector["app"]
	report.Expect(app == "web", "Selector app is %s, expected web.", utils.Shown(app))
	report.Expect(len(service.Spec.Ports) > 0 && service.Spec.Ports[0].Port == 80, "The service's first port should be 80.")
}

func NodePortService(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	service, err := clientset.CoreV1().Services(netlab).Get(ctx, "web-nodeport", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "Service", "web-nodeport", netlab) {
		report.Expect(service.Spec.Type == corev1.ServiceTypeNodePort, "Service type is %s, expected NodePort.", utils.Shown(string(service.Spec.Type)))
		expectWebPort80(&report, service)
		report.Expect(len(service.Spec.Ports) > 0 && service.Spec.Ports[0].NodePort == 30080, "The node port should be 30080.")
	}
	return report.Result("Question 38 - Expose deployment web in namespace netlab with NodePort service web-nodeport on port 80 and node port 30080", "Easy")
}

func HeadlessService(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	service, err := clientset.CoreV1().Services(netlab).Get(ctx, "web-headless", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "Service", "web-headless", netlab) {
		report.Expect(service.Spec.ClusterIP == corev1.ClusterIPNone, "Cluster IP is %s, expected None.", utils.Shown(service.Spec.ClusterIP))
		expectWebPort80(&report, service)
	}
	return report.Result("Question 39 - Create headless service web-headless in namespace netlab for pods app=web on port 80", "Easy")
}

func ExternalNameService(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	service, err := clientset.CoreV1().Services(netlab).Get(ctx, "external-api", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "Service", "external-api", netlab) {
		report.Expect(service.Spec.Type == corev1.ServiceTypeExternalName, "Service type is %s, expected ExternalName.", utils.Shown(string(service.Spec.Type)))
		report.Expect(service.Spec.ExternalName == "api.example.com", "External name is %s, expected api.example.com.", utils.Shown(service.Spec.ExternalName))
	}
	return report.Result("Question 40 - Create service external-api in namespace netlab of type ExternalName pointing to api.example.com", "Easy")
}

func SessionAffinityService(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	service, err := clientset.CoreV1().Services(netlab).Get(ctx, "web-sticky", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "Service", "web-sticky", netlab) {
		expectWebPort80(&report, service)
		report.Expect(service.Spec.SessionAffinity == corev1.ServiceAffinityClientIP, "Session affinity is %s, expected ClientIP.", utils.Shown(string(service.Spec.SessionAffinity)))
		var timeout *int32
		if cfg := service.Spec.SessionAffinityConfig; cfg != nil && cfg.ClientIP != nil {
			timeout = cfg.ClientIP.TimeoutSeconds
		}
		report.Expect(utils.PtrEquals(timeout, 600), "Session affinity timeout is %s, expected 600 seconds.", utils.ShownPtr(timeout))
	}
	return report.Result("Question 41 - Create service web-sticky in namespace netlab for pods app=web on port 80 with client IP session affinity and a 600 second timeout", "Medium")
}

func LocalTrafficPolicyService(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	service, err := clientset.CoreV1().Services(netlab).Get(ctx, "web-local", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "Service", "web-local", netlab) {
		report.Expect(service.Spec.Type == corev1.ServiceTypeNodePort, "Service type is %s, expected NodePort.", utils.Shown(string(service.Spec.Type)))
		expectWebPort80(&report, service)
		policy := service.Spec.ExternalTrafficPolicy
		report.Expect(policy == corev1.ServiceExternalTrafficPolicyLocal, "External traffic policy is %s, expected Local.", utils.Shown(string(policy)))
	}
	return report.Result("Question 42 - Create NodePort service web-local in namespace netlab for pods app=web on port 80 that sends external traffic only to pods on the node that received it", "Medium")
}

func ReadinessProbe(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	deployment, err := clientset.AppsV1().Deployments(netlab).Get(ctx, "web", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "Deployment", "web", netlab) {
		container := utils.FirstContainer(deployment.Spec.Template.Spec)
		if report.Expect(container != nil && container.ReadinessProbe != nil, "The first container has no readiness probe.") {
			httpGet := container.ReadinessProbe.HTTPGet
			report.Expect(httpGet != nil && httpGet.Path == "/" && httpGet.Port.IntVal == 80, "The readiness probe should be an HTTP GET on path / and port 80.")
		}
	}
	return report.Result("Question 43 - Add an HTTP readiness probe on path / and port 80 to deployment web in namespace netlab, so only ready pods receive traffic", "Medium")
}

func CoreDNSRewrite(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	configMap, err := clientset.CoreV1().ConfigMaps("kube-system").Get(ctx, "coredns", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "ConfigMap", "coredns", "kube-system") {
		report.Expect(hasRewrite(configMap.Data["Corefile"], "legacy.internal", "web.netlab.svc.cluster.local"),
			"The Corefile has no rewrite rule from legacy.internal to web.netlab.svc.cluster.local.")
	}
	return report.Result("Question 44 - Customize CoreDNS with a rewrite rule so that legacy.internal resolves to service web in namespace netlab", "Hard")
}

// hasRewrite reports whether a Corefile has a rewrite rule mapping from to to.
func hasRewrite(corefile, from, to string) bool {
	for _, line := range strings.Split(corefile, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 3 || fields[0] != "rewrite" {
			continue
		}
		fromAt, toAt := -1, -1
		for i, field := range fields {
			switch field {
			case from:
				fromAt = i
			case to:
				toAt = i
			}
		}
		if fromAt > 0 && toAt > fromAt {
			return true
		}
	}
	return false
}

func PodDNSConfig(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	pod, err := clientset.CoreV1().Pods(netlab).Get(ctx, "dns-custom", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "Pod", "dns-custom", netlab) {
		report.ExpectImage(pod.Spec, "nginx:alpine")
		report.Expect(pod.Spec.DNSPolicy == corev1.DNSNone, "DNS policy is %s, expected None.", utils.Shown(string(pod.Spec.DNSPolicy)))
		if config := pod.Spec.DNSConfig; report.Expect(config != nil, "The pod has no DNS config.") {
			report.Expect(utils.Contains(config.Nameservers, "1.1.1.1"), "Nameserver 1.1.1.1 is missing.")
			report.Expect(utils.Contains(config.Searches, "netlab.svc.cluster.local"), "Search domain netlab.svc.cluster.local is missing.")
			ndots := false
			for _, option := range config.Options {
				ndots = ndots || (option.Name == "ndots" && utils.PtrEquals(option.Value, "2"))
			}
			report.Expect(ndots, "Option ndots should be 2.")
		}
	}
	return report.Result("Question 45 - Create pod dns-custom in namespace netlab with image nginx:alpine that ignores cluster DNS and uses nameserver 1.1.1.1, search domain netlab.svc.cluster.local and option ndots 2", "Medium")
}

func hasPolicyType(policy *networkingv1.NetworkPolicy, want networkingv1.PolicyType) bool {
	for _, policyType := range policy.Spec.PolicyTypes {
		if policyType == want {
			return true
		}
	}
	return false
}

func selectsEveryPod(selector metav1.LabelSelector) bool {
	return len(selector.MatchLabels) == 0 && len(selector.MatchExpressions) == 0
}

// allowsTCPPort reports whether the rule ports include the TCP port.
func allowsTCPPort(ports []networkingv1.NetworkPolicyPort, want int32) bool {
	for _, port := range ports {
		tcp := port.Protocol == nil || *port.Protocol == corev1.ProtocolTCP
		if tcp && port.Port != nil && port.Port.IntVal == want {
			return true
		}
	}
	return false
}

func DefaultDenyIngress(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	policy, err := clientset.NetworkingV1().NetworkPolicies(netlab).Get(ctx, "default-deny-ingress", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "NetworkPolicy", "default-deny-ingress", netlab) {
		report.Expect(selectsEveryPod(policy.Spec.PodSelector), "The policy should select every pod in the namespace.")
		report.Expect(hasPolicyType(policy, networkingv1.PolicyTypeIngress), "The policy types should include Ingress.")
		report.Expect(len(policy.Spec.Ingress) == 0, "The policy has ingress rules, so it still allows some traffic.")
	}
	return report.Result("Question 46 - Create network policy default-deny-ingress in namespace netlab that blocks all incoming traffic to every pod", "Easy")
}

func IngressFromNamespace(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	policy, err := clientset.NetworkingV1().NetworkPolicies(netlab).Get(ctx, "api-from-frontend", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "NetworkPolicy", "api-from-frontend", netlab) {
		app := policy.Spec.PodSelector.MatchLabels["app"]
		report.Expect(app == "api", "The policy selects pods app=%s, expected app=api.", utils.Shown(app))

		allowed := false
		for _, rule := range policy.Spec.Ingress {
			for _, peer := range rule.From {
				fromTeam := peer.NamespaceSelector != nil && peer.NamespaceSelector.MatchLabels["team"] == "frontend"
				allowed = allowed || (fromTeam && allowsTCPPort(rule.Ports, 8080))
			}
		}
		report.Expect(allowed, "No ingress rule allows namespaces labelled team=frontend on TCP port 8080.")
	}
	return report.Result("Question 47 - Create network policy api-from-frontend in namespace netlab so pods app=api accept TCP 8080 only from namespaces labelled team=frontend", "Hard")
}

func EgressExceptPrivate(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	policy, err := clientset.NetworkingV1().NetworkPolicies(netlab).Get(ctx, "web-egress-internet", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "NetworkPolicy", "web-egress-internet", netlab) {
		app := policy.Spec.PodSelector.MatchLabels["app"]
		report.Expect(app == "web", "The policy selects pods app=%s, expected app=web.", utils.Shown(app))
		report.Expect(hasPolicyType(policy, networkingv1.PolicyTypeEgress), "The policy types should include Egress.")

		allowed := false
		for _, rule := range policy.Spec.Egress {
			for _, peer := range rule.To {
				block := peer.IPBlock
				allowed = allowed || (block != nil && block.CIDR == "0.0.0.0/0" &&
					utils.Contains(block.Except, "10.0.0.0/8") && allowsTCPPort(rule.Ports, 443))
			}
		}
		report.Expect(allowed, "No egress rule allows 0.0.0.0/0 except 10.0.0.0/8 on TCP port 443.")
	}
	return report.Result("Question 48 - Create network policy web-egress-internet in namespace netlab so pods app=web can reach any address on TCP 443 except 10.0.0.0/8", "Hard")
}

func TLSSecret(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	secret, err := clientset.CoreV1().Secrets(netlab).Get(ctx, "shop-tls", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "Secret", "shop-tls", netlab) {
		report.Expect(secret.Type == corev1.SecretTypeTLS, "Secret type is %s, expected kubernetes.io/tls.", utils.Shown(string(secret.Type)))
		report.Expect(len(secret.Data[corev1.TLSPrivateKeyKey]) > 0, "The secret has no tls.key.")
		report.Expect(certificateIsFor(secret.Data[corev1.TLSCertKey], "shop.example.com"), "tls.crt is not a certificate for shop.example.com.")
	}
	return report.Result("Question 49 - Create TLS secret shop-tls in namespace netlab with a self-signed certificate for shop.example.com", "Medium")
}

// certificateIsFor reports whether the PEM certificate names host, either in
// its subject alternative names or in its common name.
func certificateIsFor(certPEM []byte, host string) bool {
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return false
	}
	certificate, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false
	}
	return certificate.VerifyHostname(host) == nil || certificate.Subject.CommonName == host
}

func IngressWithTLS(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	ingress, err := clientset.NetworkingV1().Ingresses(netlab).Get(ctx, "shop", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "Ingress", "shop", netlab) {
		report.Expect(utils.PtrEquals(ingress.Spec.IngressClassName, "nginx"), "Ingress class is %s, expected nginx.", utils.ShownPtr(ingress.Spec.IngressClassName))

		routed := false
		for _, rule := range ingress.Spec.Rules {
			if rule.Host != "shop.example.com" || rule.HTTP == nil {
				continue
			}
			for _, path := range rule.HTTP.Paths {
				service := path.Backend.Service
				routed = routed || (path.Path == "/" && service != nil && service.Name == "web" && service.Port.Number == 80)
			}
		}
		report.Expect(routed, "No rule sends host shop.example.com path / to service web on port 80.")

		tls := false
		for _, entry := range ingress.Spec.TLS {
			tls = tls || (entry.SecretName == "shop-tls" && utils.Contains(entry.Hosts, "shop.example.com"))
		}
		report.Expect(tls, "No TLS entry uses secret shop-tls for host shop.example.com.")
	}
	return report.Result("Question 50 - Create ingress shop in namespace netlab with ingress class nginx that sends host shop.example.com path / to service web on port 80, with TLS from secret shop-tls", "Medium")
}
