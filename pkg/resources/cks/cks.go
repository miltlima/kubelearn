// Package cks holds the security questions (CKS track).
package cks

import (
	"context"

	"kubelearn/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const hardening = "hardening"

func PodSecurityNamespaces(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	var report utils.Report
	for _, want := range []struct{ namespace, level string }{
		{"baseline-labs", "baseline"},
		{"restricted-labs", "restricted"},
	} {
		namespace, err := clientset.CoreV1().Namespaces().Get(ctx, want.namespace, metav1.GetOptions{})
		if report.Found(err, "Namespace", want.namespace, "") {
			level := namespace.Labels["pod-security.kubernetes.io/enforce"]
			report.Expect(level == want.level, "Namespace %s enforces %s, expected %s.", want.namespace, utils.Shown(level), want.level)
		}
	}
	return report.Result("Question 27 - Create namespaces baseline-labs and restricted-labs that enforce the baseline and the restricted Pod Security Standard", "Easy")
}

func SeccompProfile(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	pod, err := clientset.CoreV1().Pods(hardening).Get(ctx, "seccomp-default", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "Pod", "seccomp-default", hardening) {
		container := report.ExpectImage(pod.Spec, "nginx:alpine")
		profile := seccompType(pod.Spec, container)
		report.Expect(profile == corev1.SeccompProfileTypeRuntimeDefault, "Seccomp profile is %s, expected RuntimeDefault.", utils.Shown(string(profile)))
	}
	return report.Result("Question 28 - Create pod seccomp-default in namespace hardening with image nginx:alpine and the RuntimeDefault seccomp profile", "Medium")
}

// seccompType returns the seccomp profile of the container, falling back to
// the pod-level one.
func seccompType(spec corev1.PodSpec, container *corev1.Container) corev1.SeccompProfileType {
	if container != nil && container.SecurityContext != nil && container.SecurityContext.SeccompProfile != nil {
		return container.SecurityContext.SeccompProfile.Type
	}
	if spec.SecurityContext != nil && spec.SecurityContext.SeccompProfile != nil {
		return spec.SecurityContext.SeccompProfile.Type
	}
	return ""
}

func AppArmorProfile(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	pod, err := clientset.CoreV1().Pods(hardening).Get(ctx, "apparmor-default", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "Pod", "apparmor-default", hardening) {
		if container := report.ExpectImage(pod.Spec, "nginx:alpine"); container != nil {
			profile := pod.Annotations["container.apparmor.security.beta.kubernetes.io/"+container.Name]
			report.Expect(profile == "runtime/default", "AppArmor profile of container %s is %s, expected runtime/default.", container.Name, utils.Shown(profile))
		}
	}
	return report.Result("Question 29 - Create pod apparmor-default in namespace hardening with image nginx:alpine and the AppArmor profile runtime/default on its container", "Medium")
}

func StrictSecurityContext(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	pod, err := clientset.CoreV1().Pods(hardening).Get(ctx, "strict-security", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "Pod", "strict-security", hardening) {
		container := report.ExpectImage(pod.Spec, "nginx:alpine")
		var sc *corev1.SecurityContext
		if container != nil {
			sc = container.SecurityContext
		}
		psc := pod.Spec.SecurityContext

		nonRoot := sc != nil && (utils.PtrEquals(sc.RunAsNonRoot, true) || (sc.RunAsUser != nil && *sc.RunAsUser > 0))
		nonRoot = nonRoot || (psc != nil && (utils.PtrEquals(psc.RunAsNonRoot, true) || (psc.RunAsUser != nil && *psc.RunAsUser > 0)))
		report.Expect(nonRoot, "The pod may run as root: set runAsNonRoot or a non-zero runAsUser.")

		var drop, add []corev1.Capability
		if sc != nil && sc.Capabilities != nil {
			drop, add = sc.Capabilities.Drop, sc.Capabilities.Add
		}
		report.Expect(hasCapability(drop, "ALL"), "The container should drop ALL capabilities.")
		report.Expect(len(add) == 1 && add[0] == "NET_BIND_SERVICE", "The container should add only the NET_BIND_SERVICE capability.")
		report.Expect(sc != nil && utils.PtrEquals(sc.ReadOnlyRootFilesystem, true), "The container's root filesystem should be read-only.")
	}
	return report.Result("Question 30 - Create pod strict-security in namespace hardening with image nginx:alpine that runs as non-root, drops all capabilities, adds only NET_BIND_SERVICE and has a read-only root filesystem", "Hard")
}

func hasCapability(capabilities []corev1.Capability, want corev1.Capability) bool {
	for _, capability := range capabilities {
		if capability == want {
			return true
		}
	}
	return false
}

func RestrictEgress(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	policy, err := clientset.NetworkingV1().NetworkPolicies("secure-egress").Get(ctx, "restrict-egress", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "NetworkPolicy", "restrict-egress", "secure-egress") {
		selector := policy.Spec.PodSelector
		report.Expect(len(selector.MatchLabels) == 0 && len(selector.MatchExpressions) == 0, "The policy should select every pod in the namespace.")

		egressType := false
		for _, policyType := range policy.Spec.PolicyTypes {
			egressType = egressType || policyType == networkingv1.PolicyTypeEgress
		}
		report.Expect(egressType, "The policy types should include Egress.")

		dns, cidr, open := false, false, false
		for _, rule := range policy.Spec.Egress {
			open = open || (len(rule.To) == 0 && len(rule.Ports) == 0)
			for _, port := range rule.Ports {
				dns = dns || (port.Protocol != nil && *port.Protocol == corev1.ProtocolUDP && port.Port != nil && port.Port.IntVal == 53)
			}
			for _, peer := range rule.To {
				if peer.IPBlock != nil && peer.IPBlock.CIDR == "10.10.0.0/16" {
					tcpOnly := true
					for _, port := range rule.Ports {
						tcpOnly = tcpOnly && (port.Protocol == nil || *port.Protocol == corev1.ProtocolTCP)
					}
					cidr = cidr || tcpOnly
				}
			}
		}
		report.Expect(dns, "No egress rule allows DNS on UDP port 53.")
		report.Expect(cidr, "No egress rule allows TCP to CIDR 10.10.0.0/16.")
		report.Expect(!open, "An egress rule without destinations or ports allows all traffic.")
	}
	return report.Result("Question 31 - In namespace secure-egress create network policy restrict-egress for all pods that allows egress only to DNS on UDP port 53 and to CIDR 10.10.0.0/16 on TCP", "Hard")
}

func ServiceAccountWithoutToken(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	sa, err := clientset.CoreV1().ServiceAccounts(hardening).Get(ctx, "no-token-sa", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "ServiceAccount", "no-token-sa", hardening) {
		report.Expect(utils.PtrEquals(sa.AutomountServiceAccountToken, false),
			"automountServiceAccountToken is %s, expected false.", utils.ShownPtr(sa.AutomountServiceAccountToken))
	}
	return report.Result("Question 32 - Create service account no-token-sa in namespace hardening that does not automount its API token", "Easy")
}

func LeastPrivilegeRole(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	role, roleErr := clientset.RbacV1().Roles("audit-labs").Get(ctx, "secret-reader", metav1.GetOptions{})
	binding, bindingErr := clientset.RbacV1().RoleBindings("audit-labs").Get(ctx, "secret-reader-binding", metav1.GetOptions{})

	var report utils.Report
	if report.Found(roleErr, "Role", "secret-reader", "audit-labs") &&
		report.Expect(len(role.Rules) == 1, "The role should have exactly one rule, found %d.", len(role.Rules)) {
		rule := role.Rules[0]
		report.Expect(len(rule.Resources) == 1 && rule.Resources[0] == "secrets", "The rule should apply only to secrets.")
		report.Expect(len(rule.Verbs) == 1 && rule.Verbs[0] == "get", "The rule should allow only the get verb.")
		report.Expect(len(rule.ResourceNames) == 1 && rule.ResourceNames[0] == "db-credentials", "The rule should be limited to the secret db-credentials.")
	}
	if report.Found(bindingErr, "RoleBinding", "secret-reader-binding", "audit-labs") {
		report.Expect(binding.RoleRef.Kind == "Role" && binding.RoleRef.Name == "secret-reader", "The binding should reference role secret-reader.")
		report.Expect(bindsServiceAccount(binding.Subjects, "auditor", "audit-labs"), "The binding should include service account auditor.")
	}
	return report.Result("Question 33 - In namespace audit-labs create role secret-reader that can only get the secret db-credentials, and bind it to service account auditor with role binding secret-reader-binding", "Medium")
}

func bindsServiceAccount(subjects []rbacv1.Subject, name, namespace string) bool {
	for _, subject := range subjects {
		if subject.Kind == "ServiceAccount" && subject.Name == name && subject.Namespace == namespace {
			return true
		}
	}
	return false
}

func RemoveClusterAdmin(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	bindings, listErr := clientset.RbacV1().ClusterRoleBindings().List(ctx, metav1.ListOptions{})
	view, viewErr := clientset.RbacV1().RoleBindings(hardening).Get(ctx, "dev-view", metav1.GetOptions{})

	var report utils.Report
	if report.Expect(listErr == nil, "Could not list cluster role bindings from the cluster: %v.", listErr) {
		for _, binding := range bindings.Items {
			report.Expect(!(binding.RoleRef.Name == "cluster-admin" && bindsServiceAccount(binding.Subjects, "dev", hardening)),
				"Cluster role binding %s still gives service account dev the cluster-admin role.", binding.Name)
		}
	}
	if report.Found(viewErr, "RoleBinding", "dev-view", hardening) {
		report.Expect(view.RoleRef.Kind == "ClusterRole" && view.RoleRef.Name == "view", "Role binding dev-view should reference cluster role view.")
		report.Expect(bindsServiceAccount(view.Subjects, "dev", hardening), "Role binding dev-view should include service account dev.")
	}
	return report.Result("Question 34 - Service account dev in namespace hardening is bound to cluster-admin. Remove that access and give it the view cluster role in namespace hardening only, with role binding dev-view", "Hard")
}

func RemovePrivileges(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	deployment, err := clientset.AppsV1().Deployments(hardening).Get(ctx, "legacy-app", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "Deployment", "legacy-app", hardening) {
		spec := deployment.Spec.Template.Spec
		for _, container := range spec.Containers {
			privileged := container.SecurityContext != nil && utils.PtrEquals(container.SecurityContext.Privileged, true)
			report.Expect(!privileged, "Container %s still runs privileged.", container.Name)
		}
		report.Expect(!spec.HostPID, "The pods still share the host PID namespace.")
		report.Expect(!spec.HostNetwork, "The pods still share the host network namespace.")
	}
	return report.Result("Question 35 - Deployment legacy-app in namespace hardening runs privileged and shares the host PID and network namespaces. Make it run without any of them", "Hard")
}

func ImmutableSecret(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	secret, err := clientset.CoreV1().Secrets(hardening).Get(ctx, "api-keys", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "Secret", "api-keys", hardening) {
		token := string(secret.Data["token"])
		report.Expect(token == "s3cr3t", "Key token is %s, expected s3cr3t.", utils.Shown(token))
		report.Expect(utils.PtrEquals(secret.Immutable, true), "The secret is not immutable.")
	}
	return report.Result("Question 36 - Create secret api-keys in namespace hardening with data token=s3cr3t and make it immutable", "Easy")
}

func SandboxedRuntime(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	runtimeClass, classErr := clientset.NodeV1().RuntimeClasses().Get(ctx, "gvisor", metav1.GetOptions{})
	pod, podErr := clientset.CoreV1().Pods(hardening).Get(ctx, "sandboxed", metav1.GetOptions{})

	var report utils.Report
	if report.Found(classErr, "RuntimeClass", "gvisor", "") {
		report.Expect(runtimeClass.Handler == "runsc", "Runtime class handler is %s, expected runsc.", utils.Shown(runtimeClass.Handler))
	}
	if report.Found(podErr, "Pod", "sandboxed", hardening) {
		report.ExpectImage(pod.Spec, "nginx:alpine")
		report.Expect(utils.PtrEquals(pod.Spec.RuntimeClassName, "gvisor"), "The pod's runtime class is %s, expected gvisor.", utils.ShownPtr(pod.Spec.RuntimeClassName))
	}
	return report.Result("Question 37 - Create runtime class gvisor with handler runsc, and pod sandboxed in namespace hardening with image nginx:alpine that uses it", "Medium")
}
