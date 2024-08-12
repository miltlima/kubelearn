package hard

import (
	"context"
	"kubelearn/pkg/utils"

	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

func CreateNetPolRule(clientset *kubernetes.Clientset) utils.Result {
	netPol, err := clientset.NetworkingV1().NetworkPolicies("colors").Get(context.TODO(), "allow-policy-colors", metav1.GetOptions{})

	passed := err == nil && hasCorrectIngressRule(netPol.Spec.Ingress)

	return utils.Result{
		TestName:   "Question 11 - Create a network policy allow-policy-colors to allow redmobile-webserver to access bluemobile-dbcache",
		Passed:     passed,
		Difficulty: "Hard",
	}
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
