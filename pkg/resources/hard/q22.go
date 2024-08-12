package hard

import (
	"context"
	"kubelearn/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CreateIngressYellow(clientset *kubernetes.Clientset) utils.Result {
	ingress, err := clientset.NetworkingV1().Ingresses("colors").Get(context.TODO(), "ingress-colors", metav1.GetOptions{})

	passed := err == nil && len(ingress.Spec.Rules) > 0 &&
		ingress.Spec.Rules[0].Host == "yellow.com" &&
		len(ingress.Spec.Rules[0].HTTP.Paths) > 0 &&
		ingress.Spec.Rules[0].HTTP.Paths[0].Path == "/yellow" &&
		ingress.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Name == "yellow-service"

	return utils.Result{
		TestName:   "Question 22 - Create an ingress ingress-colors with host yellow.com, path /yellow, and service yellow-service in namespace colors",
		Passed:     passed,
		Difficulty: "Hard",
	}
}
