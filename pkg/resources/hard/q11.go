package hard

import (
	"context"
	"kubelearn/pkg/utils"

	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

func CreateNetPolRule(ctx context.Context, clientset kubernetes.Interface) utils.Result {
	netPol, err := clientset.NetworkingV1().NetworkPolicies("colors").Get(ctx, "allow-policy-colors", metav1.GetOptions{})

	var report utils.Report
	if report.Found(err, "NetworkPolicy", "allow-policy-colors", "colors") {
		report.Expect(hasCorrectIngressRule(netPol.Spec.Ingress),
			"No ingress rule lets the redmobile-webserver pods reach the bluemobile-dbcache port.")
	}
	return report.Result("Question 11 - Create a network policy allow-policy-colors to allow redmobile-webserver to access bluemobile-dbcache", "Hard")
}

// hasCorrectIngressRule checks if the network policy has the correct ingress rule
func hasCorrectIngressRule(ingressRules []v1.NetworkPolicyIngressRule) bool {
	for _, rule := range ingressRules {
		for _, from := range rule.From {
			if from.PodSelector != nil {
				selector, err := metav1.LabelSelectorAsSelector(from.PodSelector)
				if err == nil && selector.Matches(labels.Set{"tier": "frontend"}) {
					for _, port := range rule.Ports {
						if port.Port != nil && port.Port.IntVal == 6379 {
							return true
						}
					}
				}
			}
		}
	}
	return false
}
